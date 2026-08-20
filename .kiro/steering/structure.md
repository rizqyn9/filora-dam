---
inclusion: always
---

# Project Structure

Monorepo with three independent apps. No shared packages.

```
filora-dam/
├── api/                          # Go backend (all business logic)
│   ├── cmd/
│   │   ├── server/main.go        # HTTP server entry + compose root
│   │   └── worker/main.go        # Archive replication worker
│   ├── internal/
│   │   ├── auth/                  # Auth middleware, token validation
│   │   ├── clerk/                 # Clerk webhook + user sync
│   │   ├── config/                # Env-based configuration
│   │   ├── database/
│   │   │   ├── db/                # sqlc-generated code
│   │   │   ├── queries/           # .sql query definitions
│   │   │   └── schema.sql         # Canonical schema (source of truth)
│   │   ├── lib/                   # Shared utilities
│   │   ├── modules/               # Vertical slices (one per domain)
│   │   │   ├── account/
│   │   │   ├── asset/
│   │   │   ├── dashboard/
│   │   │   ├── folder/
│   │   │   ├── rbac/
│   │   │   ├── session/
│   │   │   ├── space/
│   │   │   ├── storage/
│   │   │   └── tag/
│   │   └── server/
│   │       └── middleware/
│   └── migrations/                # golang-migrate .up.sql/.down.sql
├── cli/                           # Go CLI client (thin HTTP wrapper)
│   ├── cmd/
│   └── internal/
│       ├── client/
│       └── commands/
├── web/                           # React frontend (thin UI client)
│   └── src/
│       ├── components/            # Shared UI components
│       │   ├── layout/
│       │   └── ui/                # Shadcn primitives
│       └── features/              # Feature-scoped code
│           ├── account/
│           ├── assets/
│           ├── dashboard/
│           ├── folders/
│           └── roles/
├── docs/
│   ├── product/                   # What Filora is
│   ├── architecture/              # How it fits together
│   ├── database/                  # Schema, ERD, design standards
│   └── adr/                       # Architecture Decision Records
├── .kiro/
│   ├── steering/                  # Kiro persistent context
│   └── hooks/                     # Event-driven automation
├── AGENTS.md                      # Agent operating rules (all AI tools)
└── CONTEXT.md                     # Domain glossary (authoritative vocabulary)
```

## Module Anatomy (API vertical slice)

Each module in `api/internal/modules/<name>/` owns:

| File | Responsibility |
|------|---------------|
| `handler.go` | HTTP: validate request, call service, return response |
| `service.go` | Business logic, orchestration, authz checks |
| `repository.go` | DB access via sqlc-generated code |
| `models.go` | DTOs, request/response structs |
| `routes.go` | Route registration (optional) |

No cross-module imports. Cross-module needs use interfaces injected at `cmd/server/main.go`.

## Key Files

| Purpose | Path |
|---------|------|
| Canonical database schema | `api/internal/database/schema.sql` |
| sqlc config | `api/sqlc.yaml` |
| API entry point | `api/cmd/server/main.go` |
| Domain glossary | `CONTEXT.md` |
| Agent rules | `AGENTS.md` |
