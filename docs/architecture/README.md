# Architecture Documentation

How Filora is built and how the pieces fit together. This describes the
**target design**; see [product/roadmap.md](../product/roadmap.md) for the gap
against the current legacy API code.

| Doc | Covers |
|-----|--------|
| [overview.md](./overview.md) | The three apps, layering, request/response flow, key principles |
| [project-structure.md](./project-structure.md) | `apps/api` Go layout, module anatomy, dependency rules, wiring |
| [implementation-plan.md](./implementation-plan.md) | Phased, database-first plan to (re)build `apps/api` |
| [storage.md](./storage.md) | Two-layer storage model, adapters, upload & archive flows |
| [auth.md](./auth.md) | Clerk (web), CLI tokens, and how RBAC + membership are enforced |

Related: [Product](../product/README.md) · [Database](../database/README.md) ·
[`/AGENTS.md`](../../AGENTS.md) (engineering rules)
