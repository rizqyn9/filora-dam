# Feature Catalog

What Filora does, grouped by area. Each feature notes its **status**:

- ✅ **designed** — modeled in the database/design; ready to implement
- 🟡 **partial** — some legacy implementation exists but differs from target
- 🔭 **backlog** — planned, intentionally deferred

See [roadmap.md](./roadmap.md) for sequencing and the legacy→target gap.

---

## Accounts & access

| Feature | Status | Notes |
|---------|--------|-------|
| Web login & sessions via Clerk | ✅ designed | No passwords stored by Filora |
| User mirror (webhook + JIT) | ✅ designed | Local `users` row synced from Clerk |
| Terminal (CLI) login | ✅ designed | Opaque, hashed tokens |
| Multiple concurrent CLI sessions | ✅ designed | Each session labeled & independently revocable |
| Global RBAC (roles + scoped permissions) | ✅ designed | `superuser`/`admin`/`member`/`viewer` |
| Superuser (full access) | ✅ designed | Wildcard permission |
| Legacy JWT + password auth | 🟡 partial | Current API code; to be replaced by Clerk |

## Galleries

| Feature | Status | Notes |
|---------|--------|-------|
| Default gallery per user | ✅ designed | Created on provisioning |
| Own multiple galleries | ✅ designed | |
| Join other users' galleries | ✅ designed | Via membership |
| Gallery membership roles | ✅ designed | `owner` / `editor` / `viewer` |
| Invite to a gallery by email | ✅ designed | `invitations` |
| Per-gallery storage quota | ✅ designed | `galleries.storage_quota` |

## Albums

| Feature | Status | Notes |
|---------|--------|-------|
| Create albums within a gallery | ✅ designed | |
| Add an asset to multiple albums | ✅ designed | Many-to-many |
| Album cover image | ✅ designed | Optional `cover_asset_id` |
| Manual ordering of album assets | ✅ designed | `sort_order` |
| Invite to an album by email | ✅ designed | Album owner invites |

## Assets & organization

| Feature | Status | Notes |
|---------|--------|-------|
| Upload photos/videos/docs/other | 🟡 partial | Legacy upload works; target flow differs (2 layers) |
| Automatic type & MIME detection | 🟡 partial | image/video/document/archive/file |
| Per-gallery deduplication (SHA-256) | ✅ designed | Legacy dedup was per-user |
| Tagging (normalized, per gallery) | ✅ designed | Legacy used a text array |
| Search by name | 🟡 partial | Legacy endpoint exists |
| Filter by type / tag | 🟡 partial | |
| Soft delete (trash) & recovery | ✅ designed | `deleted_at` |
| Download | 🟡 partial | Served from the serving layer |

## Storage

| Feature | Status | Notes |
|---------|--------|-------|
| Two-layer storage (serving + archive) | ✅ designed | Every asset in both layers |
| Serving layer (Cloudinary/ImageKit) | 🟡 partial | Adapters exist for Cloudinary/ImageKit/R2 |
| Archive layer (GCS Archive / R2) | 🔭 backlog | GCS adapter not built yet |
| Multiple accounts per layer | ✅ designed | Spreads capacity past free-tier caps |
| Global, admin-managed accounts | ✅ designed | Legacy accounts were per-user |
| Per-account usage summary | ✅ designed | `storage_account_usage` view |
| Async archive replication + retry | ✅ designed | `archive_sync_jobs` |
| Account election strategy | 🔭 backlog | Which account a new upload lands on |

## Dashboard & insights

| Feature | Status | Notes |
|---------|--------|-------|
| Storage usage summary | 🟡 partial | Legacy dashboard exists; will move to gallery/account views |
| Asset counts by type | 🟡 partial | |
| Recent activity | 🟡 partial | |

## Clients

| Feature | Status | Notes |
|---------|--------|-------|
| Web app (React 19) | 🟡 partial | Scaffolded |
| CLI (Go / Cobra) | 🔭 backlog | Not started |
| REST API (Go / Fiber) | 🟡 partial | Legacy endpoints live; redesign pending |

## Deferred (explicitly out of MVP)

- Disaster-recovery orchestration / scheduled backups beyond the archive copy.
- Asset versioning / history.
- Favorites, public share links, notifications.
- Analytics/observability stack (OpenTelemetry, Grafana LGTM).

See [roadmap.md](./roadmap.md) for details.
