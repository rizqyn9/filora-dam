# Filora DAM

Private, family-scale, multi-cloud Digital Asset Management (DAM) platform.

Upload and organize photos, videos, and documents in **galleries** and **albums**.
Filora spreads files across multiple free-tier cloud accounts and keeps a cheap
**archive copy** — while hiding all of that storage complexity from the user.

> New here? Start with the [documentation index](docs/README.md) →
> [Product overview](docs/product/overview.md).

## Project Structure

```
filora-dam/
├── apps/
│   ├── api/          # Go REST API server (all business logic)
│   ├── cli/          # Go CLI client (planned)
│   └── web/          # React 19 frontend
├── docs/             # Documentation (product, architecture, database)
├── AGENTS.md         # Engineering rules for all contributors/agents
└── CLAUDE.md         # Points AI agents at AGENTS.md
```

## Applications

| App | Stack | Status |
|-----|-------|--------|
| **API** ([README](apps/api/README.md)) | Go, Fiber v3, sqlc, PostgreSQL (Neon) | Phases 0–10 implemented; R2 adapter live |
| **Web** ([README](apps/web/README.md)) | React 19, TypeScript, TanStack Query/Router, Tailwind v4, shadcn/ui, Zod | Scaffolded |
| **CLI** ([README](apps/cli/README.md)) | Go, Cobra | Basic client (login, sessions, galleries, upload/list/download) |

## Features (target design)

- Multi-cloud storage pooled across many accounts (Cloudinary, ImageKit, R2, GCS)
- **Two storage layers**: serving (hot) + archive (cheap backup) — every asset in both
- Galleries, albums, and normalized tagging
- Sharing via email **invitations** with `owner`/`editor`/`viewer` roles
- **RBAC** (superuser/admin/member/viewer) with Clerk-based web auth
- Terminal (CLI) login with **multiple concurrent sessions**
- Per-gallery quota, per-gallery deduplication, soft-delete (trash)

See the [feature catalog](docs/product/features.md) for status of each item.

## Documentation

- [Documentation index](docs/README.md) — start here
- [Product](docs/product/README.md) — overview, concepts, features, roles, roadmap
- [Architecture](docs/architecture/README.md) — apps, flows, storage, auth
- [Database](docs/database/README.md) — schema, ERD, rules, RBAC
- [AGENTS.md](AGENTS.md) — engineering rules for all contributors

## Getting Started

### Prerequisites

- Go 1.23+
- A PostgreSQL database (Neon recommended)
- Node.js 20+ / Bun (for the web app)
- A [Clerk](https://clerk.com) application (web auth)

### Quick start (API + database)

```bash
git clone https://github.com/rizqyn9/filora-dam.git
cd filora-dam/apps/api

cp .env.example .env          # set DATABASE_URL + Clerk keys
make db-apply                 # apply schema.sql to your database (no migrations)
make db-seed                  # seed baseline roles/permissions
make run                      # start the API
```

> Filora uses a **manual schema** (`internal/database/schema.sql`) applied with
> `psql` — there is no migration tool. See [database rules](docs/database/rules.md#applying-the-schema).

## Development philosophy

**Core principles**

1. **Build first** — working software over perfect architecture.
2. **Refactor later** — no abstractions until 2+ concrete cases exist.
3. **Database first** — schema → queries → repository → service → handler → UI.
4. **Consistency first** — follow existing patterns.

**Tech philosophy**

- **API first** — all business logic in the API.
- **Thin clients** — CLI and web only handle presentation.
- **Storage abstraction** — providers are implementation details.
- **Database as truth** — PostgreSQL is the source of truth.

All contributors (human and AI) must follow [AGENTS.md](AGENTS.md).

## Status

MVP / design stage. The database design is complete; the API is being migrated
from its legacy implementation to the target design. See the
[roadmap](docs/product/roadmap.md).

## License

Private project — all rights reserved.
