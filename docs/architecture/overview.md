# Architecture Overview

## High level

```
                 Family users
                      │
        ┌─────────────┴─────────────┐
        │                           │
   Web (React 19)              CLI (Go, Cobra)
        │                           │
        └─────────────┬─────────────┘
                      │  HTTPS / REST
                 API (Go + Fiber)          ← all business logic
                      │
   ┌──────────────────┼───────────────────┐
   │                  │                    │
Services         Storage orchestrator   Clerk (identity)
   │                  │
Repositories     Storage adapters ── Cloudinary / ImageKit / R2 / GCS
   │
PostgreSQL (Neon)  ← source of truth
```

Filora is **API-first**: all business logic lives in the Go API. The web app and
CLI are **thin clients** that only handle presentation and user interaction.

## The three apps

Filora is a monorepo of three **independent** applications (no shared packages;
each app stands alone).

### `apps/api` — the backend
- **Go 1.23+, Fiber v3, sqlc, pgx/v5, PostgreSQL (Neon).**
- Owns all business logic, validation, authorization, and storage orchestration.
- Modular **vertical-slice** structure — each module owns its full stack.

### `apps/web` — the front-end
- **React 19, TypeScript, TanStack Query + Router, Tailwind v4, shadcn/ui, Zod.**
- Talks to the API over REST. Login/session via Clerk.

### `apps/cli` — the terminal client *(planned)*
- **Go, Cobra.** A thin HTTP client over the API.
- Supports terminal login and **multiple concurrent sessions**.

## API module structure (vertical slice)

Each feature module owns its layers, instead of splitting by technical layer:

```
apps/api/internal/modules/<module>/
├── handler.go       # HTTP routes: parse + validate request, call service, format response
├── service.go       # Business logic & orchestration
├── repository.go    # Database access (sqlc-generated queries)
└── models.go        # Data structures
```

Planned modules (aligned to the domain): `account` (Clerk users), `rbac`
(roles/permissions), `session` (CLI tokens), `gallery`, `album`, `tag`, `asset`,
`storage` (providers + adapters + orchestration), `dashboard`.

> The current code has a smaller/legacy module set (account/asset/storage/
> dashboard). See [roadmap](../product/roadmap.md).

## Layering & responsibilities

| Layer | Responsibility | Must not |
|-------|----------------|----------|
| Handler | Validate input, call a service, shape the HTTP response | Contain business logic |
| Service | Business rules, orchestration, permission checks | Touch the DB directly |
| Repository | Persistence via sqlc | Contain business logic |
| Adapter | Talk to a specific cloud provider | Leak provider types into business logic |

## Request / response flow

```
HTTP request
  → Fiber handler        (validate with validator v10)
  → Auth middleware      (Clerk token OR CLI token → current user)
  → Service              (RBAC + membership checks, business logic)
  → Repository           (sqlc queries)
  → PostgreSQL
  ← map rows → models → service → handler → JSON response
```

### Standard response envelope

Success:
```json
{ "success": true, "data": {} }
```
Error:
```json
{ "success": false, "error": { "code": "ERROR_CODE", "message": "Human readable message" } }
```

Errors propagate up the stack (`Repository → Service → Handler`) using Go error
wrapping (`%w`); the handler converts them to the right HTTP status.

## Data & storage boundaries

- **PostgreSQL is the source of truth.** Cloud providers are never authoritative.
- **Storage is abstracted.** Business logic depends on a `StorageAdapter`
  interface, never on a provider SDK. See [storage.md](./storage.md).
- **Two storage layers.** Every asset is stored on a serving layer and copied to
  an archive layer.

## Key principles

1. **API-first** — all logic in the API; clients are thin.
2. **Storage abstraction** — providers are implementation details behind adapters.
3. **Database as truth** — Postgres is authoritative; metadata first.
4. **Type safety** — sqlc for compile-time-checked SQL; Zod on the web edge.
5. **Explicit errors** — handle and wrap errors; no silent failures.
6. **Context propagation** — `context.Context` throughout for cancellation/timeouts.
7. **Simplicity over flexibility** — family scale; no premature abstraction.

## Non-goals

Microservices, event sourcing, CQRS, domain events, DI containers, plugin
frameworks. Keep it a single, simple, well-organized modular monolith.

## See also

- [Storage architecture](./storage.md)
- [Auth & access enforcement](./auth.md)
- [Database design](../database/README.md)
- [`/AGENTS.md`](../../AGENTS.md) — detailed engineering rules & conventions
