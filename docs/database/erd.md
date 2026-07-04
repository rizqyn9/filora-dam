# Entity-Relationship Diagram

Visual model of Filora's schema. Mirrors [`schema.sql`](../../apps/api/internal/database/schema.sql);
for exact column types and constraints see [schema.md](./schema.md).

> Rendering: the diagram below uses [Mermaid](https://mermaid.js.org/). GitHub, VS Code,
> and most Markdown viewers render it natively.

---

## Diagram

```mermaid
erDiagram
    users ||--o{ user_roles          : "has"
    roles ||--o{ user_roles          : "assigned via"
    roles ||--o{ role_permissions    : "grants"
    permissions ||--o{ role_permissions : "granted in"
    users ||--o{ cli_sessions        : "owns"

    users ||--o{ galleries           : "owns"
    galleries ||--o{ gallery_members : "shared via"
    users ||--o{ gallery_members     : "member of"

    galleries ||--o{ assets          : "contains"
    users ||--o{ assets              : "uploaded"

    galleries ||--o{ albums          : "contains"
    users ||--o{ albums              : "owns"
    albums ||--o{ album_members      : "shared via"
    users ||--o{ album_members       : "member of"
    albums ||--o{ album_assets       : "groups"
    assets ||--o{ album_assets       : "in"

    galleries ||--o{ invitations     : "invite to"
    albums ||--o{ invitations        : "invite to"
    users ||--o{ invitations         : "sent"

    galleries ||--o{ tags            : "scopes"
    tags ||--o{ asset_tags           : "applied via"
    assets ||--o{ asset_tags         : "tagged"

    assets ||--o{ storage_locations  : "stored at"
    storage_providers ||--o{ storage_locations : "hosts"
    assets ||--o{ archive_sync_jobs  : "replicated by"
    storage_providers ||--o{ archive_sync_jobs : "targets"

    users {
        bigint      id PK
        text        clerk_user_id UK
        text        email UK
        text        name
        boolean     is_active
    }
    roles {
        bigint id PK
        text   slug UK
        boolean is_system
    }
    permissions {
        bigint id PK
        text   resource
        text   action
    }
    role_permissions {
        bigint           role_id PK,FK
        bigint           permission_id PK,FK
        permission_scope scope
    }
    user_roles {
        bigint user_id PK,FK
        bigint role_id PK,FK
        bigint granted_by FK
    }
    cli_sessions {
        uuid   id PK
        bigint user_id FK
        text   token_hash UK
        timestamptz revoked_at
    }
    galleries {
        bigint  id PK
        bigint  owner_id FK
        text    name
        boolean is_default
        bigint  storage_quota
        bigint  storage_used
    }
    gallery_members {
        bigint      gallery_id PK,FK
        bigint      user_id PK,FK
        member_role role
        bigint      invited_by FK
    }
    albums {
        bigint id PK
        bigint gallery_id FK
        bigint owner_id FK
        text   name
        uuid   cover_asset_id FK
    }
    album_members {
        bigint      album_id PK,FK
        bigint      user_id PK,FK
        member_role role
        bigint      invited_by FK
    }
    album_assets {
        bigint  album_id PK,FK
        uuid    asset_id PK,FK
        bigint  added_by FK
        integer sort_order
    }
    tags {
        bigint id PK
        bigint gallery_id FK
        text   name
    }
    asset_tags {
        uuid   asset_id PK,FK
        bigint tag_id PK,FK
    }
    assets {
        uuid   id PK
        bigint gallery_id FK
        bigint uploaded_by FK
        text   name
        text   type
        text   mime_type
        bigint size
        text   hash
        jsonb  metadata
        timestamptz deleted_at
    }
    invitations {
        bigint            id PK
        bigint            gallery_id FK
        bigint            album_id FK
        text              email
        member_role       role
        text              token UK
        invitation_status status
        bigint            invited_by FK
    }
    clerk_webhook_events {
        bigint      id PK
        text        event_id UK
        text        event_type
        jsonb       payload
        timestamptz processed_at
    }
    archive_sync_jobs {
        bigint        id PK
        uuid          asset_id FK
        storage_layer target_layer
        bigint        provider_id FK
        job_status    status
        integer       attempts
        timestamptz   next_retry_at
    }
    storage_providers {
        bigint        id PK
        storage_layer layer
        text          name
        text          type
        jsonb         credentials
        bigint        quota
        bigint        used
        boolean       is_active
        bigint        created_by FK
    }
    storage_locations {
        uuid            id PK
        uuid            asset_id FK
        bigint          provider_id FK
        storage_layer   layer
        text            provider_key
        text            url
        location_status status
    }
```

> `storage_account_usage` is a **view** (not a table) summarizing usage per
> storage account; it is omitted from the ERD.

---

## Relationships

| From | To | Cardinality | On delete | Notes |
|------|----|-------------|-----------|-------|
| `user_roles.user_id` | `users.id` | many-to-one | CASCADE | A user's set of roles = their "role group" |
| `user_roles.role_id` | `roles.id` | many-to-one | CASCADE | |
| `user_roles.granted_by` | `users.id` | many-to-one | SET NULL | Who assigned the role |
| `role_permissions.role_id` | `roles.id` | many-to-one | CASCADE | |
| `role_permissions.permission_id` | `permissions.id` | many-to-one | CASCADE | Grant carries a `scope` (`own`/`all`) |
| `cli_sessions.user_id` | `users.id` | many-to-one | CASCADE | Many concurrent terminal sessions per user |
| `galleries.owner_id` | `users.id` | many-to-one | CASCADE | Each user has 1 default gallery |
| `gallery_members.gallery_id` | `galleries.id` | many-to-one | CASCADE | Local role: owner/editor/viewer |
| `gallery_members.user_id` | `users.id` | many-to-one | CASCADE | Owner also has a member row |
| `gallery_members.invited_by` | `users.id` | many-to-one | SET NULL | |
| `assets.gallery_id` | `galleries.id` | many-to-one | CASCADE | Asset lives in exactly one gallery |
| `assets.uploaded_by` | `users.id` | many-to-one | SET NULL | Contributor; used for `own` scope; dedup unique `(gallery_id, hash)` |
| `albums.gallery_id` | `galleries.id` | many-to-one | CASCADE | Album nested in a gallery |
| `albums.owner_id` | `users.id` | many-to-one | CASCADE | |
| `albums.cover_asset_id` | `assets.id` | many-to-one | SET NULL | Optional cover image |
| `album_members.album_id` | `albums.id` | many-to-one | CASCADE | Owner can invite users |
| `album_members.user_id` | `users.id` | many-to-one | CASCADE | |
| `invitations.gallery_id` | `galleries.id` | many-to-one | CASCADE | Set for a gallery invite (CHECK: exactly one target) |
| `invitations.album_id` | `albums.id` | many-to-one | CASCADE | Set for an album invite |
| `invitations.invited_by` / `accepted_user_id` | `users.id` | many-to-one | SET NULL | Sender / accepted user |
| `album_assets.album_id` | `albums.id` | many-to-one | CASCADE | M2M: asset ↔ album |
| `album_assets.asset_id` | `assets.id` | many-to-one | CASCADE | |
| `tags.gallery_id` | `galleries.id` | many-to-one | CASCADE | Tag vocabulary is per gallery; unique `(gallery_id, name)` |
| `asset_tags.asset_id` | `assets.id` | many-to-one | CASCADE | M2M: asset ↔ tag |
| `asset_tags.tag_id` | `tags.id` | many-to-one | CASCADE | |
| `assets.deleted_by` | `users.id` | many-to-one | SET NULL | Soft-delete (trash) actor |
| `storage_locations.asset_id` | `assets.id` | many-to-one | CASCADE | Each asset gets ≥1 serving + ≥1 archive copy |
| `storage_locations.provider_id` | `storage_providers.id` | many-to-one | **RESTRICT** | Can't delete an account still hosting files; unique `(asset_id, provider_id)` |
| `archive_sync_jobs.asset_id` | `assets.id` | many-to-one | CASCADE | Async replication into a layer |
| `archive_sync_jobs.provider_id` | `storage_providers.id` | many-to-one | SET NULL | Chosen account (election = backlog) |
| `storage_providers.created_by` | `users.id` | many-to-one | SET NULL | Audit only; accounts are global |

> `clerk_webhook_events` has no foreign keys (idempotency log for Clerk webhooks).

---

## Logical groupings

- **Identity & access**: `users`, `roles`, `permissions`, `role_permissions`, `user_roles`, `cli_sessions`, `clerk_webhook_events`
  (Clerk mirror + RBAC + terminal sessions — see [rbac.md](./rbac.md)).
- **Organization**: `galleries`, `gallery_members`, `albums`, `album_members`, `invitations`, `album_assets`, `tags`, `asset_tags`.
- **Assets & storage**: `assets`, `storage_locations`, `storage_providers`, `archive_sync_jobs`
  (metadata is truth; two storage layers — see [schema.md](./schema.md#storage)).
