# Roadmap & Status

Filora is at the **MVP / design stage**, solo-developed. The goal is to ship a
working family DAM quickly, validate it, and keep the code maintainable.

---

## Where we are

- ✅ **Database design** is complete and is the source of truth
  ([schema.sql](../../apps/api/internal/database/schema.sql), [docs](../database/README.md)).
- 🟡 **API** has a **legacy implementation** (JWT + password auth, per-user
  assets, single-layer storage) that predates the current design.
- 🟡 **Web** app is scaffolded (React 19 + Vite + shadcn).
- 🔭 **CLI** not started.

## The legacy → target gap

The current API code and the new design differ in important ways. This is the
work ahead:

| Area | Legacy (current code) | Target (designed) |
|------|-----------------------|-------------------|
| Auth | JWT + bcrypt passwords | Clerk (web) + opaque CLI tokens |
| Authorization | Owner-only checks | RBAC (roles + scope) + membership roles |
| Users | Local accounts | Clerk mirror (webhook + JIT) |
| Organization | Flat, per-user assets | Galleries → albums, tags, invitations |
| Storage | Single provider copy, per-user accounts | Two layers (serving + archive), global accounts, multi-account |
| Quota | Per user | Per gallery |
| IDs | UUID v4 | bigint (control) + UUID v7 (assets) |
| Delete | Hard delete | Soft delete (trash) |
| Migrations | golang-migrate | Manual `schema.sql` (no migrate) |

## Suggested implementation order

Following the project's **database-first** rule (schema → queries → repository →
service → handler → UI). A detailed, phased build plan lives in
[architecture/implementation-plan.md](../architecture/implementation-plan.md).

1. **Foundations**
   - Apply new `schema.sql` + `seed.sql` to Neon.
   - Regenerate sqlc; rewrite queries for the new schema.
2. **Identity & access**
   - Clerk integration (webhook + JIT user sync); remove password/JWT code.
   - CLI token issuance & auth middleware.
   - RBAC + membership permission checks.
3. **Organization**
   - Galleries (with default-gallery provisioning), gallery membership.
   - Albums, album membership, album assets.
   - Tags + tagging.
   - Invitations (email invite + accept).
4. **Assets & storage**
   - Upload to the serving layer; per-gallery dedup; soft delete/trash.
   - Storage account management (global, admin) + usage view.
   - Archive layer adapter (GCS) + `archive_sync_jobs` worker.
5. **Clients**
   - Web app screens over the new API.
   - CLI (Cobra) with multi-session support.

## Backlog (deferred on purpose)

- **Account election** — strategy for choosing which account within a layer
  receives a new upload (round-robin, most-free-space, weighted).
- **Archive providers** — Alibaba OSS Archive as an alternative to GCS Archive.
- **Asset versioning / history.**
- **Disaster-recovery** orchestration and scheduled backups beyond the archive copy.
- **Nice-to-haves** — favorites, public share links, notifications.
- **Observability** — OpenTelemetry + Grafana LGTM (kept out of MVP).

## Guiding constraints

- Don't optimize for hypothetical future scale.
- No new abstractions until a second concrete case exists.
- Keep operational complexity low (single API, Neon Postgres, free-tier storage).

See [`/AGENTS.md`](../../AGENTS.md) for the full engineering rules.
