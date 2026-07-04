# Roadmap & Status

Filora is at the **MVP / design stage**, solo-developed. The goal is to ship a
working family DAM quickly, validate it, and keep the code maintainable.

---

## Where we are

- ✅ **Database design** is complete and is the source of truth
  ([schema.sql](../../apps/api/internal/database/schema.sql), [docs](../database/README.md)).
- ✅ **API rebuild (phases 0–9)** implemented on the target design: Clerk auth,
  RBAC, CLI sessions, galleries, albums, tags, storage accounts, assets, the
  archive worker, and the dashboard. See
  [implementation-plan.md](../architecture/implementation-plan.md) and
  [apps/api/API.md](../../apps/api/API.md).
- 🟡 **Storage adapters** — the **R2 (S3-compatible) adapter is implemented** and
  works for both layers; `cloudinary`/`imagekit`/`gcs` remain stubbed. With an
  active `r2` storage account configured, uploads and archive replication work
  end-to-end.
- 🟡 **Web** app is scaffolded (React 19 + Vite + shadcn).
- 🔭 **CLI** not started.

## Remaining before end-to-end

1. Configure an active `r2` serving account (add `public_base_url` for servable
   URLs); optionally an `r2`/`gcs` archive account. Additional provider adapters
   (Cloudinary/ImageKit/GCS) as needed.
2. Configure Clerk (keys + webhook) and grant the owner the `superuser` role.
3. Broaden tests (repository/service/workflow against a test database).
4. Build the web app and CLI against the API.

## Legacy → target (done)

The legacy implementation was removed and rebuilt on the target design. For
reference, the shifts that were made:

| Area | Legacy (removed) | Now |
|------|------------------|-----|
| Auth | JWT + bcrypt passwords | Clerk (web) + opaque CLI tokens |
| Authorization | Owner-only checks | RBAC (roles + scope) + membership roles |
| Users | Local accounts | Clerk mirror (webhook + JIT) |
| Organization | Flat, per-user assets | Galleries → albums, tags, invitations |
| Storage | Single provider copy, per-user accounts | Two layers (serving + archive), global accounts, multi-account |
| Quota | Per user | Per gallery |
| IDs | UUID v4 | bigint (control) + UUID v7 (assets) |
| Delete | Hard delete | Soft delete (trash) |
| Migrations | golang-migrate | Manual `schema.sql` (no migrate) |

## Implementation order (reference)

The rebuild followed the project's **database-first** rule (schema → queries →
repository →
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
