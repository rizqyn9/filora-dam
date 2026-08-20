# ADR-003: Flat Tags Over Hierarchical

Tags are flat labels with no parent-child relationships.

---

## Status

Accepted

## Context

Tags provide cross-cutting grouping beyond the folder hierarchy. Two models
were evaluated:

1. **Hierarchical tags** — tags have parent-child relationships
   (e.g. `Event > Birthday`, `Location > Bali > Beach`). Powerful but adds
   complexity to schema (self-referential table, recursive queries), UI (tree
   picker), and search (match parent includes children?).
2. **Flat tags** — simple string labels, no nesting. Scoped per space.

## Decision

Use flat tags. A tag is a short string label, unique per space. No parent-child
relationship.

### Convention for pseudo-hierarchy

Users who want grouping can use naming conventions (e.g. `event:birthday`,
`location:bali`) but the system treats these as opaque strings with no special
semantics.

## Consequences

**Positive:**
- Simple schema: `tags` table + `asset_tags` junction. No recursive queries.
- Simple UI: flat list, autocomplete, multi-select.
- Simple search: exact match or prefix/contains filter.
- Fast to implement, easy to extend later if hierarchy is truly needed.

**Negative:**
- No built-in "show me all tags under Events" query. Users must rely on naming
  conventions or search.
- If hierarchy is ever needed, migration path is adding a nullable `parent_id`
  column — non-breaking.

---

**Next:** [ADR-004](./004-archive-first-dam.md) — Archive-first DAM positioning.
