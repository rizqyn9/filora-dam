# Schema Reference

Per-table column definitions, constraints, indexes, and domain context for every table in Filora's database.

---

## Identity & Access

### users

Mirrors Clerk identities. Created via webhook on first sign-in.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `clerk_user_id` | TEXT | No | — | Unique. Clerk's external ID (`user_2abc...`) |
| `email` | TEXT | No | — | Unique |
| `name` | TEXT | No | — | Display name |
| `avatar_url` | TEXT | Yes | — | Profile image URL from Clerk |
| `is_active` | BOOLEAN | No | `TRUE` | Soft-disable without deletion |
| `last_seen_at` | TIMESTAMPTZ | Yes | — | Updated on activity |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:** `idx_users_email` (email)

---

### roles

Global RBAC roles. System roles (`is_system = TRUE`) cannot be deleted.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `slug` | TEXT | No | — | Unique. Machine name: `superuser`, `admin`, `member`, `viewer` |
| `name` | TEXT | No | — | Human-readable label |
| `description` | TEXT | Yes | — | |
| `is_system` | BOOLEAN | No | `FALSE` | Protected from deletion |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

---

### permissions

Catalog of resource + action pairs. Wildcard `('*', '*')` = full access.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `resource` | TEXT | No | — | `asset`, `gallery`, `album`, `tag`, `storage`, `user`, `role`, `session`, `dashboard`, `*` |
| `action` | TEXT | No | — | `read`, `create`, `update`, `delete`, `download`, `invite`, `assign`, `manage`, `revoke`, `*` |
| `description` | TEXT | Yes | — | Human explanation |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Constraints:** `UNIQUE (resource, action)`

---

### role_permissions

Grants: which permissions a role holds, at which scope.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `role_id` | BIGINT | No | — | PK (composite). FK → `roles` CASCADE |
| `permission_id` | BIGINT | No | — | PK (composite). FK → `permissions` CASCADE |
| `scope` | `permission_scope` | No | `'own'` | `own` = user's resources only; `all` = everything |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_role_permissions_permission_id` (permission_id)

---

### user_roles

Assignment of global roles to users.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `user_id` | BIGINT | No | — | PK (composite). FK → `users` CASCADE |
| `role_id` | BIGINT | No | — | PK (composite). FK → `roles` CASCADE |
| `granted_by` | BIGINT | Yes | — | FK → `users` SET NULL. Who assigned this role |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_user_roles_role_id` (role_id)

---

### cli_sessions

Opaque token-based sessions for terminal/CLI login. Multiple concurrent per user.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | UUID | No | `uuid_generate_v7()` | PK |
| `user_id` | BIGINT | No | — | FK → `users` CASCADE |
| `token_hash` | TEXT | No | — | Unique. SHA-256 of the raw token (raw shown once) |
| `label` | TEXT | Yes | — | Device/terminal name |
| `ip_address` | INET | Yes | — | Connection IP |
| `user_agent` | TEXT | Yes | — | Client identifier |
| `last_used_at` | TIMESTAMPTZ | Yes | — | Refreshed on use |
| `expires_at` | TIMESTAMPTZ | Yes | — | NULL = never expires |
| `revoked_at` | TIMESTAMPTZ | Yes | — | NULL = active session |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:**
- `idx_cli_sessions_user_id` (user_id)
- `idx_cli_sessions_active` (user_id) WHERE `revoked_at IS NULL`

---

### clerk_webhook_events

Idempotency log for Clerk webhook deliveries. Prevents duplicate processing.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `event_id` | TEXT | No | — | Unique. Svix message ID (idempotency key) |
| `event_type` | TEXT | No | — | e.g. `user.created`, `user.updated`, `user.deleted` |
| `payload` | JSONB | No | — | Full Clerk webhook body |
| `processed_at` | TIMESTAMPTZ | Yes | — | NULL until successfully handled |
| `error` | TEXT | Yes | — | Last processing error |
| `received_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_clerk_webhook_events_unprocessed` (received_at) WHERE `processed_at IS NULL`

---

## Organization

### galleries

Top-level asset container. Each user gets one auto-created default gallery.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `owner_id` | BIGINT | No | — | FK → `users` CASCADE |
| `name` | TEXT | No | — | |
| `description` | TEXT | Yes | — | |
| `is_default` | BOOLEAN | No | `FALSE` | Auto-created gallery for a user |
| `storage_quota` | BIGINT | No | `5368709120` | 5 GB in bytes |
| `storage_used` | BIGINT | No | `0` | Current usage in bytes |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_galleries_owner_id` (owner_id)
- `idx_galleries_one_default` UNIQUE (owner_id) WHERE `is_default` — enforces max one default per user

---

### gallery_members

Membership/access control for galleries. Owner also gets a row with role `'owner'`.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `gallery_id` | BIGINT | No | — | PK (composite). FK → `galleries` CASCADE |
| `user_id` | BIGINT | No | — | PK (composite). FK → `users` CASCADE |
| `role` | `member_role` | No | `'viewer'` | `owner`, `editor`, `viewer` |
| `invited_by` | BIGINT | Yes | — | FK → `users` SET NULL |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_gallery_members_user_id` (user_id)

---

### albums

Grouping of assets within a gallery. Owned by a user, scoped to a gallery.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `gallery_id` | BIGINT | No | — | FK → `galleries` CASCADE |
| `owner_id` | BIGINT | No | — | FK → `users` CASCADE |
| `name` | TEXT | No | — | |
| `description` | TEXT | Yes | — | |
| `cover_asset_id` | UUID | Yes | — | FK → `assets` SET NULL. Display cover |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_albums_gallery_id` (gallery_id)
- `idx_albums_owner_id` (owner_id)

---

### album_members

Per-album access control via local membership role.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `album_id` | BIGINT | No | — | PK (composite). FK → `albums` CASCADE |
| `user_id` | BIGINT | No | — | PK (composite). FK → `users` CASCADE |
| `role` | `member_role` | No | `'viewer'` | `owner`, `editor`, `viewer` |
| `invited_by` | BIGINT | Yes | — | FK → `users` SET NULL |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_album_members_user_id` (user_id)

---

### invitations

Invite a user (by email) to a gallery or album. Targets exactly one via CHECK constraint.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `gallery_id` | BIGINT | Yes | — | FK → `galleries` CASCADE. XOR with album_id |
| `album_id` | BIGINT | Yes | — | FK → `albums` CASCADE. XOR with gallery_id |
| `email` | TEXT | No | — | Invitee email (may not be a user yet) |
| `role` | `member_role` | No | `'viewer'` | Role granted on acceptance |
| `token` | TEXT | No | — | Unique. Opaque token for invite link |
| `status` | `invitation_status` | No | `'pending'` | `pending`, `accepted`, `revoked`, `expired` |
| `invited_by` | BIGINT | Yes | — | FK → `users` SET NULL |
| `accepted_user_id` | BIGINT | Yes | — | FK → `users` SET NULL. Set on acceptance |
| `expires_at` | TIMESTAMPTZ | Yes | — | NULL = no expiry |
| `accepted_at` | TIMESTAMPTZ | Yes | — | |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Constraints:** `CHECK ((gallery_id IS NOT NULL)::int + (album_id IS NOT NULL)::int = 1)`

**Indexes:**
- `idx_invitations_email` (email)
- `idx_invitations_pending_gallery` UNIQUE (gallery_id, email) WHERE `status = 'pending' AND gallery_id IS NOT NULL`
- `idx_invitations_pending_album` UNIQUE (album_id, email) WHERE `status = 'pending' AND album_id IS NOT NULL`

---

### album_assets

Join table: many-to-many between albums and assets. Supports ordering.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `album_id` | BIGINT | No | — | PK (composite). FK → `albums` CASCADE |
| `asset_id` | UUID | No | — | PK (composite). FK → `assets` CASCADE |
| `added_by` | BIGINT | Yes | — | FK → `users` SET NULL |
| `sort_order` | INTEGER | No | `0` | Display order within album |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_album_assets_asset_id` (asset_id)

---

### tags

Per-gallery tag vocabulary. Unique name within a gallery.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `gallery_id` | BIGINT | No | — | FK → `galleries` CASCADE |
| `name` | TEXT | No | — | |
| `created_by` | BIGINT | Yes | — | FK → `users` SET NULL |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Constraints:** `UNIQUE (gallery_id, name)`

---

### asset_tags

Join table: many-to-many between assets and tags.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `asset_id` | UUID | No | — | PK (composite). FK → `assets` CASCADE |
| `tag_id` | BIGINT | No | — | PK (composite). FK → `tags` CASCADE |
| `created_at` | TIMESTAMPTZ | No | `now()` | |

**Indexes:** `idx_asset_tags_tag_id` (tag_id)

---

## Assets

### assets

Core entity: a logical file (image, video, document). Metadata is the source of truth, not storage.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | UUID | No | `uuid_generate_v7()` | PK |
| `gallery_id` | BIGINT | No | — | FK → `galleries` CASCADE |
| `uploaded_by` | BIGINT | Yes | — | FK → `users` SET NULL. Used for RBAC 'own' scope |
| `name` | TEXT | No | — | Original filename or user-given name |
| `type` | TEXT | No | — | CHECK: `image`, `video`, `document`, `archive`, `file` |
| `mime_type` | TEXT | No | — | e.g. `image/jpeg`, `application/pdf` |
| `size` | BIGINT | No | — | File size in bytes |
| `hash` | TEXT | No | — | SHA-256 of file content. Dedup key |
| `metadata` | JSONB | Yes | — | Dimensions, duration, EXIF — varies by type |
| `deleted_at` | TIMESTAMPTZ | Yes | — | NULL = active; set = soft-deleted (trash) |
| `deleted_by` | BIGINT | Yes | — | FK → `users` SET NULL. Who trashed it |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_assets_gallery_id` (gallery_id)
- `idx_assets_uploaded_by` (uploaded_by)
- `idx_assets_type` (type)
- `idx_assets_created_at` (created_at DESC)
- `idx_assets_gallery_active` (gallery_id) WHERE `deleted_at IS NULL`
- `idx_assets_gallery_hash` UNIQUE (gallery_id, hash) WHERE `deleted_at IS NULL` — dedup per gallery

**Deduplication:** Same file (by SHA-256) cannot exist twice in the same gallery while active. Trashing an asset frees the hash for re-upload.

---

## Storage

### storage_providers

Global storage accounts managed by admins. Not user-owned.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `layer` | `storage_layer` | No | — | `serving` (hot) or `archive` (cold) |
| `name` | TEXT | No | — | Human label |
| `type` | TEXT | No | — | CHECK: `cloudinary`, `imagekit`, `r2`, `gcs` |
| `credentials` | JSONB | No | — | Provider-specific auth (encrypted at rest) |
| `quota` | BIGINT | Yes | — | NULL = no fixed quota |
| `used` | BIGINT | No | `0` | Current usage in bytes |
| `is_active` | BOOLEAN | No | `TRUE` | Deactivate instead of deleting |
| `created_by` | BIGINT | Yes | — | FK → `users` SET NULL. Audit |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_storage_providers_layer` (layer)
- `idx_storage_providers_is_active` (is_active)

---

### storage_locations

Physical copies of an asset across providers/layers. Each asset targets >= 1 serving + >= 1 archive location.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | UUID | No | `uuid_generate_v7()` | PK |
| `asset_id` | UUID | No | — | FK → `assets` CASCADE |
| `provider_id` | BIGINT | No | — | FK → `storage_providers` RESTRICT |
| `layer` | `storage_layer` | No | — | Mirrors the provider's layer |
| `provider_key` | TEXT | No | — | Path/key within the provider |
| `url` | TEXT | Yes | — | Public URL (may be NULL for archive) |
| `status` | `location_status` | No | `'pending'` | `pending`, `stored`, `failed` |
| `metadata` | JSONB | Yes | — | Provider response data |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_storage_locations_asset_id` (asset_id)
- `idx_storage_locations_provider_id` (provider_id)
- `idx_storage_locations_asset_layer` (asset_id, layer)
- `idx_storage_locations_status` (status)
- `idx_storage_locations_asset_provider` UNIQUE (asset_id, provider_id) — one copy per provider

---

### archive_sync_jobs

Background replication jobs. Drives async copy from serving to archive layer.

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| `id` | BIGINT | No | IDENTITY | PK |
| `asset_id` | UUID | No | — | FK → `assets` CASCADE |
| `target_layer` | `storage_layer` | No | `'archive'` | Destination layer |
| `provider_id` | BIGINT | Yes | — | FK → `storage_providers` SET NULL. Filled on account election |
| `status` | `job_status` | No | `'pending'` | `pending`, `running`, `completed`, `failed` |
| `attempts` | INTEGER | No | `0` | Current retry count |
| `max_attempts` | INTEGER | No | `5` | Give up after this many |
| `last_error` | TEXT | Yes | — | Most recent error message |
| `next_retry_at` | TIMESTAMPTZ | Yes | — | When to retry next |
| `created_at` | TIMESTAMPTZ | No | `now()` | |
| `updated_at` | TIMESTAMPTZ | No | `now()` | Trigger-maintained |

**Indexes:**
- `idx_archive_sync_jobs_asset_id` (asset_id)
- `idx_archive_sync_jobs_queue` (next_retry_at) WHERE `status IN ('pending', 'failed')` — worker polling
- `idx_archive_sync_jobs_open` UNIQUE (asset_id, target_layer) WHERE `status IN ('pending', 'running')` — one active job per asset+layer

---

## Views

### storage_account_usage

Read-only summary for the storage management UI.

| Column | Source | Description |
|--------|--------|-------------|
| `id` | `storage_providers.id` | Provider ID |
| `name` | `storage_providers.name` | Provider name |
| `layer` | `storage_providers.layer` | serving or archive |
| `type` | `storage_providers.type` | cloudinary, imagekit, r2, gcs |
| `is_active` | `storage_providers.is_active` | Active flag |
| `quota` | `storage_providers.quota` | Total quota in bytes |
| `used` | `storage_providers.used` | Current usage in bytes |
| `used_percent` | Computed | `round((used / quota) * 100, 2)`. NULL if no quota |
| `location_count` | COUNT | Total storage_locations for this provider |
| `stored_count` | COUNT filtered | Locations with status = `'stored'` |
| `pending_count` | COUNT filtered | Locations with status = `'pending'` |
| `failed_count` | COUNT filtered | Locations with status = `'failed'` |

---

## Helper Functions

| Function | Returns | Purpose |
|----------|---------|---------|
| `uuid_generate_v7()` | `uuid` | Time-ordered UUID v7 for high-volume PKs |
| `set_updated_at()` | `trigger` | Sets `NEW.updated_at = now()` on UPDATE |

---

## Enum Reference

```sql
CREATE TYPE permission_scope   AS ENUM ('own', 'all');
CREATE TYPE member_role        AS ENUM ('owner', 'editor', 'viewer');
CREATE TYPE storage_layer      AS ENUM ('serving', 'archive');
CREATE TYPE location_status    AS ENUM ('pending', 'stored', 'failed');
CREATE TYPE invitation_status  AS ENUM ('pending', 'accepted', 'revoked', 'expired');
CREATE TYPE job_status         AS ENUM ('pending', 'running', 'completed', 'failed');
```

---

**Next:** [ERD](./erd.md) — Visual diagram of all relationships.
