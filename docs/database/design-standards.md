# Database Design Standards

Conventions and rules governing Filora's PostgreSQL schema design, naming, types, and evolution.

---

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Tables | `snake_case`, plural nouns | `assets`, `storage_providers` |
| Columns | `snake_case` | `uploaded_by`, `clerk_user_id` |
| Join tables | `parent_child` (alphabetical or logical order) | `album_assets`, `asset_tags` |
| Indexes | `idx_{table}_{column(s)}` | `idx_assets_gallery_id` |
| Unique indexes | `idx_{table}_{purpose}` | `idx_assets_gallery_hash` |
| Triggers | `trg_{table}_{purpose}` | `trg_assets_updated_at` |
| Functions | `snake_case`, verb-first | `set_updated_at()`, `uuid_generate_v7()` |
| Enums | `snake_case`, singular concept | `permission_scope`, `member_role` |
| Enum values | `snake_case`, short | `'pending'`, `'stored'`, `'own'` |
| Foreign keys | `{referenced_table_singular}_id` | `gallery_id`, `user_id` |
| Boolean columns | `is_` or `has_` prefix | `is_active`, `is_default`, `is_system` |

### Naming Rules

- No abbreviations unless universally understood (`id`, `url`, `ip`).
- FK columns name the referenced table in singular form + `_id`.
- Avoid redundant prefixes (column `name` not `gallery_name` inside `galleries`).
- Audit columns use `_by` suffix: `uploaded_by`, `invited_by`, `granted_by`, `deleted_by`.

---

## ID Strategy

Filora uses two ID types based on table purpose:

| Category | Type | Generation | Used For |
|----------|------|------------|----------|
| Lookup/control tables | `BIGINT GENERATED ALWAYS AS IDENTITY` | Sequential | `users`, `roles`, `permissions`, `galleries`, `albums`, `tags`, `storage_providers` |
| High-volume/external rows | `UUID` (v7, time-ordered) | `uuid_generate_v7()` | `assets`, `storage_locations`, `cli_sessions` |
| Join tables | Composite PK | N/A | `user_roles`, `role_permissions`, `gallery_members`, `album_members`, `album_assets`, `asset_tags` |

### Why UUID v7?

- Time-ordered: preserves index locality (B-tree friendly, no random scatter).
- Globally unique: safe for external exposure (URLs, API responses) without leaking row counts.
- No coordination: can be generated anywhere without sequence contention.

### When to Use Which

- Use **BIGINT** for tables that are admin-managed, low-volume, or never exposed in URLs.
- Use **UUID v7** for tables where IDs appear in API responses, are high-write, or need external uniqueness.
- Use **composite PK** for pure relationship/join tables with no independent identity.

---

## Column Types

| Data | Type | Notes |
|------|------|-------|
| Text (variable) | `TEXT` | No `VARCHAR(n)` — validate length in app layer |
| Integers | `BIGINT` or `INTEGER` | `BIGINT` for IDs and byte sizes; `INTEGER` for counters |
| Booleans | `BOOLEAN` | Always `NOT NULL` with a `DEFAULT` |
| Timestamps | `TIMESTAMPTZ` | Always timezone-aware; `NOT NULL DEFAULT now()` for creation |
| JSON/structured | `JSONB` | For metadata, credentials, webhook payloads |
| IP addresses | `INET` | Native PostgreSQL type |
| Enums | Custom `CREATE TYPE ... AS ENUM` | Defined at schema top |
| Money/quota | `BIGINT` (bytes) | Store raw bytes, format in app layer |

### Column Ordering (within CREATE TABLE)

1. Primary key
2. Foreign keys (most important relationship first)
3. Core business columns
4. Metadata/optional columns (JSONB, nullable)
5. Audit columns (`deleted_at`, `deleted_by`)
6. Timestamps (`created_at`, `updated_at`)

---

## Timestamps

Every table gets:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

The `updated_at` column is maintained by a trigger:

```sql
CREATE TRIGGER trg_{table}_updated_at
    BEFORE UPDATE ON {table}
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

Exceptions: join tables with no `updated_at` (they are insert/delete only, never updated).

---

## Soft Delete

Only `assets` uses soft delete (`deleted_at` + `deleted_by`). All other tables use hard delete with `ON DELETE CASCADE` or `ON DELETE SET NULL`.

### Soft Delete Rules

- Partial indexes filter `WHERE deleted_at IS NULL` for active-row queries.
- Unique constraints respect soft delete: `UNIQUE(gallery_id, hash) WHERE deleted_at IS NULL` allows re-upload after trash.
- Queries must explicitly filter `deleted_at IS NULL` unless listing trash.

---

## Foreign Key Behavior

| Scenario | ON DELETE | Example |
|----------|-----------|---------|
| Parent owns children (cascade cleanup) | `CASCADE` | `galleries → assets`, `users → cli_sessions` |
| Reference for audit/display (allow orphan) | `SET NULL` | `assets.uploaded_by`, `albums.cover_asset_id` |
| Prevent deletion of referenced row | `RESTRICT` | `storage_locations.provider_id` (can't delete active provider) |

### Decision Guide

- Use `CASCADE` when children are meaningless without the parent.
- Use `SET NULL` for audit/attribution columns where the reference is informational.
- Use `RESTRICT` when deletion would cause data integrity issues (active storage accounts).

---

## Indexing Strategy

### Always Index

- Every foreign key column (PostgreSQL does NOT auto-index FK targets).
- Columns used in `WHERE` clauses of common queries.
- Columns used in `ORDER BY` on large tables.

### Partial Indexes

Use `WHERE` on indexes to reduce size and speed up filtered queries:

```sql
-- Only index active (non-trashed) assets
CREATE INDEX idx_assets_gallery_active ON assets (gallery_id) WHERE deleted_at IS NULL;

-- Only index active CLI sessions
CREATE INDEX idx_cli_sessions_active ON cli_sessions (user_id) WHERE revoked_at IS NULL;

-- Worker queue: only runnable jobs
CREATE INDEX idx_archive_sync_jobs_queue ON archive_sync_jobs (next_retry_at)
    WHERE status IN ('pending', 'failed');
```

### Unique Constraints with Conditions

Partial unique indexes enforce business rules that only apply to a subset of rows:

```sql
-- One default gallery per user
CREATE UNIQUE INDEX idx_galleries_one_default ON galleries (owner_id) WHERE is_default;

-- One pending invite per (target, email)
CREATE UNIQUE INDEX idx_invitations_pending_gallery ON invitations (gallery_id, email)
    WHERE status = 'pending' AND gallery_id IS NOT NULL;
```

### What NOT to Index

- Low-cardinality columns alone (e.g., `BOOLEAN`). Combine with other columns or use partial indexes.
- Columns only used in rarely-run admin queries.
- Tables with < 1000 rows (sequential scan is faster than index lookup).

---

## Enum Design

### When to Use Enums

- Fixed, small set of values (< 10) that rarely changes.
- Values have semantic meaning in the database layer (used in `WHERE`, `CHECK`, partial indexes).

### When NOT to Use Enums

- Values change frequently (use a lookup table instead).
- More than ~10 values or values are user-defined.

### Current Enums

| Enum | Values | Used By |
|------|--------|---------|
| `permission_scope` | `own`, `all` | `role_permissions.scope` |
| `member_role` | `owner`, `editor`, `viewer` | `gallery_members.role`, `album_members.role`, `invitations.role` |
| `storage_layer` | `serving`, `archive` | `storage_providers.layer`, `storage_locations.layer` |
| `location_status` | `pending`, `stored`, `failed` | `storage_locations.status` |
| `invitation_status` | `pending`, `accepted`, `revoked`, `expired` | `invitations.status` |
| `job_status` | `pending`, `running`, `completed`, `failed` | `archive_sync_jobs.status` |

### Adding Enum Values

PostgreSQL supports `ALTER TYPE ... ADD VALUE` (append only, no removal). Plan values carefully.

---

## Constraints & Validation

### CHECK Constraints

Use for invariants that the database must enforce regardless of application code:

```sql
-- Asset type must be one of the known categories
CHECK (type IN ('image', 'video', 'document', 'archive', 'file'))

-- Storage provider type
CHECK (type IN ('cloudinary', 'imagekit', 'r2', 'gcs'))

-- Exactly one target for invitations
CONSTRAINT invitations_one_target CHECK (
    (gallery_id IS NOT NULL)::int + (album_id IS NOT NULL)::int = 1
)
```

### NOT NULL by Default

- All columns are `NOT NULL` unless there is a specific reason for nullability.
- Nullable columns must have a comment explaining why (e.g., "NULL = never expires", "NULL = active").

---

## Schema Evolution

### Current Approach (MVP)

Single `schema.sql` file applied manually. No migration tool.

```bash
psql "$DATABASE_URL" -f internal/database/schema.sql
psql "$DATABASE_URL" -f internal/database/seed.sql
```

### Future: golang-migrate

When the project moves past MVP, schema changes use numbered migration files:

```
migrations/
  000001_initial_schema.up.sql
  000001_initial_schema.down.sql
  000002_add_asset_description.up.sql
  000002_add_asset_description.down.sql
```

### Rules for Schema Changes

1. Design the SQL migration first.
2. Update sqlc query definitions.
3. Regenerate sqlc code.
4. Update repository layer.
5. Update service/handler as needed.
6. Update these docs in the same PR.

### Safe Migration Patterns

| Operation | Safe? | Notes |
|-----------|-------|-------|
| Add column (nullable) | Yes | No lock, no rewrite |
| Add column (NOT NULL + DEFAULT) | Yes (PG 11+) | No rewrite thanks to fast default |
| Drop column | Careful | Remove all code references first |
| Rename column | No | Breaks queries; use add+migrate+drop |
| Add index | Yes | Use `CONCURRENTLY` in production |
| Add enum value | Yes | Append only |
| Remove enum value | No | Not supported; create new type |

---

## JSONB Usage

JSONB is used for semi-structured data that varies per row or per provider:

| Table | Column | Contains |
|-------|--------|----------|
| `assets` | `metadata` | Dimensions, duration, EXIF data — varies by asset type |
| `storage_providers` | `credentials` | Provider-specific auth keys (encrypted at rest by Neon) |
| `storage_locations` | `metadata` | Provider response data, transformation URLs |
| `clerk_webhook_events` | `payload` | Raw Clerk webhook body for replay/debugging |

### JSONB Rules

- Never use JSONB for data that needs relational queries (joins, foreign keys, aggregations).
- Do not index JSONB fields unless a specific query pattern demands it.
- Keep JSONB schemas documented (expected keys) even if not enforced by PostgreSQL.

---

## Views

Views are used for read-only convenience aggregations:

| View | Purpose |
|------|---------|
| `storage_account_usage` | Per-provider usage stats for the storage management UI |

Rules: views are read-only (no `INSTEAD OF` triggers). They simplify reads but are never used for writes.

---

**Next:** [Schema Reference](./schema-reference.md) — Detailed per-table column reference.
