# Filora DAM API

The Go REST API server — the source of all business logic for Filora. The web
and CLI clients are thin layers over this API.

> **Status — clean-slate rebuild.** The Go code has been fully **reset**; only
> the finalized database design (`schema.sql`/`seed.sql`) and tooling remain.
> The whole app is being rebuilt for the **target design** (Clerk auth, RBAC,
> galleries/albums, two-layer storage) following the
> [implementation plan](../../docs/architecture/implementation-plan.md).
>
> Note: the API does not build yet. The legacy [`API.md`](API.md) /
> [`TESTING.md`](TESTING.md) / [`TESTING_MANUAL.md`](TESTING_MANUAL.md) are kept
> (banner-marked) for reference and will be rewritten during the rebuild.

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

The Go code has been **reset to a clean slate**. Only the database design and
tooling remain; everything is rebuilt following
[project-structure.md](../../docs/architecture/project-structure.md) per the
[implementation plan](../../docs/architecture/implementation-plan.md).

Present:

```
apps/api/
├── internal/database/
│   ├── schema.sql               # Canonical schema (source of truth, manual apply)
│   └── seed.sql                 # Baseline RBAC roles/permissions
├── API.md · TESTING*.md         # legacy reference (banner-marked)
├── sqlc.yaml · Makefile · go.mod · go.sum · .env.example
```

Target layout to build (see
[project-structure.md](../../docs/architecture/project-structure.md)):

```
apps/api/
├── cmd/server/main.go · cmd/worker/main.go
└── internal/
    ├── config/ · database/{db.go,queries/,db/} · server/ · middleware/ · auth/ · lib/
    └── modules/{account,session,rbac,gallery,album,tag,asset,storage,dashboard}/
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

_Planned modules (target):_ `account` (Clerk users), `rbac` (roles/permissions),
`session` (CLI tokens), `gallery`, `album`, `tag`, `asset`, `storage`
(providers + adapters + two-layer orchestration), `dashboard`.

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
