# Testing

How Filora's API is tested and how to exercise it locally.

## Automated tests

Run all tests:

```bash
make test          # go test -v ./...
make test-coverage # with HTML coverage report
```

Current coverage focuses on pure, dependency-free logic (fast, no database):

- `internal/auth` — RBAC `Decide` (wildcards, scope precedence)
- `internal/lib` — MIME classification
- `internal/clerk` — Svix webhook signature verify + payload parse
- `internal/modules/session` — CLI token generation/prefix
- `internal/modules/gallery`, `.../asset` — role ranking, storage key building
- `internal/modules/storage/adapters` — credential validation + factory
- `internal/server` — health/root endpoints via `app.Test`

### Testing philosophy

Prefer integration-style tests around repositories, services, and workflows over
heavy mocking (see [`/AGENTS.md`](../../AGENTS.md)). Repository/service tests that
need Postgres should run against a disposable database (e.g. a local/Neon branch
via `DATABASE_URL`); these are added as the surface stabilizes.

## Manual testing

Prerequisites: a Postgres database and schema applied.

```bash
cp .env.example .env         # set DATABASE_URL (+ Clerk keys)
make db-apply && make db-seed
make run                     # server
make run-worker              # archive worker (separate terminal)
```

Health check:

```bash
curl -s http://localhost:$PORT/health | jq
```

Authenticated requests need a bearer token (Clerk session token, or a Filora CLI
token from `POST /api/v1/cli/sessions`). See [API.md](API.md) for the full
endpoint reference.

> Uploads require an active serving storage account **and** a concrete storage
> adapter implementation (currently stubbed). Until an adapter + credentials are
> wired, `POST /galleries/:id/assets` returns an error.
