# Filora DAM API

The Go REST API server — the source of all business logic for Filora. The web
and CLI clients are thin layers over this API.

> **Status.** The **target design** (Clerk auth, RBAC, galleries/albums,
> two-layer storage) is documented in [`/docs`](../../docs/README.md) and the
> database is finalized in [`internal/database/schema.sql`](internal/database/schema.sql).
> The **code in this app is still the earlier (legacy) implementation** (JWT +
> password auth, per-user assets, single-layer storage) and is being migrated.
> Where this README describes current code, it is marked _legacy_. See the
> [roadmap](../../docs/product/roadmap.md) for the gap and plan.

## Tech Stack

- **Go** 1.23+, **Fiber v3** (web framework)
- **PostgreSQL** (Neon) via **pgx/v5**
- **sqlc** — type-safe SQL codegen
- **validator v10** — input validation
- **Clerk** — web auth _(target; legacy code uses JWT/bcrypt)_

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
├── cmd/server/main.go            # Entry point
├── internal/
│   ├── config/                   # Environment configuration (validator v10)
│   ├── database/
│   │   ├── schema.sql            # Canonical schema (source of truth, manual apply)
│   │   ├── seed.sql              # Baseline RBAC roles/permissions
│   │   ├── db.go                 # pgx connection pool
│   │   ├── db/                   # sqlc-generated code
│   │   └── queries/              # SQL queries for sqlc
│   ├── lib/                      # Shared helpers (response, hashing, mime, …)
│   ├── middleware/               # Auth middleware
│   └── modules/                  # Feature modules (vertical slice)
├── API.md                        # API reference (legacy — see banner)
├── sqlc.yaml
├── Makefile
├── go.mod / go.sum
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

_Legacy modules present today:_ `account`, `asset`, `storage` (+ `adapters/`),
`dashboard`.
_Planned (target):_ add `rbac`, `session`, `gallery`, `album`, `tag`; rework
`account` for Clerk and `storage` for two layers. See
[architecture overview](../../docs/architecture/overview.md).

### Adding a feature (database-first)

1. Update `internal/database/schema.sql`.
2. Write/adjust queries in `internal/database/queries/*.sql`.
3. `make sqlc` to regenerate typed code.
4. Implement repository → service → handler.
5. Add tests.

## Environment Variables

Target (see [`.env.example`](.env.example)):

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

> _Legacy note:_ the current code still expects a `JWT_SECRET` and per-provider
> env vars until the Clerk + DB-managed-account migration lands.

## Documentation

- [Docs index](../../docs/README.md)
- [Architecture](../../docs/architecture/README.md) · [Auth](../../docs/architecture/auth.md) · [Storage](../../docs/architecture/storage.md)
- [Database](../../docs/database/README.md)
- [API reference](API.md) _(legacy)_
- [Engineering rules](../../AGENTS.md)

## License

Proprietary — Filora Digital Asset Management Platform.
