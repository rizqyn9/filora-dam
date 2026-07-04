# Filora DAM API Reference

REST API for the Filora Digital Asset Management platform. This reflects the
current implementation (target design: Clerk auth, RBAC, galleries/albums,
two-layer storage).

> **Status:** endpoints below are implemented. Physical upload/archive run
> through the `StorageAdapter` abstraction; concrete provider SDKs
> (Cloudinary/ImageKit/R2/GCS) are stubbed until credentials + SDK wiring are
> added, so `POST .../assets` returns an error until a real adapter is configured.

## Base URL

```
http://localhost:<PORT>/api/v1
```

## Authentication

All endpoints except health and the Clerk webhook require:

```
Authorization: Bearer <token>
```

`<token>` is either:
- a **Clerk session token** (web), or
- a **Filora CLI token** (prefixed `flr_`) obtained from `POST /cli/sessions`.

Both resolve to the same authenticated user (a local mirror of the Clerk
identity, auto-created on first request).

## Authorization

Two tiers (see [docs/database/rbac.md](../../docs/database/rbac.md)):
- **Global RBAC**: roles → permissions (`resource:action`) with scope `own`/`all`.
- **Per-resource membership**: `owner`/`editor`/`viewer` on galleries and albums.

## Response envelope

Success:
```json
{ "success": true, "data": { } }
```
Error:
```json
{ "success": false, "error": { "code": "ERROR_CODE", "message": "..." } }
```

| Status | Code (example) | Meaning |
|--------|----------------|---------|
| 400 | `BAD_REQUEST`, `VALIDATION_ERROR` | Invalid input |
| 401 | `UNAUTHORIZED` | Missing/invalid token |
| 403 | `FORBIDDEN` | Insufficient permission/role |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Conflict (e.g. invitation not pending) |
| 503 | `NO_SERVING_STORAGE` | No active serving storage account |
| 507 | `INSUFFICIENT_STORAGE` | Gallery quota exceeded |
| 500 | `INTERNAL` | Server error |

---

## Health

| Method | Path | Notes |
|--------|------|-------|
| GET | `/` | Service info |
| GET | `/health` | `{ status, database }` |

## Account

| Method | Path | Description |
|--------|------|-------------|
| GET | `/me` | Current user profile |
| PATCH | `/me` | Update `{ name, avatar_url? }` |
| POST | `/webhooks/clerk` | Clerk (Svix) webhook — signature-verified, public |

## CLI sessions

| Method | Path | Description |
|--------|------|-------------|
| POST | `/cli/sessions` | Issue token `{ label? }` → `{ token, session }` (token shown once) |
| GET | `/cli/sessions` | List active sessions |
| DELETE | `/cli/sessions/:id` | Revoke a session |
| POST | `/cli/sessions/revoke-all` | Revoke all sessions |

## RBAC (requires `role:*`)

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/rbac/roles` | `role:read` |
| POST | `/rbac/roles` | `role:manage` |
| GET | `/rbac/roles/:id` | `role:read` |
| PATCH | `/rbac/roles/:id` | `role:manage` |
| DELETE | `/rbac/roles/:id` | `role:manage` (system roles protected) |
| GET | `/rbac/roles/:id/permissions` | `role:read` |
| POST | `/rbac/roles/:id/permissions` | `role:manage` — `{ permission_id, scope }` |
| DELETE | `/rbac/roles/:id/permissions/:permissionId` | `role:manage` |
| GET | `/rbac/permissions` | `role:read` |
| POST | `/rbac/permissions` | `role:manage` |
| GET | `/rbac/users/:id/roles` | `role:read` |
| POST | `/rbac/users/:id/roles` | `role:assign` — `{ role_id }` |
| DELETE | `/rbac/users/:id/roles/:roleId` | `role:assign` |

## Galleries

| Method | Path | Access |
|--------|------|--------|
| POST | `/galleries` | `gallery:create` |
| GET | `/galleries` | member's galleries |
| GET | `/galleries/:id` | member (viewer+) |
| PATCH | `/galleries/:id` | editor+ |
| DELETE | `/galleries/:id` | owner (default gallery protected) |
| GET | `/galleries/:id/members` | member |
| PATCH | `/galleries/:id/members/:userId` | owner — `{ role }` |
| DELETE | `/galleries/:id/members/:userId` | owner |
| POST | `/galleries/:id/invitations` | owner — `{ email, role: editor\|viewer }` |
| GET | `/galleries/:id/invitations` | owner |
| DELETE | `/galleries/:id/invitations/:invId` | owner |
| POST | `/invitations/accept` | invited user — `{ token }` (email must match) |

## Albums

| Method | Path | Access |
|--------|------|--------|
| POST | `/galleries/:galleryId/albums` | gallery editor+ |
| GET | `/galleries/:galleryId/albums` | gallery member |
| GET | `/albums/:id` | album/gallery viewer+ |
| PATCH | `/albums/:id` | editor+ — `{ name, description?, cover_asset_id? }` |
| DELETE | `/albums/:id` | owner |
| GET | `/albums/:id/members` | viewer+ |
| POST | `/albums/:id/members` | owner — `{ user_id, role: editor\|viewer }` |
| DELETE | `/albums/:id/members/:userId` | owner |
| GET | `/albums/:id/assets` | viewer+ (returns asset ids) |
| POST | `/albums/:id/assets` | editor+ — `{ asset_id, sort_order? }` |
| DELETE | `/albums/:id/assets/:assetId` | editor+ |

## Tags

| Method | Path | Access |
|--------|------|--------|
| POST | `/galleries/:galleryId/tags` | gallery editor+ — `{ name }` |
| GET | `/galleries/:galleryId/tags` | gallery viewer+ |
| PATCH | `/tags/:id` | gallery editor+ |
| DELETE | `/tags/:id` | gallery editor+ |
| POST | `/tags/:id/assets` | gallery editor+ — `{ asset_id }` |
| DELETE | `/tags/:id/assets/:assetId` | gallery editor+ |

## Assets

| Method | Path | Access |
|--------|------|--------|
| POST | `/galleries/:galleryId/assets` | editor+ — `multipart/form-data` field `file` |
| GET | `/galleries/:galleryId/assets` | viewer+ — `?limit&offset` |
| GET | `/galleries/:galleryId/assets/search` | viewer+ — `?q=&limit&offset` |
| GET | `/galleries/:galleryId/assets/filter/:type` | viewer+ — type: image/video/document/archive/file |
| GET | `/galleries/:galleryId/assets/trash` | editor+ |
| GET | `/assets/:id` | viewer+ |
| PATCH | `/assets/:id` | editor+ — `{ name }` |
| DELETE | `/assets/:id` | editor+ (soft delete → trash) |
| POST | `/assets/:id/restore` | editor+ |
| GET | `/assets/:id/download` | viewer+ — `302` redirect to serving URL |

Upload behavior: SHA-256 dedup per gallery (duplicate returns the existing
asset), gallery quota enforced (`507`), stored on the serving layer, and an
archive replication job is enqueued.

## Storage (requires `storage:*`)

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/storage/providers` | `storage:read` |
| POST | `/storage/providers` | `storage:create` — `{ layer, name, type, credentials, quota? }` |
| GET | `/storage/providers/:id` | `storage:read` |
| PATCH | `/storage/providers/:id` | `storage:update` |
| DELETE | `/storage/providers/:id` | `storage:delete` (deactivates) |
| GET | `/storage/usage` | `storage:read` — per-account usage summary |

`credentials` are never returned. Provider `type`: `cloudinary`, `imagekit`,
`r2`, `gcs`. `layer`: `serving`, `archive`.

## Dashboard

| Method | Path | Access |
|--------|------|--------|
| GET | `/galleries/:galleryId/dashboard` | gallery member — stats, type counts, recent, quota |
| GET | `/dashboard/system` | `dashboard:read` scope `all` — archive job health |

---

## Notes

- Pagination: `limit` (default 20, max 100), `offset` (default 0).
- Timestamps are ISO-8601 UTC.
- Sizes are bytes.
- IDs: users/roles/galleries/albums/tags/providers are integers; assets and CLI
  sessions are UUID v7.
