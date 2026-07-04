# Database Design Rules

These are hard constraints for Filora's database. Follow them for every schema change.

---

## 1. Principles

1. **Metadata first** — PostgreSQL is always the source of truth. Cloud storage providers are never authoritative.
2. **Provider agnostic** — assets are decoupled from where their bytes physically live.
3. **Simple MVP** — this is a private, family-scale app. Do not design for hypothetical scale.
4. **Correctness > Simplicity > Readability > DX > Performance > Extensibility.**
5. **Build first, refactor later** — do not add tables/columns for features that do not exist yet.

---

## 2. Source of truth & migrations

- The single source of truth for the schema is [`apps/api/internal/database/schema.sql`](../../apps/api/internal/database/schema.sql).
- **There are no migration tools.** No `golang-migrate`, no versioned migration files.
- The schema is applied **manually** (see [Applying the schema](#applying-the-schema)).
- Backward compatibility is **not** a concern at this stage; the schema may be rewritten freely.
- Any schema change must be made in `schema.sql` **and** reflected in these docs in the same change.

---

## 3. ID strategy

Choose the ID type by the table's role, not by habit.

| Use | ID type | Applies to |
|-----|---------|-----------|
| Small control / lookup tables | `BIGINT GENERATED ALWAYS AS IDENTITY` | `users`, `roles`, `permissions`, `clerk_webhook_events`, `galleries`, `albums`, `invitations`, `tags`, `storage_providers`, `archive_sync_jobs` |
| High-volume or externally-exposed rows | **UUID v7** via `uuid_generate_v7()` | `assets`, `storage_locations`, `cli_sessions` |
| Pure join tables | Composite primary key | `role_permissions`, `user_roles`, `gallery_members`, `album_members`, `album_assets`, `asset_tags` |

Rules:
- Never use UUID v4. If a UUID is needed, it must be **v7** (time-ordered) so it indexes well.
- `uuid_generate_v7()` is defined in `schema.sql` (portable, built on core `gen_random_uuid()`, no extensions). PostgreSQL has no native `uuidv7()` until PG 18.
- Do not expose incremental IDs in public/shareable URLs; those rows use UUID v7.

---

## 4. Types & columns

- Timestamps are always `TIMESTAMPTZ`, default `now()`.
- `updated_at` is maintained by the `set_updated_at()` trigger — add the trigger to any table with `updated_at`.
- Prefer `TEXT` over `VARCHAR(n)` unless a hard length limit is a real requirement.
- Use `JSONB` (never `JSON`) for semi-structured data (`credentials`, `metadata`).
- Use enums for small, stable value sets (e.g. `permission_scope`). Use `TEXT` + `CHECK` for sets that may grow (e.g. provider `type`, asset `type`).
- Money/size values are `BIGINT` in bytes.
- Booleans are `NOT NULL` with an explicit default.

---

## 5. Constraints & integrity

- Every foreign key declares an explicit `ON DELETE` action (`CASCADE` or `SET NULL`). No implicit `NO ACTION`.
- Enforce uniqueness in the database, not just in code (e.g. `users.clerk_user_id`, `users.email`, `permissions (resource, action)`, `assets (gallery_id, hash)`, `tags (gallery_id, name)`).
- Ownership columns reference `users(id)`: `owner_id` for owned resources (galleries, albums), `uploaded_by` for asset contributors.
- Deduplication of assets is per-gallery via the unique index `assets (gallery_id, hash)`.

---

## 6. Naming conventions

- Tables: `snake_case`, plural (`storage_providers`).
- Columns: `snake_case`, singular.
- Primary key: `id`.
- Foreign keys: `<referenced_table_singular>_id` (`asset_id`, `provider_id`, `role_id`).
- Indexes: `idx_<table>_<columns>`.
- Triggers: `trg_<table>_<purpose>`.
- Enums: singular, descriptive (`permission_scope`).
- Junction tables: `<a>_<b>` (`role_permissions`, `user_roles`).

---

## 7. Indexing

- Index every foreign key that is queried (most of them).
- Add partial indexes for common filtered queries (e.g. active CLI sessions: `WHERE revoked_at IS NULL`).
- Do not add speculative indexes. Add an index when a query needs it.

---

## 8. Auth boundaries (see [rbac.md](./rbac.md))

- Web login/session state lives in **Clerk**, not our DB. We only store `clerk_user_id` + profile mirror. **Never store passwords.**
- Terminal auth uses `cli_sessions`. Store only the **SHA-256 hash** of the token, never the raw token.
- Authorization is two-tier: global **RBAC** (`role → permission @ scope`) for
  capability, plus per-resource **membership** (`owner`/`editor`/`viewer`) on
  galleries and albums. Check both — never hard-code user-id comparisons.
- Storage accounts are **global** (admin-managed via the `storage` permission),
  not owned by end users.

## 8a. Domain invariants

- Every user has exactly one **default** gallery (`galleries.is_default`, enforced
  by a partial unique index). Create it on user provisioning.
- A resource `owner_id` must also appear in its membership table with role `owner`.
- Every asset should end up with **≥1 `serving` and ≥1 `archive`** storage location;
  archive copies are produced asynchronously via `archive_sync_jobs`.
- Dedup is per gallery and ignores trashed rows (`assets` partial unique on
  `(gallery_id, hash) WHERE deleted_at IS NULL`).
- Tags are per gallery (`tags` unique on `(gallery_id, name)`).
- Storage **quota is per gallery** (`galleries.storage_quota` / `storage_used`);
  physical capacity is spread across multiple accounts per layer.
- An `invitations` row targets **exactly one** of `gallery_id` / `album_id`
  (CHECK), with at most one pending invite per target+email.
- Clerk webhooks are idempotent: record each `event_id` in `clerk_webhook_events`
  and skip if already present.
- **Soft delete** applies to `assets` only. Queries for live assets must filter
  `WHERE deleted_at IS NULL`.

---

## 9. Forbidden patterns

Do not introduce, unless explicitly requested:

- ORMs for complex operations (use explicit SQL via sqlc).
- Generic/base repositories, repository factories.
- Soft-delete columns on every table "just in case".
- Polymorphic foreign keys / EAV / arbitrary key-value tables.
- Triggers containing business logic (triggers are only for `updated_at`).
- Storing provider state as the source of truth.
- UUID v4, or sequences exposed publicly.

---

## 10. Change checklist

Before finishing a schema change:

- [ ] `schema.sql` updated.
- [ ] ID type matches [section 3](#3-id-strategy).
- [ ] FKs have explicit `ON DELETE`.
- [ ] Needed indexes added; no speculative ones.
- [ ] `updated_at` trigger added if the table has `updated_at`.
- [ ] Docs updated: [schema.md](./schema.md), [erd.md](./erd.md), and this file if a rule changed.
- [ ] Seed data ([`seed.sql`](../../apps/api/internal/database/seed.sql)) updated if new roles/permissions are required.
- [ ] Downstream: sqlc queries regenerated and repositories updated (tracked separately from design).

---

## Applying the schema

No migration runner. Apply manually with `psql`:

```bash
cd apps/api
make db-apply   # psql "$DATABASE_URL" -f internal/database/schema.sql
make db-seed    # psql "$DATABASE_URL" -f internal/database/seed.sql
```

### Superuser bootstrap

Users are created from Clerk, so the owner row does not exist until they sign in
once. After first sign-in, grant the superuser role (replace the email):

```sql
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'owner@example.com' AND r.slug = 'superuser'
ON CONFLICT DO NOTHING;
```
