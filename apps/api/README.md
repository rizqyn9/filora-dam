# Filora DAM API

The Go REST API server — the source of all business logic for Filora. The web
and CLI clients are thin layers over this API.

> **Status.** Phases 0–9 of the rebuild are implemented: identity (Clerk),
> RBAC, CLI sessions, galleries, albums, tags, storage accounts, assets, the
> archive worker, and the dashboard. See the
> [implementation plan](../../docs/architecture/implementation-plan.md) and
> [API reference](API.md).
>
> **Remaining:** concrete storage-provider SDKs (Cloudinary/ImageKit/R2/GCS) are
> stubbed (`ErrNotImplemented`), so uploads/archive replication are non-functional
> until an adapter + credentials are wired. See [storage adapters](internal/modules/storage/adapters).

## Tech Stack

- **Go** 1.23+, **Fiber v3** (web framework)
- **PostgreSQL** (Neon) via **pgx/v5**
- **sqlc** — type-safe SQL codegen
- **validator v10** — input validation
- **Clerk** — web auth; opaque CLI tokens for the terminal
- **zerolog** — structured logging

## Prerequisites

- Go 1.23+
- A PostgreSQL database (Neon recommended)
- `psql` (to apply the schema) and `sqlc` (to regenerate query code)

## Quick Start

```bash
cd apps/api
cp .env.example .env         # set DATABASE_URL (+ Clerk keys for target)
go mod download

make db-apply                # apply internal/database/schema.sql
make db-seed                 # seed baseline roles/permissions
make run                     # start the server
```

Health check: `curl http://localhost:$PORT/health`

> **No migrations.** The schema is a single canonical file applied manually with
> `psql`; there is no `migrate` tool. See
> [database rules](../../docs/database/rules.md#applying-the-schema).

## Commands

```bash
make run             # Run the server
make build           # Build the binary
make test            # Run tests
make test-coverage   # Tests with HTML coverage
make fmt             # Format code (gofmt)
make lint            # Run golangci-lint
make db-apply        # Apply schema.sql to $DATABASE_URL
make db-seed         # Seed baseline RBAC roles/permissions
make sqlc            # Regenerate sqlc code from queries
make deps            # go mod download && tidy
make clean           # Remove build artifacts
```

## Project Structure

```
apps/api/
├── cmd/
│   ├── server/main.go           # HTTP API (compose root)
│   └── worker/main.go           # archive replication worker
├── internal/
│   ├── config/                  # env load + validation
│   ├── database/                # pgx pool, schema.sql, seed.sql, queries/, db/ (sqlc)
│   ├── server/ · middleware/ · auth/ · lib/
│   └── modules/{account,session,rbac,gallery,album,tag,asset,storage,dashboard}/
├── API.md · TESTING.md
├── sqlc.yaml · Makefile · go.mod · go.sum · .env.example
```

### Module structure (vertical slice)

Each feature module owns its full stack:

```
internal/modules/<module>/
├── handler.go       # HTTP routes: validate, call service, format response
├── service.go       # Business logic, orchestration, permission checks
├── repository.go    # Database access (sqlc)
└── models.go        # Data structures
```

### Adding a feature (database-first)

1. Update `internal/database/schema.sql`.
2. Write/adjust queries in `internal/database/queries/*.sql`.
3. `make sqlc` to regenerate typed code.
4. Implement repository → service → handler.
5. Add tests.

## Environment Variables

See [`.env.example`](.env.example):

```
PORT=3000
ENV=development
DATABASE_URL=postgres://user:pass@host/filora
CLERK_SECRET_KEY=...
CLERK_WEBHOOK_SIGNING_SECRET=...
CLI_TOKEN_TTL_HOURS=720
```

Storage provider credentials are stored **per account in the database**
(`storage_providers.credentials`), not in env.

## Documentation

- [Docs index](../../docs/README.md)
- [Architecture](../../docs/architecture/README.md) · [Auth](../../docs/architecture/auth.md) · [Storage](../../docs/architecture/storage.md)
- [Database](../../docs/database/README.md)
- [API reference](API.md) · [Testing](TESTING.md)
- [Engineering rules](../../AGENTS.md)

## License

Proprietary — Filora Digital Asset Management Platform.
