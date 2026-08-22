---
inclusion: always
---

# Technology & Structure

Three independent apps at workspace root. No shared packages.

## Stack

### API (`api/`)

| Component | Technology |
|-----------|-----------|
| Language | Go 1.23+ |
| HTTP framework | Fiber v3 |
| Database | PostgreSQL (Neon) |
| Query generation | sqlc |
| Migrations | golang-migrate |
| Validation | go-playground/validator v10 |
| Auth | Opaque tokens + in-process LRU cache |
| Logging | zerolog (startup), slog (application via OTel) |
| Linting | golangci-lint |

### Web (`web/`)

| Component | Technology |
|-----------|-----------|
| Language | TypeScript |
| Framework | React 19 |
| Bundler | Vite |
| Router | TanStack Router (file-based) |
| Data fetching | TanStack Query |
| Styling | Tailwind v4 |
| Components | Shadcn UI (base-nova, base-ui backed) |
| Validation | Zod v4 |
| Package manager | bun |
| Linting | oxlint |

### CLI (`cli/`)

| Component | Technology |
|-----------|-----------|
| Language | Go |
| Role | Thin HTTP client, no business logic |

### Storage Providers (adapter-based)

| Provider | Layer | Status |
|----------|-------|--------|
| Cloudflare R2 | L1/L2 | Implemented |
| Cloudinary | L1 (serving) | Implemented |
| ImageKit | L1 (serving) | Implemented |
| GCS Archive | L2 (archive) | Planned |

All adapters implement a `StorageAdapter` interface. Business logic never touches provider SDKs directly.

## Project Layout

```
filora-dam/
├── api/
│   ├── cmd/server/main.go          # Compose root
│   ├── cmd/worker/main.go          # Archive replication worker (unused, inline in storage module)
│   ├── internal/
│   │   ├── auth/                    # Middleware, token cache, context helpers
│   │   ├── config/                  # Env-based configuration
│   │   ├── database/
│   │   │   ├── db/                  # sqlc-generated code
│   │   │   ├── queries/             # .sql query definitions
│   │   │   └── schema.sql           # Canonical schema (source of truth)
│   │   ├── lib/                     # Shared utilities (response, validate, crypto, token)
│   │   ├── modules/                 # Vertical slices
│   │   │   ├── asset/
│   │   │   ├── auth/                # Login, register, logout, change-password
│   │   │   ├── folder/
│   │   │   ├── space/
│   │   │   ├── storage/             # Accounts, registry, uploader, archive worker
│   │   │   └── tag/
│   │   ├── server/                  # Fiber app setup
│   │   └── telemetry/               # OTel init, metrics, middleware
│   └── migrations/                  # golang-migrate .up.sql/.down.sql
├── web/
│   └── src/
│       ├── components/
│       │   ├── assets/              # Asset browser, grid, list, upload, dialogs
│       │   ├── folders/             # Folder mutation dialogs
│       │   ├── layout/              # Sidebar, space-switcher, folder-tree
│       │   ├── spaces/              # Create space dialog
│       │   └── ui/                  # Shadcn primitives
│       ├── features/
│       │   ├── assets/              # api.ts, schemas.ts, use-upload-files.ts
│       │   ├── auth/                # api.ts, schemas.ts, auth-store.ts
│       │   ├── folders/             # api.ts, schemas.ts
│       │   └── spaces/              # api.ts, schemas.ts
│       ├── hooks/                   # use-theme.ts
│       ├── lib/                     # api.ts, query-client.ts, utils.ts, format.ts
│       ├── routes/                  # TanStack Router file-based routes
│       └── stores/                  # ui-store.ts
├── cli/
│   └── internal/
│       ├── client/
│       └── commands/
├── docs/
│   ├── product/
│   ├── architecture/
│   ├── database/
│   └── adr/
├── AGENTS.md
└── CONTEXT.md
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

No cross-module imports. Cross-module needs use consumer-defined interfaces injected at `cmd/server/main.go`.

## Design Order (database-first)

SQL migration → sqlc queries → repository → service → handler → UI.
