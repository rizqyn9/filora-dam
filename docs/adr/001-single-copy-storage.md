# ADR-001: Single-Copy Storage with References

Assets are stored once physically; spaces and folders reference them via join records.

---

## Status

Accepted

## Context

Filora pools many free-tier storage accounts to maximize capacity. Users can
share assets across spaces and place one asset in multiple folders. Two models
were considered:

1. **Copy-per-location** — duplicate the bytes for every space/folder that
   references an asset.
2. **Single-copy with references** — store bytes once, link via database records.

Free-tier accounts have hard quota limits (typically 1–25 GB per account). Every
physical duplicate halves effective capacity. With 3–5 family members sharing
content, duplication would exhaust quotas rapidly.

## Decision

Store each asset's bytes exactly once in the serving layer (and once in the
archive layer). Spaces and folders hold **references** (join table rows) to the
asset, not copies.

### Ownership rules

- An asset has an `owner_id` (the user who uploaded it).
- A space holds asset references. Multiple spaces can reference the same asset.
- A folder holds asset references (many-to-many). Multiple folders in the same
  space can reference the same asset.
- Removing a reference from a folder/space does not delete the physical file.
- Physical deletion occurs only when **zero references remain** across all spaces
  and folders, and the asset has been in trash past the retention period.

### Sharing semantics

When User A shares an asset with User B's space, the system creates a new
reference record pointing to the same asset. No bytes are copied.

## Consequences

**Positive:**
- Maximizes free-tier storage utilization.
- Deduplication is natural — same SHA-256 = same asset, one copy.
- Sharing is instant and zero-cost in storage terms.

**Negative:**
- Requires reference counting or orphan detection to know when physical deletion
  is safe.
- Owner deletion needs care: warn if other references exist, or soft-delete and
  let references go stale gracefully.
- Query complexity slightly higher (joins through reference table).

**Accepted tradeoffs:**
- The reference-counting complexity is small compared to the storage savings at
  family scale with free-tier accounts.
- Orphan cleanup can be a simple periodic job (backlog).

---

**Next:** [ADR-002](./002-provider-side-preview.md) — Provider-side preview for images.
