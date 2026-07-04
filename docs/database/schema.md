# Schema Reference

Table-by-table reference. Authoritative DDL lives in
[`apps/api/internal/database/schema.sql`](../../apps/api/internal/database/schema.sql).
For relationships see [erd.md](./erd.md); for the auth model see [rbac.md](./rbac.md).

Sections: [Shared](#shared-building-blocks) · [Identity & access](#identity--access) ·
[Organization](#organization) · [Assets](#assets) · [Storage](#storage)

---

## Shared building blocks

### Function `uuid_generate_v7()`
Returns a time-ordered UUID v7. Portable implementation built on core
`gen_random_uuid()` (no extensions). Used as the default for UUID primary keys.
PostgreSQL gains a native `uuidv7()` only in PG 18; this covers Neon (PG ≤ 17).

### Function `set_updated_at()` + triggers
Sets `NEW.updated_at = now()` on `UPDATE`. Attached as `trg_<table>_updated_at`
to every table with an `updated_at` column (`users`, `roles`, `galleries`,
`albums`, `tags`, `storage_providers`, `assets`, `storage_locations`).

### Enums
| Enum | Values | Used by |
|------|--------|---------|
| `permission_scope` | `own`, `all` | `role_permissions.scope` |
| `member_role` | `owner`, `editor`, `viewer` | `gallery_members.role`, `album_members.role`, `invitations.role` |
| `storage_layer` | `serving`, `archive` | `storage_providers.layer`, `storage_locations.layer`, `archive_sync_jobs.target_layer` |
| `location_status` | `pending`, `stored`, `failed` | `storage_locations.status` |
| `invitation_status` | `pending`, `accepted`, `revoked`, `expired` | `invitations.status` |
| `job_status` | `pending`, `running`, `completed`, `failed` | `archive_sync_jobs.status` |

---

## Identity & access

### users
Local mirror of a Clerk identity. No passwords are ever stored.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `clerk_user_id` | `text` | NOT NULL, UNIQUE | Clerk id, e.g. `user_2ab...` |
| `email` | `text` | NOT NULL, UNIQUE | |
| `name` | `text` | NOT NULL | |
| `avatar_url` | `text` | | |
| `is_active` | `boolean` | NOT NULL, default `true` | |
| `last_seen_at` | `timestamptz` | | |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | trigger-maintained |

**Indexes:** `idx_users_email`; UNIQUE on `clerk_user_id`, `email`.
> Storage quota is tracked per gallery, not per user (see `galleries`).

### roles
Named RBAC roles; a user's set of roles is their "role group". Seeded rows in
[`seed.sql`](../../apps/api/internal/database/seed.sql).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `slug` | `text` | NOT NULL, UNIQUE | `superuser`, `admin`, `member`, `viewer` |
| `name` | `text` | NOT NULL | |
| `description` | `text` | | |
| `is_system` | `boolean` | NOT NULL, default `false` | seeded roles must not be deleted |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

### permissions
Catalog of `(resource, action)` pairs. Wildcards use `'*'`; `('*','*')` = full access.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `resource` | `text` | NOT NULL | `asset`, `gallery`, `album`, `tag`, `storage`, `user`, `role`, `session`, `dashboard`, `*` |
| `action` | `text` | NOT NULL | `read`, `create`, `update`, `delete`, `download`, `invite`, `assign`, `manage`, `revoke`, `*` |
| `description` | `text` | | |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**Constraints:** UNIQUE `(resource, action)`.

### role_permissions
Grant: which permission a role has, at what scope.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `role_id` | `bigint` | PK, FK → `roles(id)` CASCADE | |
| `permission_id` | `bigint` | PK, FK → `permissions(id)` CASCADE | |
| `scope` | `permission_scope` | NOT NULL, default `'own'` | `own` = owned rows only; `all` = whole workspace |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(role_id, permission_id)`. **Index:** `idx_role_permissions_permission_id`.

### user_roles
Assignment: which roles a user holds.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `user_id` | `bigint` | PK, FK → `users(id)` CASCADE | |
| `role_id` | `bigint` | PK, FK → `roles(id)` CASCADE | |
| `granted_by` | `bigint` | FK → `users(id)` SET NULL | who granted it |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(user_id, role_id)`. **Index:** `idx_user_roles_role_id`.

### cli_sessions
Terminal sessions. A user may hold many concurrently; each is revocable.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `uuid_generate_v7()` | |
| `user_id` | `bigint` | NOT NULL, FK → `users(id)` CASCADE | |
| `token_hash` | `text` | NOT NULL, UNIQUE | SHA-256 of raw token; raw token shown once |
| `label` | `text` | | device / terminal name |
| `ip_address` | `inet` | | |
| `user_agent` | `text` | | |
| `last_used_at` | `timestamptz` | | |
| `expires_at` | `timestamptz` | | NULL = never expires |
| `revoked_at` | `timestamptz` | | NULL = active |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_cli_sessions_user_id`; partial `idx_cli_sessions_active (user_id) WHERE revoked_at IS NULL`.

### clerk_webhook_events
Idempotency store for Clerk user-sync webhooks (Clerk may deliver an event more
than once or out of order).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `event_id` | `text` | NOT NULL, UNIQUE | Svix message / Clerk event id — the idempotency key |
| `event_type` | `text` | NOT NULL | e.g. `user.created`, `user.updated`, `user.deleted` |
| `payload` | `jsonb` | NOT NULL | raw event |
| `processed_at` | `timestamptz` | | NULL until handled |
| `error` | `text` | | last processing error |
| `received_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** partial `idx_clerk_webhook_events_unprocessed (received_at) WHERE processed_at IS NULL`.

---

## Organization

### galleries
Top-level asset space. Each user gets one **default** gallery; users can own
several and join others'.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `owner_id` | `bigint` | NOT NULL, FK → `users(id)` CASCADE | |
| `name` | `text` | NOT NULL | |
| `description` | `text` | | |
| `is_default` | `boolean` | NOT NULL, default `false` | auto-created gallery |
| `storage_quota` | `bigint` | NOT NULL, default 5 GB | per-gallery limit, bytes |
| `storage_used` | `bigint` | NOT NULL, default `0` | bytes |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_galleries_owner_id`; partial UNIQUE
`idx_galleries_one_default (owner_id) WHERE is_default` (one default per user).

> Quota is per gallery. Physical capacity is spread across multiple storage
> accounts per layer (`storage_providers`) to get around per-account free-tier limits.

### gallery_members
Who can access a gallery and at what local role. The owner also gets a row here
with role `owner`, so access checks are a single membership lookup.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `gallery_id` | `bigint` | PK, FK → `galleries(id)` CASCADE | |
| `user_id` | `bigint` | PK, FK → `users(id)` CASCADE | |
| `role` | `member_role` | NOT NULL, default `'viewer'` | owner/editor/viewer |
| `invited_by` | `bigint` | FK → `users(id)` SET NULL | |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(gallery_id, user_id)`. **Index:** `idx_gallery_members_user_id`.

### albums
Grouping of assets within a gallery. Owner can invite users.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `gallery_id` | `bigint` | NOT NULL, FK → `galleries(id)` CASCADE | album is nested in a gallery |
| `owner_id` | `bigint` | NOT NULL, FK → `users(id)` CASCADE | |
| `name` | `text` | NOT NULL | |
| `description` | `text` | | |
| `cover_asset_id` | `uuid` | FK → `assets(id)` SET NULL | optional cover |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_albums_gallery_id`, `idx_albums_owner_id`.

### album_members
Album sharing (owner invites users).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `album_id` | `bigint` | PK, FK → `albums(id)` CASCADE | |
| `user_id` | `bigint` | PK, FK → `users(id)` CASCADE | |
| `role` | `member_role` | NOT NULL, default `'viewer'` | |
| `invited_by` | `bigint` | FK → `users(id)` SET NULL | |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(album_id, user_id)`. **Index:** `idx_album_members_user_id`.

### invitations
Invite a user **by email** to a gallery or album. The invitee need not exist yet;
they get a link with `token` and, on acceptance (after Clerk sign-in), a
membership row is created.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `gallery_id` | `bigint` | FK → `galleries(id)` CASCADE | set for a gallery invite |
| `album_id` | `bigint` | FK → `albums(id)` CASCADE | set for an album invite |
| `email` | `text` | NOT NULL | invitee email |
| `role` | `member_role` | NOT NULL, default `'viewer'` | role granted on acceptance |
| `token` | `text` | NOT NULL, UNIQUE | opaque token for the invite link |
| `status` | `invitation_status` | NOT NULL, default `'pending'` | pending/accepted/revoked/expired |
| `invited_by` | `bigint` | FK → `users(id)` SET NULL | |
| `accepted_user_id` | `bigint` | FK → `users(id)` SET NULL | filled on acceptance |
| `expires_at` | `timestamptz` | | |
| `accepted_at` | `timestamptz` | | |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Constraints:** CHECK `invitations_one_target` — exactly one of `gallery_id` /
`album_id` is set. **Indexes:** `idx_invitations_email`; partial UNIQUE per
target+email while `status = 'pending'` (`idx_invitations_pending_gallery`,
`idx_invitations_pending_album`).

### album_assets
Many-to-many between albums and assets.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `album_id` | `bigint` | PK, FK → `albums(id)` CASCADE | |
| `asset_id` | `uuid` | PK, FK → `assets(id)` CASCADE | |
| `added_by` | `bigint` | FK → `users(id)` SET NULL | |
| `sort_order` | `integer` | NOT NULL, default `0` | manual ordering |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(album_id, asset_id)`. **Index:** `idx_album_assets_asset_id`.

### tags
Per-gallery tag vocabulary.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `gallery_id` | `bigint` | NOT NULL, FK → `galleries(id)` CASCADE | |
| `name` | `text` | NOT NULL | |
| `created_by` | `bigint` | FK → `users(id)` SET NULL | |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Constraints:** UNIQUE `(gallery_id, name)`.

### asset_tags
Many-to-many between assets and tags.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `asset_id` | `uuid` | PK, FK → `assets(id)` CASCADE | |
| `tag_id` | `bigint` | PK, FK → `tags(id)` CASCADE | |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**PK:** `(asset_id, tag_id)`. **Index:** `idx_asset_tags_tag_id`.

---

## Assets

### assets
Logical asset record; metadata here is the source of truth. Lives in exactly one
gallery. Tagging is normalized (see `asset_tags`), so there is no `tags` column.
Soft-deletable (trash) via `deleted_at`.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `uuid_generate_v7()` | |
| `gallery_id` | `bigint` | NOT NULL, FK → `galleries(id)` CASCADE | container |
| `uploaded_by` | `bigint` | FK → `users(id)` SET NULL | contributor; used for RBAC `own` scope |
| `name` | `text` | NOT NULL | |
| `type` | `text` | NOT NULL, CHECK in (`image`,`video`,`document`,`archive`,`file`) | |
| `mime_type` | `text` | NOT NULL | |
| `size` | `bigint` | NOT NULL | bytes |
| `hash` | `text` | NOT NULL | SHA-256, for dedup |
| `metadata` | `jsonb` | | e.g. width/height |
| `deleted_at` | `timestamptz` | | NULL = active; set = in trash |
| `deleted_by` | `bigint` | FK → `users(id)` SET NULL | |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_assets_gallery_id`, `idx_assets_uploaded_by`, `idx_assets_type`,
`idx_assets_created_at (created_at DESC)`; partial `idx_assets_gallery_active (gallery_id) WHERE deleted_at IS NULL`;
partial UNIQUE `idx_assets_gallery_hash (gallery_id, hash) WHERE deleted_at IS NULL`
— dedup is per gallery and ignores trashed rows (so re-upload after deletion is allowed).

---

## Storage

Two layers, both required for every asset:

- **serving** — hot, publicly-servable free-tier accounts (Cloudinary / ImageKit).
- **archive** — cold, cheap archive-class storage (GCS Archive / R2 / etc.).

Storage accounts are **global** and managed by superadmin / users holding the
`storage` permission — they are not owned by end users.

> **Backlog — account election.** How a new upload picks which account within a
> layer (round-robin, most-free-space, weighted, etc.) is not modeled yet. It
> will be a selection strategy over active accounts, added when needed.

### storage_providers
A global storage account, bound to one layer.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | |
| `layer` | `storage_layer` | NOT NULL | `serving` or `archive` |
| `name` | `text` | NOT NULL | |
| `type` | `text` | NOT NULL, CHECK in (`cloudinary`,`imagekit`,`r2`,`gcs`) | |
| `credentials` | `jsonb` | NOT NULL | provider secrets |
| `quota` | `bigint` | | NULL = no fixed quota |
| `used` | `bigint` | NOT NULL, default `0` | bytes |
| `is_active` | `boolean` | NOT NULL, default `true` | |
| `created_by` | `bigint` | FK → `users(id)` SET NULL | audit only |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_storage_providers_layer`, `idx_storage_providers_is_active`.

### storage_locations
A physical copy of an asset in one provider account.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `uuid_generate_v7()` | |
| `asset_id` | `uuid` | NOT NULL, FK → `assets(id)` CASCADE | |
| `provider_id` | `bigint` | NOT NULL, FK → `storage_providers(id)` **RESTRICT** | can't delete an account still hosting files |
| `layer` | `storage_layer` | NOT NULL | mirrors the provider's layer |
| `provider_key` | `text` | NOT NULL | key/path within the provider |
| `url` | `text` | | public URL (may be NULL for archive) |
| `status` | `location_status` | NOT NULL, default `'pending'` | `pending` → `stored` / `failed` |
| `metadata` | `jsonb` | | |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_storage_locations_asset_id`, `idx_storage_locations_provider_id`,
`idx_storage_locations_asset_layer (asset_id, layer)`, `idx_storage_locations_status`;
UNIQUE `idx_storage_locations_asset_provider (asset_id, provider_id)`.

### archive_sync_jobs
Background replication of an asset into a target layer (typically `archive`).
Uploads hit the serving layer synchronously; archive replication runs async with
retry. On success the worker creates/updates the `storage_location` and marks the
job `completed`.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `bigint` | PK, identity | queue order |
| `asset_id` | `uuid` | NOT NULL, FK → `assets(id)` CASCADE | |
| `target_layer` | `storage_layer` | NOT NULL, default `'archive'` | |
| `provider_id` | `bigint` | FK → `storage_providers(id)` SET NULL | chosen account (backlog: election) |
| `status` | `job_status` | NOT NULL, default `'pending'` | pending/running/completed/failed |
| `attempts` | `integer` | NOT NULL, default `0` | |
| `max_attempts` | `integer` | NOT NULL, default `5` | |
| `last_error` | `text` | | |
| `next_retry_at` | `timestamptz` | | when the job becomes due again |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL, default `now()` | |

**Indexes:** `idx_archive_sync_jobs_asset_id`; partial
`idx_archive_sync_jobs_queue (next_retry_at) WHERE status IN ('pending','failed')`;
partial UNIQUE `idx_archive_sync_jobs_open (asset_id, target_layer) WHERE status IN ('pending','running')`
— at most one open job per asset+layer.

### storage_account_usage (view)
Convenience read model for the storage-management UI. One row per account:

| Column | Notes |
|--------|-------|
| `id`, `name`, `layer`, `type`, `is_active`, `quota`, `used` | from `storage_providers` |
| `used_percent` | `used / quota * 100` (NULL when quota is NULL/0) |
| `location_count` | total copies hosted on the account |
| `stored_count` / `pending_count` / `failed_count` | copies by `status` |
