# Filora Documentation

Central entry point for Filora's documentation.

> **Filora** is a private, family-scale, multi-cloud Digital Asset Management (DAM)
> platform. Users upload and organize photos, videos, and documents; where the
> bytes actually live (which provider / account) is fully abstracted away.

---

## How to use these docs

**For humans** — start with [Product](./product/README.md) to understand *what*
Filora is, then [Architecture](./architecture/README.md) for *how* it fits
together, then [Database](./database/README.md) for the data model.

**For AI agents** — read in this order before making changes:
1. [`/AGENTS.md`](../AGENTS.md) — operating rules & conventions (authoritative).
2. [product/](./product/README.md) — domain language & scope.
3. [architecture/](./architecture/README.md) — system boundaries & flows.
4. [database/](./database/README.md) — schema is the source of truth for data.

If docs and code/SQL disagree, the **code/SQL wins** and the docs must be fixed.

---

## Map

| Area | Start here | Covers |
|------|-----------|--------|
| Product | [product/README.md](./product/README.md) | Vision, users, domain concepts, features, roles, roadmap |
| Architecture | [architecture/README.md](./architecture/README.md) | Apps, layering, request/upload flows, storage layers, auth |
| Database | [database/README.md](./database/README.md) | Schema reference, ERD, design rules, RBAC model |

### Per-app docs

| App | Docs |
|-----|------|
| API (Go) | [`apps/api/README.md`](../apps/api/README.md), [`apps/api/API.md`](../apps/api/API.md) |
| Web (React) | [`apps/web/README.md`](../apps/web/README.md) |
| CLI (Go) | _planned_ |

---

## Implementation status

Filora is at the **MVP / design stage**. The documentation here describes the
**target design** (Clerk auth, RBAC, galleries/albums, two-layer storage).

The current API code still reflects an **earlier implementation** (JWT + password
auth, per-user assets). Documents that describe that legacy behavior are marked
with a status banner. See [product/roadmap.md](./product/roadmap.md) for the gap
and migration plan.
