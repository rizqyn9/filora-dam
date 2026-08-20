---
inclusion: always
---

# Technology Stack

Three independent apps at workspace root. No shared packages.

## API (`api/`)

| Component | Technology |
|-----------|-----------|
| Language | Go 1.23+ |
| HTTP framework | Fiber v3 |
| Database | PostgreSQL (Neon) |
| Query generation | sqlc |
| Migrations | golang-migrate |
| Validation | go-playground/validator v10 |
| Auth (web + CLI) | Opaque tokens + in-process LRU cache |
| Logging | zerolog |
| Linting | golangci-lint |

Architecture: modular vertical-slice. Each module in `internal/modules/<name>/` owns handler → service → repository → models. Cross-module via consumer-defined interfaces injected at compose root (`cmd/server/main.go`).

## Web (`web/`)

| Component | Technology |
|-----------|-----------|
| Language | TypeScript |
| Framework | React 19 |
| Bundler | Vite |
| Router | TanStack Router |
| Data fetching | TanStack Query |
| Styling | Tailwind v4 |
| Components | Shadcn UI |
| Validation | Zod v4 |
| Package manager | bun |
| Linting | oxlint |
| Formatting | oxfmt (prettier fallback) |

## CLI (`cli/`)

| Component | Technology |
|-----------|-----------|
| Language | Go |
| Role | Thin HTTP client, no business logic |

## Storage Providers (adapter-based)

| Provider | Layer | Status |
|----------|-------|--------|
| Cloudinary | L1 (serving) | Planned |
| ImageKit | L1 (serving) | Planned |
| Cloudflare R2 | L1/L2 | Implemented |
| GCS Archive | L2 (archive) | Planned |

All adapters implement a `StorageAdapter` interface. Business logic never touches provider SDKs directly.

## Infrastructure

| Service | Provider |
|---------|----------|
| Database | Neon (PostgreSQL) |
| Auth/Identity | Self-managed (bcrypt + opaque tokens) |
| Storage | Multi-provider pool (see above) |
| Hosting | TBD |

## Design Order (database-first)

SQL migration → sqlc queries → repository → service → handler → UI.
