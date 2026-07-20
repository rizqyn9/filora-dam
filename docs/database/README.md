# Database Documentation

Complete reference for Filora's PostgreSQL data model, design decisions, and conventions.

---

## Source of Truth

The canonical schema lives at `apps/api/internal/database/schema.sql`. If these docs and the SQL disagree, the SQL wins and these docs must be updated.

## Documents

| Doc | Covers |
|-----|--------|
| [ERD](./erd.md) | Entity-relationship diagram (ASCII), table groupings, key relationships |
| [Design Standards](./design-standards.md) | Naming, types, ID strategy, indexing, enum, and migration conventions |
| [Schema Reference](./schema-reference.md) | Per-table column details, constraints, indexes, and domain notes |

## Quick Facts

| Property | Value |
|----------|-------|
| Engine | PostgreSQL (hosted on Neon) |
| Query layer | sqlc (type-safe, no ORM) |
| Migrations | Single canonical `schema.sql` applied manually (no migration tool yet) |
| ID strategy | BIGINT identity for lookup tables, UUID v7 for high-volume/external rows |
| Soft delete | `deleted_at` on `assets` only |
| Timestamps | `created_at` + `updated_at` (via trigger) on all tables |

## Domain Groups

```
Identity & Access    users, roles, permissions, role_permissions,
                     user_roles, cli_sessions, clerk_webhook_events

Organization         galleries, gallery_members, albums, album_members,
                     invitations, album_assets, tags, asset_tags

Assets               assets

Storage              storage_providers, storage_locations, archive_sync_jobs
```

---

**Next:** [ERD](./erd.md) — Visual overview of all tables and their relationships.
