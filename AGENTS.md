# Filora Agent Operating Manual

Read this before any task. All AI agents (Claude Code, Kiro, Antigravity) follow
the same rules here.

## Required reading (before changing code)
1. `docs/README.md` — documentation index
2. `docs/product/overview.md` — what Filora is
3. `docs/architecture/overview.md` + `docs/architecture/project-structure.md` — how it fits together (dir trees, module anatomy, dependency rules live here)
4. `docs/database/README.md` — schema is the source of truth for data

## Project
Filora is a multi-cloud Digital Asset Management (DAM) platform. Users upload,
organize, sync, and manage digital assets. Storage complexity is fully abstracted
— users never know where/how/which account stores their files. Filora manages it.

**Stage: MVP, solo-developed.** Goals: ship working features fast, validate PMF,
stay maintainable, minimize ops. Do NOT optimize for hypothetical future needs.

## Tech stack
- **API** (`apps/api`): Go 1.23+, Fiber v3, sqlc, PostgreSQL (Neon), validator v10, golang-migrate
- **CLI** (`apps/cli`): Go — thin HTTP client, no business logic
- **Web** (`apps/web`): React 19, TypeScript, TanStack Query + Router, Tailwind v4, Shadcn UI, Zod v4, bun
- **Storage providers**: Cloudinary, ImageKit, Cloudflare R2. **Backup (planned)**: Alibaba OSS / GCS Archive

Three independent apps under `apps/`. No shared packages — each app is independent.

## Architecture (see docs/architecture/project-structure.md for the full blueprint)
- API uses **modular vertical-slice** architecture. Each module owns its full stack: `handler.go` (HTTP), `service.go` (business logic), `repository.go` (DB via sqlc), `models.go` (DTOs/structs), optional `routes.go`. Modules live in `internal/modules/<name>/`. Avoid layer-first layout.
- Cross-module needs use **consumer-defined interfaces**, injected at the compose root (`cmd/server/main.go`). No DI container.
- CLI is a thin REST client: `cmd/`, `internal/client/`, `internal/commands/`. No business logic.

## Decision hierarchy
Correctness > Simplicity > Readability > Developer Experience > Performance > Extensibility.
Never sacrifice simplicity for theoretical flexibility.

## Golden rules
- **Build first** — working software beats perfect architecture.
- **Refactor later** — no abstraction until ≥2 concrete implementations exist. Duplication is acceptable; wrong abstractions are expensive.
- **Database first** — design order: SQL migration → sqlc queries → repository → service → handler → UI.
- **Consistency first** — follow existing patterns; no personal architectural preferences.

## Layer responsibilities
- **Handler**: validate request, call service, return response. Stay thin — no business logic, no DB.
- **Service**: business rules, orchestration, workflows, authz checks. No Fiber, no direct DB.
- **Repository**: DB access/queries/persistence via sqlc-generated code. No business logic.
- **Adapter**: talk to exactly one cloud provider. Never leak provider SDK types upward.

## Storage rules
Storage providers are implementation details. Business logic must NEVER depend on
Cloudinary/ImageKit/R2 SDKs directly — only adapters may. All adapters implement:
```go
type StorageAdapter interface {
    Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

## Metadata rule
Files live in storage. **Truth lives in PostgreSQL.** Never use cloud storage as
the source of truth for business logic.

## API is the source of business logic
Web and CLI are thin clients. Business logic must NOT live in React components,
CLI commands, or handlers.

## Validation
Validate ALL external input (HTTP body/query/route params, env vars, webhooks, CLI
args) with validator v10. Never trust external data.

## Go rules
- `gofmt` + `golangci-lint`. Handle errors explicitly — never ignore. Wrap with `%w`:
  ```go
  result, err := service.DoSomething(ctx)
  if err != nil { return fmt.Errorf("failed to do something: %w", err) }
  ```
- Interfaces for abstraction (StorageAdapter, Repository). Composition over inheritance. `context.Context` first arg everywhere. PascalCase exported / camelCase unexported.
- Avoid: `panic()` (except unrecoverable startup), `interface{}`, globals (except config).
- Keep functions small. Use concise names (`asset`, not `assetEntityAggregateRoot`).

## Database rules
SQL migrations are the source of truth. Every schema change needs: (1) SQL migration
`timestamp_name.up.sql`, (2) sqlc query defs, (3) repository update. Never modify
production schemas manually. Use sqlc for type-safe queries; write explicit SQL, no ORM.

## API response format
```json
{ "success": true, "data": {} }
{ "success": false, "error": { "code": "ERROR_CODE", "message": "Human readable" } }
```
Keep it consistent. Convert errors to HTTP responses in the handler layer.

## Performance
Avoid N+1 queries, unnecessary joins, unused relations, full table scans. Prefer
explicit selects. Don't optimize prematurely — measure first.

## Testing
Prioritize repository, service, and workflow tests. Prefer integration tests over
excessive mocking. Test observable behavior, not implementation details.

## Logging & observability
Structured logging (zerolog/zap) for uploads, downloads, deletions, storage/backup
failures. Avoid noisy logs. Keep observability simple for MVP (future: OpenTelemetry
+ Grafana LGTM) — don't add complex frameworks unless required.

## Forbidden patterns (unless explicitly requested)
Generic/base repositories, base services, repository/service factories, CQRS, event
sourcing, domain events, DI containers, plugin frameworks, microservices.

## Agent behaviour
Before changing code: understand the request → search existing code → follow
conventions → reuse existing implementations → minimize changes → avoid unrelated
refactors. When uncertain, choose the simplest implementation.

## Final rule
Filora is a product, not an architecture exercise. Every abstraction must justify
its existence. Working software beats perfect architecture.
