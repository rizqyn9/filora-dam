# API Implementation Plan

A phased, database-first plan to rebuild `apps/api` for the target design. It
operationalizes [project-structure.md](./project-structure.md), the
[architecture](./overview.md), and the [database design](../database/README.md).

**Working style** (from [`/AGENTS.md`](../../AGENTS.md)): build first, keep it
simple, no premature abstraction. Each phase is independently verifiable and
leaves the app in a compiling, runnable state.

**Per-module Definition of Done:** `schema (exists)` → `queries/*.sql` →
`make sqlc` → `repository` → `service (+authz)` → `handler` → `routes` →
`validation` → `tests` → wired in compose root.

---

## Phase dependency graph

```
0 Skeleton
└─ 1 Identity (account + Clerk auth)
   ├─ 2 Authorization (auth + rbac)
   │  ├─ 3 CLI sessions
   │  └─ 4 Galleries (+ default-gallery provisioning)
   │     └─ 5 Albums & Tags
   6 Storage accounts (after 2)
   └─ 7 Assets (needs 4 + 6)
      ├─ 8 Archive worker (needs 6 + 7)
      └─ 9 Dashboard (needs 7)
10 Hardening & docs (last)
```

---

## Phase 0 — Skeleton & tooling

**Objective:** a runnable, empty API that compiles and serves `/health`.

Tasks:
- Confirm `go.mod` module path; add base deps: Fiber v3, pgx/v5, godotenv,
  validator v10, zerolog.
- `internal/config/config.go` — load & validate env (validator v10).
- `internal/database/db.go` — pgx pool (`New`, `Close`, `Ping`).
- `internal/lib/` — `response.go` (envelope), `errors.go` (app error → HTTP),
  `validate.go`, `pagination.go`, `hash.go`, `mime.go`.
- `internal/server/` — Fiber app, middleware (`requestid`, `logger`, `recover`,
  `cors`), central error handler, `GET /` and `GET /health`.
- `cmd/server/main.go` — compose root; graceful shutdown.
- Verify `make sqlc` runs against `schema.sql` (empty queries OK).

**Acceptance:** `make build` succeeds; `make run` serves `GET /health` →
`{ "success": true, "data": { "status": "ok" } }`.

**Env keys:** `PORT`, `ENV`, `DATABASE_URL`.

---

## Phase 1 — Identity (users + Clerk)

**Objective:** authenticated requests resolve to a local `users` row.

Tasks:
- `queries/users.sql`: get by id / clerk_user_id / email, upsert, deactivate,
  `last_seen_at`.
- `modules/account/`: repository + service + handler for the current user
  (`GET /me`, `PATCH /me`).
- Clerk web auth: `middleware/auth.go` verifies the Clerk session token
  (clerk-sdk-go v2) → `clerk_user_id` → **load-or-JIT-create** user → context.
- `internal/auth/context.go`: `WithUser` / `UserFromContext`.
- Clerk webhook endpoint (`POST /webhooks/clerk`): verify signature (Svix),
  dedupe via `clerk_webhook_events`, handle `user.created/updated/deleted`.

**Acceptance:** a request with a valid Clerk token returns the mirrored user;
first-time user is created (JIT); webhook replays are idempotent.

**Env keys:** `CLERK_SECRET_KEY`, `CLERK_WEBHOOK_SIGNING_SECRET`.

**Note:** default-gallery provisioning is added in Phase 4 (needs the gallery module).

---

## Phase 2 — Authorization (RBAC)

**Objective:** enforce global RBAC; manage roles/permissions.

Tasks:
- `queries/rbac.sql`: effective permissions for a user; roles/permissions CRUD;
  assign/revoke user roles.
- `internal/auth/authorizer.go` + `repository.go`: `Authorize(ctx, user,
  resource, action) → Decision{Allowed, Scope}` with wildcard handling
  (`*:*`, `resource:*`), widest-scope wins.
- `modules/rbac/`: admin CRUD for roles, permissions, and user-role assignment
  (guarded by `role:*`).
- Helper in `lib`/`auth` for services: `RequireAll` / `RequireOwnOr...`.

**Acceptance:** `Authorize` returns correct decisions for seeded roles;
superuser bypasses; admin cannot `role:manage`; a member gets `own` scope.

---

## Phase 3 — CLI sessions

**Objective:** terminal login with multiple concurrent sessions.

Tasks:
- `queries/cli_sessions.sql`: create, lookup by `token_hash`, list active,
  touch `last_used_at`, revoke, revoke-all.
- `modules/session/`: issue token (return raw once; store `sha256`), list, revoke.
  Endpoints under `/cli/sessions`.
- Extend `middleware/auth.go`: accept `Authorization: Bearer <cli-token>`; hash,
  look up active session (not revoked/expired) → user; update `last_used_at`.

**Acceptance:** issuing a token returns it once; the token authenticates API
calls; revoking it blocks further use; multiple sessions coexist.

**Env keys:** `CLI_TOKEN_TTL_HOURS`.

---

## Phase 4 — Galleries (+ membership, invitations)

**Objective:** galleries with sharing; every user has a default gallery.

Tasks:
- `queries/galleries.sql`: gallery CRUD; members (upsert/list/remove);
  invitations (create/lookup by token/list/accept/revoke); default-gallery lookup.
- `modules/gallery/`:
  - CRUD (owner also inserted as `owner` member).
  - Membership management (`invite`, list, change role, remove).
  - Invitations: create (email + role + token), accept (creates membership +
    sets `accepted_user_id`), revoke.
  - Expose `Membership(ctx, galleryID, userID) (role, error)` for other modules.
- **Provisioning hook:** on user create (Phase 1 JIT/webhook), create the
  default gallery + `owner` membership.
- Authz: `gallery:*` global + membership for `own`.

**Acceptance:** new user gets a default gallery; owner can invite by email;
invitee accepts and gains membership; viewers cannot mutate.

---

## Phase 5 — Albums & Tags

**Objective:** organize assets within a gallery.

Tasks:
- **Album** — `queries/albums.sql`; `modules/album/`: album CRUD (in a gallery),
  members (invite), `album_assets` add/remove/reorder, cover. Consumes
  `gallery.Membership` via a consumer-defined interface.
- **Tag** — `queries/tags.sql`; `modules/tag/`: tag CRUD per gallery,
  `asset_tags` attach/detach, list tags & filter assets by tag.

**Acceptance:** album groups assets (M2M), album invites work; tags are unique
per gallery; assets can be filtered by tag.

*(Depends on Phase 7 for real assets; can be built with asset stubs and
finalized after Phase 7.)*

---

## Phase 6 — Storage accounts

**Objective:** admin-managed, global storage accounts across two layers.

Tasks:
- `queries/storage.sql`: providers CRUD; usage; `storage_account_usage` reads.
- `modules/storage/`:
  - Provider account CRUD (guarded by `storage:*`); layer + type validation;
    credentials stored in `credentials` JSONB.
  - `adapters/adapter.go` — `StorageAdapter` interface.
  - Serving adapters: `cloudinary.go`, `imagekit.go`. (`r2.go`, `gcs.go` are
    stubs here; archive `gcs.go` completed in Phase 8.)
  - Orchestration helpers: pick an active account in a layer (simple
    "first active with space"; **account election = backlog**).
  - Usage summary endpoint (reads the view).

**Acceptance:** admin can register/deactivate accounts; usage summary lists
accounts with used/quota and copy counts; adapters upload/download/delete/exist
against a real serving provider.

---

## Phase 7 — Assets (upload/download/trash)

**Objective:** the core DAM flow on the serving layer.

Tasks:
- `queries/assets.sql`: create; get; list by gallery (active only); per-gallery
  dedup lookup; soft-delete/restore; storage_locations create/list.
- `modules/asset/`:
  - Upload: validate → SHA-256 → per-gallery dedup → gallery quota check →
    pick serving account → `StorageAdapter.Upload` → persist `assets` +
    `storage_location(serving, stored)` → **enqueue `archive_sync_jobs`**.
  - Download: permission check → serving location → signed URL / stream.
  - Metadata update, list/search/filter, tagging hooks.
  - Trash: soft-delete (`deleted_at`) + restore.
  - Consumer interfaces: `StorageService`, `GalleryAccess`, injected at wiring.
- Update `galleries.storage_used` on upload/delete.

**Acceptance:** upload stores to serving + creates a pending archive job; dedup
prevents duplicates per gallery; quota enforced; delete moves to trash;
re-upload after delete allowed.

---

## Phase 8 — Archive worker

**Objective:** async replication to the archive layer with retry.

Tasks:
- Complete `adapters/gcs.go` (GCS Archive class); ensure `r2.go` usable for archive.
- `queries/jobs.sql`: claim due jobs (`pending`/`failed`, `next_retry_at ≤ now`,
  `FOR UPDATE SKIP LOCKED`), mark running/completed/failed, backoff.
- `cmd/worker/main.go`: loop → claim → pick archive account → upload →
  create `storage_location(archive, stored)` → complete; on error increment
  `attempts`, set `next_retry_at`, `failed` until `max_attempts`.

**Acceptance:** after an upload, the worker creates an archive copy and marks the
job completed; failures retry with backoff and stop at `max_attempts`.

---

## Phase 9 — Dashboard

**Objective:** summaries for the UI.

Tasks:
- `modules/dashboard/`: per-gallery asset counts/size, counts by type, recent
  assets, storage usage (via `storage_account_usage`), archive job health.

**Acceptance:** dashboard endpoint returns accurate, permission-scoped metrics.

---

## Phase 10 — Hardening & docs

**Objective:** production-readiness for the family MVP.

Tasks:
- Consistent validation (validator v10) on every external input.
- Central error mapping; structured logging (zerolog) for uploads/downloads/
  deletes/storage & backup failures.
- Tests: repository + service + key workflow (integration-first, minimal mocking).
- Rewrite `apps/api/API.md` for the new endpoints; remove legacy banners;
  refresh `TESTING*.md`; update `features.md` statuses.
- Review env/`.env.example`; finalize `Makefile` targets.

**Acceptance:** `make lint` and `make test` pass; docs match the implemented API.

---

## Cross-cutting standards

- **Context:** `context.Context` first arg everywhere.
- **Errors:** wrap with `%w`; convert to HTTP in the handler/error handler.
- **Response envelope:** `{ success, data }` / `{ success, error{code,message} }`.
- **Security:** never store passwords or raw CLI tokens; validate all input.
- **DB:** `schema.sql` is truth; apply manually; regenerate sqlc after query changes.

## Dependencies to add (per phase, not upfront)

| Concern | Library |
|--------|---------|
| Web framework | `github.com/gofiber/fiber/v3` |
| Postgres driver | `github.com/jackc/pgx/v5` |
| Validation | `github.com/go-playground/validator/v10` |
| Env | `github.com/joho/godotenv` |
| Logging | `github.com/rs/zerolog` |
| Clerk | `github.com/clerk/clerk-sdk-go/v2` |
| Webhook verify | Svix signature verification |
| Cloudinary / ImageKit | official Go SDKs |
| R2 (S3-compatible) | `github.com/aws/aws-sdk-go-v2` |
| GCS | `cloud.google.com/go/storage` |

## Backlog (out of these phases)

Account election strategy, Alibaba OSS Archive, asset versioning, trash purge &
restore-from-archive tooling, favorites/share links/notifications, OpenTelemetry.
See [product/roadmap.md](../product/roadmap.md).
