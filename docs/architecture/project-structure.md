# API Project Structure

The layout and dependency rules for `apps/api` (Go). This is the blueprint for
the rebuild. It follows the modular **vertical-slice** approach in
[`/AGENTS.md`](../../AGENTS.md): each feature module owns its full stack, and we
add abstractions only when a second concrete case appears.

Related: [architecture overview](./overview.md) · [auth](./auth.md) ·
[storage](./storage.md) · [database](../database/README.md)

---

## Directory tree

```
apps/api/
├── cmd/
│   ├── server/main.go        # HTTP API — compose root (wires everything)
│   └── worker/main.go        # Background worker — compose root
├── internal/
│   ├── config/               # Env loading + validation (validator v10)
│   │   └── config.go
│   ├── database/
│   │   ├── db.go             # pgx pool + helpers
│   │   ├── schema.sql        # Canonical schema (source of truth, manual apply)
│   │   ├── seed.sql          # Baseline RBAC roles/permissions
│   │   ├── queries/          # sqlc source, one file per domain
│   │   │   ├── users.sql
│   │   │   ├── cli_sessions.sql
│   │   │   ├── rbac.sql
│   │   │   ├── galleries.sql
│   │   │   ├── albums.sql
│   │   │   ├── tags.sql
│   │   │   ├── assets.sql
│   │   │   ├── storage.sql
│   │   │   └── jobs.sql
│   │   └── db/               # sqlc-generated code (package "db") — do not edit
│   ├── server/               # Fiber app assembly
│   │   ├── server.go         # New(), Start(), graceful shutdown, error handler
│   │   └── routes.go         # Route registration entry point
│   ├── middleware/           # HTTP middleware
│   │   ├── auth.go           # Clerk session OR CLI token → current user in ctx
│   │   ├── requestid.go
│   │   ├── logger.go
│   │   └── recover.go
│   ├── auth/                 # Identity context + global RBAC authorizer
│   │   ├── context.go        # WithUser / UserFromContext
│   │   ├── authorizer.go     # Authorize(ctx, user, resource, action) → Decision
│   │   └── repository.go     # read effective permissions (sqlc)
│   ├── lib/                  # Neutral, dependency-free helpers
│   │   ├── response.go       # success/error envelope
│   │   ├── errors.go         # app error type → HTTP status/code mapping
│   │   ├── hash.go           # SHA-256
│   │   ├── mime.go           # MIME detection
│   │   ├── pagination.go     # limit/offset parsing
│   │   └── validate.go       # validator v10 wrapper
│   └── modules/              # Feature modules (vertical slice)
│       ├── account/          # users (Clerk mirror) + Clerk webhook sync
│       ├── session/          # CLI tokens (multi-session)
│       ├── rbac/             # roles/permissions/assignment (admin CRUD)
│       ├── gallery/          # galleries + members + invitations
│       ├── album/            # albums + members + album_assets
│       ├── tag/              # tags + asset_tags
│       ├── asset/            # assets, upload/download, trash
│       ├── storage/          # storage accounts + orchestration
│       │   └── adapters/     # cloudinary.go, imagekit.go, r2.go, gcs.go
│       └── dashboard/        # metrics / summaries
├── API.md · TESTING*.md      # legacy reference (banner-marked)
├── Makefile · sqlc.yaml · go.mod · go.sum · .env.example
```

---

## Module anatomy

Every module under `internal/modules/<name>/` owns the same files:

```
<module>/
├── handler.go     # HTTP: parse+validate request, call service, format response
├── service.go     # Business logic, orchestration, permission enforcement
├── repository.go  # Persistence: wraps sqlc queries, maps rows ↔ module models
├── models.go      # Request/response DTOs and domain structs
└── routes.go      # (optional) RegisterRoutes(router, deps)
```

Rules per layer:

| Layer | Does | Must NOT |
|-------|------|----------|
| handler | validate input, call service, shape HTTP response | hold business logic, touch DB |
| service | business rules, orchestration, authz checks | import Fiber, touch DB directly |
| repository | run sqlc queries, map to module models | contain business logic |
| adapter | talk to one cloud provider | leak provider SDK types upward |

Handlers receive `context.Context` and the current user from context; services
take `context.Context` as the first argument everywhere.

---

## Dependency rules (avoid import cycles)

Layered, one direction only. An arrow means "may import".

```
cmd/*  →  server  →  modules/*  →  auth, lib, database/db
                     modules/*  →  config (read-only), database (pool)
   middleware       →  auth, lib
   auth             →  database/db, lib
   lib              →  (nothing internal)
```

**Cross-module dependencies** use **consumer-defined interfaces** (Go idiom),
never direct imports of another module's concrete types:

- If `asset` needs to upload, it declares an interface it owns, e.g.
  `type StorageService interface { Store(ctx, ...) (...) }`, and the concrete
  `storage.Service` is injected at the compose root.
- If `asset` needs gallery membership, it declares
  `type GalleryAccess interface { Membership(ctx, galleryID, userID) (Role, error) }`,
  implemented by `gallery.Service`.

This keeps modules decoupled and cycle-free. Rough level ordering (lower may be
used by higher): `account`/`session`/`rbac`/`storage` → `gallery` →
`album`/`tag`/`asset` → `dashboard`.

`lib` has **no internal dependencies**. `auth` depends only on `database/db` and
`lib`. Modules never import `server`, `middleware`, or `cmd`.

---

## Authorization: two tiers, no cycles

See [auth.md](./auth.md) and [database/rbac.md](../database/rbac.md).

1. **Global RBAC** lives in `internal/auth`:
   `Authorize(ctx, user, resource, action) → Decision{Allowed, Scope}` where
   `Scope ∈ {own, all}`. It reads a user's effective permissions via its own
   small read repository. It knows nothing about galleries/albums.
2. **Membership (scope `own`)** is enforced **inside each module's service**
   using that module's repository (e.g. gallery service checks `gallery_members`).

So `auth` stays a low-level, domain-agnostic package; the `rbac` module only does
admin CRUD of roles/permissions and does **not** import `auth`.

Typical service check:
```go
dec, err := a.authz.Authorize(ctx, user, "asset", "delete")
if err != nil { return err }
if !dec.Allowed { return lib.ErrForbidden }
if dec.Scope == auth.ScopeOwn {
    // module-local membership/ownership check
    if err := s.assertGalleryEditor(ctx, asset.GalleryID, user.ID); err != nil { return err }
}
```

---

## Authentication (middleware)

`internal/middleware/auth.go` resolves the **current user** and puts it in the
request context:

- **Web**: verify the Clerk session token (Clerk SDK) → `clerk_user_id` →
  load-or-JIT-create the local `users` row.
- **CLI**: `Authorization: Bearer <token>` where the token is our opaque CLI
  token → SHA-256 → lookup active `cli_sessions` → user.

Downstream handlers/services read the user via `auth.UserFromContext(ctx)`.

---

## Data access & sqlc

- SQL lives in `internal/database/queries/*.sql` (one file per domain).
- `make sqlc` generates a single typed package `internal/database/db` (never
  edited by hand).
- Each module's `repository.go` wraps the generated queries and maps rows to its
  own `models.go` types. No generic/base repository (forbidden by AGENTS).
- The `schema.sql` is applied manually (`make db-apply`); there are no migrations.

Workflow to add/change data access: edit `schema.sql` → add query in
`queries/*.sql` → `make sqlc` → use it from the module repository.

---

## Compose root (wiring)

`cmd/server/main.go` is the only place that constructs concrete dependencies and
injects them — **no DI container**:

```
load config
open pgx pool (database.New)
build sqlc queries (db.New(pool))
construct authorizer (auth.NewAuthorizer(...))
construct repositories → services → handlers per module
  (inject cross-module interfaces here: asset ← storage, gallery, tag …)
build server (Fiber), register middleware + module routes
start with graceful shutdown
```

`cmd/worker/main.go` shares config/database/services but runs the background
loop instead of the HTTP server.

---

## Background worker

`cmd/worker` is a separate binary that processes `archive_sync_jobs`: it claims
due jobs, replicates assets to an archive-layer account via the storage adapters,
updates `storage_locations`, and handles retry/backoff. It reuses the same
`config`, `database`, and `storage` service code as the server. See
[storage.md](./storage.md#archive-replication-flow-async).

---

## Naming & conventions

- Packages: short, lower-case, singular (`gallery`, not `galleries`), matching
  the module directory.
- Exported constructors: `New`, `NewService`, `NewHandler`, `NewRepository`.
- Files: `handler.go`, `service.go`, `repository.go`, `models.go`, `routes.go`.
- Errors: return wrapped errors (`%w`); map to HTTP in the handler/error handler
  via `lib/errors.go`. Response envelope from `lib/response.go`.
- Every exported function takes `context.Context` first.
- Validate all external input with `lib/validate.go` (validator v10).

---

## What we deliberately avoid

Per [`/AGENTS.md`](../../AGENTS.md): no generic/base repositories, no repository/
service factories, no DI container, no CQRS/event sourcing, no plugin framework,
no microservices. One simple, well-organized modular monolith plus one worker.
