# Filora Database Documentation

This folder is the authoritative reference for Filora's database design.

> **Source of truth for data**: PostgreSQL (Neon).
> **Source of truth for schema**: [`apps/api/internal/database/schema.sql`](../../apps/api/internal/database/schema.sql).
> These docs *describe* that file — if they ever disagree, the SQL wins and the docs must be fixed.

---

## How to use these docs

**For humans**
- Start here for orientation, then jump to the topic you need via the table of contents.
- Read [rules.md](./rules.md) before changing the schema.

**For AI agents**
- Treat [rules.md](./rules.md) as hard constraints. Do not introduce patterns it forbids.
- Before proposing schema changes, confirm the change against [rules.md](./rules.md)
  and reflect it in both `schema.sql` and the relevant doc file in the same change.
- When asked about relationships, use [erd.md](./erd.md).
- When asked about a specific column/table, use [schema.md](./schema.md).
- When asked about auth/permissions, use [rbac.md](./rbac.md).

---

## Table of contents

| Doc | What it covers |
|-----|----------------|
| [rules.md](./rules.md) | Design principles, ID strategy, naming, conventions, forbidden patterns, how to apply the schema |
| [erd.md](./erd.md) | Entity-relationship diagram and relationship descriptions |
| [schema.md](./schema.md) | Table-by-table column reference with types, constraints, and indexes |
| [rbac.md](./rbac.md) | Auth model: Clerk (web), CLI sessions (terminal), and RBAC with scoped permissions |

---

## At a glance

Filora is a private, family-scale multi-cloud Digital Asset Management app.

- **Web auth/sessions**: handled by [Clerk](https://clerk.com); we mirror the identity locally.
- **Terminal auth**: our own opaque, revocable CLI tokens — multiple concurrent sessions per user.
- **Authorization** (two tiers):
  - Global **RBAC** — `users → roles → permissions`, each grant scoped `own` or `all`
    (`superuser` holds the wildcard).
  - Per-resource **membership** — local `owner`/`editor`/`viewer` role on each gallery and album.
- **Organization**: every user has a default **gallery**; galleries hold **assets** and
  **albums**; assets are tagged via a normalized per-gallery **tag** vocabulary; owners
  can **invite** members by email; deleted assets go to a soft-delete **trash**.
- **Storage**: two layers — `serving` (Cloudinary/ImageKit free tier) and `archive`
  (GCS Archive / R2). Every asset is stored in **both** layers; archive replication runs
  async via **`archive_sync_jobs`**. Accounts are global, admin-managed; multiple accounts
  per layer spread capacity. Per-layer account election is a documented backlog.
- **Quota**: tracked **per gallery** (`galleries.storage_quota` / `storage_used`).
- **IDs**: incremental `BIGINT` for control tables; **UUID v7** for asset-scale/exposed rows.
- **Migrations**: none. `schema.sql` is applied manually.

### Tables

| Domain | Tables |
|--------|--------|
| Identity & access | `users`, `roles`, `permissions`, `role_permissions`, `user_roles`, `cli_sessions`, `clerk_webhook_events` |
| Organization | `galleries`, `gallery_members`, `albums`, `album_members`, `invitations`, `album_assets`, `tags`, `asset_tags` |
| Assets | `assets` (soft-deletable) |
| Storage | `storage_providers`, `storage_locations`, `archive_sync_jobs` (+ `storage_account_usage` view) |

| Table | ID | Purpose |
|-------|----|---------|
| `users` | bigint | Mirror of Clerk identities |
| `roles` / `permissions` | bigint | Global RBAC catalog |
| `role_permissions` / `user_roles` | composite | RBAC grants & assignments |
| `cli_sessions` | uuidv7 | Terminal sessions (hashed tokens, revocable) |
| `clerk_webhook_events` | bigint | Clerk webhook idempotency log |
| `galleries` | bigint | Top-level asset space (1 default per user, holds quota) |
| `gallery_members` | composite | Gallery sharing (`owner`/`editor`/`viewer`) |
| `albums` | bigint | Asset grouping within a gallery |
| `album_members` | composite | Album sharing |
| `invitations` | bigint | Invite by email to a gallery/album |
| `album_assets` | composite | Asset ↔ album (many-to-many) |
| `tags` | bigint | Per-gallery tag vocabulary |
| `asset_tags` | composite | Asset ↔ tag (many-to-many) |
| `assets` | uuidv7 | Logical asset records (metadata is truth; soft-delete) |
| `storage_providers` | bigint | Global storage accounts, one per layer |
| `storage_locations` | uuidv7 | Physical copies across layers (per-copy status) |
| `archive_sync_jobs` | bigint | Async replication to the archive layer (retry) |

---

## Applying the schema

```bash
cd apps/api
make db-apply   # psql -f internal/database/schema.sql
make db-seed    # psql -f internal/database/seed.sql
```

See [rules.md](./rules.md#applying-the-schema) for details and superuser bootstrap.
