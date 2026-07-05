---
description: Type-check & vet both apps (Go API + React web)
agent: build
---

Run static checks across the monorepo and fix any errors you find. Work through
them until all pass; do not stop at the first failure.

API (Go) — run in `apps/api`:
- `go build ./...`
- `go vet ./...`

Web (React/TS) — run in `apps/web`:
- `bun run typecheck`
- `bun run lint`

Report a short summary per app (pass/fail + what you fixed). If `$ARGUMENTS` is
given, scope the checks to that app only (`api` or `web`).
