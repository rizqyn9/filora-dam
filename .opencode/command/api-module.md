---
description: Scaffold a new API module (vertical slice) following AGENTS.md
agent: build
---

Scaffold a new Go API feature module named `$ARGUMENTS` under
`apps/api/internal/modules/$ARGUMENTS/`, following the vertical-slice pattern in
`AGENTS.md` and `docs/architecture/project-structure.md`.

Before writing, read an existing module (e.g. `internal/modules/session/`) and
match its style exactly. Then create these files:

- `models.go` — request/response DTOs and domain structs (validator v10 tags)
- `repository.go` — wraps sqlc queries, maps rows to module models; no business logic
- `service.go` — business rules, orchestration, authz checks; takes `context.Context` first; no Fiber, no direct DB
- `handler.go` — thin: parse+validate request, call service, return `lib.Response` envelope
- `routes.go` — `RegisterRoutes(router, deps)`

Constraints (from AGENTS.md — do not violate):
- Database First: if new data access is needed, add `schema.sql` + `queries/$ARGUMENTS.sql` and note that `make sqlc` must be run. Do NOT invent an ORM.
- No generic/base repository, no factories, no DI container.
- Cross-module needs use consumer-defined interfaces, injected at `cmd/server/main.go`.
- Keep handlers thin; business logic stays in the service.

Do not wire the module into `main.go` unless I ask — just note where it should be
registered. Finish by running `go build ./...` in `apps/api` to confirm it compiles.
