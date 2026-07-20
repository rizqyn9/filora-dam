# Entity-Relationship Diagram

ASCII diagram of all Filora tables, grouped by domain, showing primary keys, foreign keys, and cardinality.

---

## Reading Guide

```
[table_name]        = table
PK                  = primary key
FK                  = foreign key
──>                 = many-to-one (FK points to the "one" side)
──<                 = one-to-many
>──<                = many-to-many (via join table)
(UUID)              = UUID v7 primary key
(BIGINT)            = BIGINT identity primary key
```

---

## Full Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        IDENTITY & ACCESS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────────┐    │
│  │    users     │         │    roles     │         │   permissions    │    │
│  │  (BIGINT)    │         │  (BIGINT)    │         │   (BIGINT)       │    │
│  │──────────────│         │──────────────│         │──────────────────│    │
│  │ clerk_user_id│         │ slug         │         │ resource         │    │
│  │ email        │         │ name         │         │ action           │    │
│  │ name         │         │ is_system    │         │ description      │    │
│  │ avatar_url   │         └──────┬───────┘         └────────┬─────────┘    │
│  │ is_active    │                │                           │              │
│  └──────┬───────┘                │                           │              │
│         │                        │                           │              │
│         │    ┌───────────────────┴───────────────────────────┘              │
│         │    │                                                              │
│         │    ▼                                                              │
│         │  ┌──────────────────────┐                                        │
│         │  │  role_permissions    │                                        │
│         │  │  (composite PK)      │                                        │
│         │  │──────────────────────│                                        │
│         │  │ role_id      FK ──> roles                                     │
│         │  │ permission_id FK ──> permissions                              │
│         │  │ scope (own|all)                                               │
│         │  └──────────────────────┘                                        │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────┐                                                      │
│  │   user_roles     │                                                      │
│  │  (composite PK)   │                                                      │
│  │──────────────────│                                                      │
│  │ user_id  FK ──> users                                                   │
│  │ role_id  FK ──> roles                                                   │
│  │ granted_by FK ──> users (nullable)                                      │
│  └──────────────────┘                                                      │
│                                                                             │
│  ┌──────────────────────┐      ┌──────────────────────────┐               │
│  │    cli_sessions      │      │  clerk_webhook_events    │               │
│  │      (UUID)          │      │       (BIGINT)           │               │
│  │──────────────────────│      │──────────────────────────│               │
│  │ user_id FK ──> users │      │ event_id (unique)        │               │
│  │ token_hash (unique)  │      │ event_type               │               │
│  │ label                │      │ payload (JSONB)          │               │
│  │ expires_at           │      │ processed_at             │               │
│  │ revoked_at           │      └──────────────────────────┘               │
│  └──────────────────────┘                                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                           ORGANIZATION                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────┐                                                     │
│  │    galleries      │                                                     │
│  │    (BIGINT)       │                                                     │
│  │───────────────────│                                                     │
│  │ owner_id FK ──> users                                                   │
│  │ name              │                                                     │
│  │ is_default        │                                                     │
│  │ storage_quota     │                                                     │
│  │ storage_used      │                                                     │
│  └─────────┬─────────┘                                                     │
│            │                                                                │
│            ├────────────────────────────────────┐                           │
│            │                                    │                           │
│            ▼                                    ▼                           │
│  ┌────────────────────┐              ┌─────────────────────┐               │
│  │  gallery_members   │              │      albums         │               │
│  │  (composite PK)    │              │     (BIGINT)        │               │
│  │────────────────────│              │─────────────────────│               │
│  │ gallery_id FK ──> galleries       │ gallery_id FK ──> galleries         │
│  │ user_id    FK ──> users           │ owner_id   FK ──> users             │
│  │ role (owner|editor|viewer)        │ name                │               │
│  │ invited_by FK ──> users           │ cover_asset_id FK ──> assets        │
│  └────────────────────┘              └─────────┬───────────┘               │
│                                                │                           │
│                                     ┌──────────┼──────────┐                │
│                                     │          │          │                │
│                                     ▼          ▼          ▼                │
│                          ┌────────────────┐ ┌──────────────────┐           │
│                          │ album_members  │ │  album_assets    │           │
│                          │ (composite PK) │ │  (composite PK)  │           │
│                          │────────────────│ │──────────────────│           │
│                          │ album_id FK    │ │ album_id FK ──> albums       │
│                          │ user_id  FK    │ │ asset_id FK ──> assets       │
│                          │ role           │ │ added_by FK ──> users        │
│                          │ invited_by FK  │ │ sort_order       │           │
│                          └────────────────┘ └──────────────────┘           │
│                                                                             │
│  ┌───────────────────────────┐                                             │
│  │      invitations          │                                             │
│  │       (BIGINT)            │                                             │
│  │───────────────────────────│                                             │
│  │ gallery_id FK ──> galleries (nullable)                                  │
│  │ album_id   FK ──> albums    (nullable)                                  │
│  │ email                     │                                             │
│  │ role (owner|editor|viewer)│                                             │
│  │ token (unique)            │                                             │
│  │ status (pending|accepted|revoked|expired)                               │
│  │ invited_by     FK ──> users                                             │
│  │ accepted_user_id FK ──> users (nullable)                                │
│  │ CHECK: exactly one of gallery_id/album_id is NOT NULL                   │
│  └───────────────────────────┘                                             │
│                                                                             │
│  ┌──────────────┐           ┌──────────────────┐                           │
│  │    tags      │           │   asset_tags     │                           │
│  │  (BIGINT)   │           │  (composite PK)  │                           │
│  │──────────────│           │──────────────────│                           │
│  │ gallery_id FK ──> galleries                 │                           │
│  │ name        │           │ asset_id FK ──> assets                        │
│  │ created_by FK ──> users │ tag_id   FK ──> tags                          │
│  │ UNIQUE(gallery_id,name) └──────────────────┘                           │
│  └──────────────┘                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                             ASSETS                                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────┐                              │
│  │               assets                     │                              │
│  │               (UUID)                     │                              │
│  │──────────────────────────────────────────│                              │
│  │ gallery_id  FK ──> galleries             │                              │
│  │ uploaded_by FK ──> users (nullable)      │                              │
│  │ name                                     │                              │
│  │ type (image|video|document|archive|file) │                              │
│  │ mime_type                                │                              │
│  │ size                                     │                              │
│  │ hash (SHA-256, dedup key)                │                              │
│  │ metadata (JSONB)                         │                              │
│  │ deleted_at (soft delete)                 │                              │
│  │ deleted_by FK ──> users (nullable)       │                              │
│  │                                          │                              │
│  │ UNIQUE(gallery_id, hash) WHERE deleted_at IS NULL                       │
│  └──────────────────────────────────────────┘                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                             STORAGE                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌────────────────────────────────────┐                                    │
│  │       storage_providers            │                                    │
│  │          (BIGINT)                  │                                    │
│  │────────────────────────────────────│                                    │
│  │ layer (serving|archive)            │                                    │
│  │ name                               │                                    │
│  │ type (cloudinary|imagekit|r2|gcs)  │                                    │
│  │ credentials (JSONB)                │                                    │
│  │ quota                              │                                    │
│  │ used                               │                                    │
│  │ is_active                          │                                    │
│  │ created_by FK ──> users            │                                    │
│  └─────────────┬──────────────────────┘                                    │
│                │                                                            │
│                ▼                                                            │
│  ┌────────────────────────────────────┐                                    │
│  │       storage_locations            │                                    │
│  │           (UUID)                   │                                    │
│  │────────────────────────────────────│                                    │
│  │ asset_id    FK ──> assets          │                                    │
│  │ provider_id FK ──> storage_providers (RESTRICT)                         │
│  │ layer (serving|archive)            │                                    │
│  │ provider_key                       │                                    │
│  │ url                                │                                    │
│  │ status (pending|stored|failed)     │                                    │
│  │ metadata (JSONB)                   │                                    │
│  │                                    │                                    │
│  │ UNIQUE(asset_id, provider_id)      │                                    │
│  └────────────────────────────────────┘                                    │
│                                                                             │
│  ┌────────────────────────────────────┐                                    │
│  │       archive_sync_jobs            │                                    │
│  │          (BIGINT)                  │                                    │
│  │────────────────────────────────────│                                    │
│  │ asset_id     FK ──> assets         │                                    │
│  │ target_layer (default: archive)    │                                    │
│  │ provider_id  FK ──> storage_providers (nullable)                        │
│  │ status (pending|running|completed|failed)                               │
│  │ attempts / max_attempts            │                                    │
│  │ last_error                         │                                    │
│  │ next_retry_at                      │                                    │
│  │                                    │                                    │
│  │ UNIQUE(asset_id, target_layer) WHERE status IN (pending, running)       │
│  └────────────────────────────────────┘                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Relationship Summary

| From | To | Cardinality | FK Column | Notes |
|------|----|-------------|-----------|-------|
| users | galleries | 1:N | `galleries.owner_id` | Each user owns 0+ galleries |
| users | galleries | 1:1 (partial) | `galleries.owner_id` | Exactly one `is_default` per user |
| galleries | gallery_members | 1:N | `gallery_members.gallery_id` | Owner also gets a member row |
| galleries | albums | 1:N | `albums.gallery_id` | Albums belong to one gallery |
| galleries | tags | 1:N | `tags.gallery_id` | Tag vocabulary is per-gallery |
| galleries | assets | 1:N | `assets.gallery_id` | Assets belong to one gallery |
| albums | album_members | 1:N | `album_members.album_id` | |
| albums | album_assets | 1:N | `album_assets.album_id` | Join table for N:M |
| assets | album_assets | 1:N | `album_assets.asset_id` | An asset can be in many albums |
| assets | asset_tags | 1:N | `asset_tags.asset_id` | |
| tags | asset_tags | 1:N | `asset_tags.tag_id` | |
| assets | storage_locations | 1:N | `storage_locations.asset_id` | >= 1 serving + >= 1 archive |
| storage_providers | storage_locations | 1:N | `storage_locations.provider_id` | RESTRICT delete |
| assets | archive_sync_jobs | 1:N | `archive_sync_jobs.asset_id` | Background replication |
| users | user_roles | 1:N | `user_roles.user_id` | Global RBAC |
| roles | user_roles | 1:N | `user_roles.role_id` | |
| roles | role_permissions | 1:N | `role_permissions.role_id` | |
| permissions | role_permissions | 1:N | `role_permissions.permission_id` | |
| users | cli_sessions | 1:N | `cli_sessions.user_id` | Multiple concurrent sessions |
| invitations | galleries OR albums | N:1 | `gallery_id` XOR `album_id` | CHECK constraint |

---

## Cardinality Patterns

```
users ──1:N──> galleries ──1:N──> assets ──1:N──> storage_locations
                   │                  │                    │
                   │                  │                    └──N:1──> storage_providers
                   │                  │
                   │                  ├──N:M──> albums     (via album_assets)
                   │                  └──N:M──> tags       (via asset_tags)
                   │
                   ├──1:N──> gallery_members
                   └──1:N──> albums ──1:N──> album_members

users ──N:M──> roles ──N:M──> permissions   (via user_roles, role_permissions)
```

---

**Next:** [Design Standards](./design-standards.md) — Naming conventions, ID strategy, indexing rules.
