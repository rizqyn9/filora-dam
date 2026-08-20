# ADR-004: Archive-First DAM Positioning

Filora is an archive-first Digital Asset Management system — safety and backup are the primary value, organization and browsing are the UX layer on top.

---

## Status

Accepted

## Context

Two mental models for a file management system:

1. **DAM-first** — the product is about browsing, organizing, and collaborating
   on assets. Backup is a secondary benefit.
2. **Archive-first** — the product is about safely storing everything with
   guaranteed redundancy. Organization (folders, tags, search) is the interface
   that makes the archive usable day-to-day.

The family's primary need: "never lose a file." Photos from phones, important
documents, videos — they need to exist somewhere safe, across multiple physical
copies, without paying enterprise prices.

## Decision

Filora is **archive-first**. Every design decision prioritizes:

1. **Data safety** — every asset reaches both serving and archive layers.
2. **Storage efficiency** — maximize free-tier capacity via account pooling and
   single-copy references.
3. **Recoverability** — soft deletes, trash retention, archive-layer redundancy.

Organization features (spaces, folders, tags, search, preview) exist to make the
archive **usable** — not as the core value prop.

### What this means in practice

- Upload flow prioritizes "get bytes to storage safely" over "generate rich
  metadata."
- Archive sync failure is a critical alert; missing thumbnail is cosmetic.
- Account election strategy optimizes for spreading data across accounts (safety
  through distribution), not for read performance.
- Features like smart categorization, AI tagging, advanced search are backlog
  items that enhance the UX layer — they do not block MVP.

## Consequences

**Positive:**
- Clear priority when tradeoffs arise: safety > UX polish.
- MVP scope stays small: upload, store, archive, basic browse. Ship fast.
- Users trust the system because the archive guarantee is simple and verifiable.

**Negative:**
- UX for browsing/organizing may feel basic at launch compared to Google Photos
  or Dropbox.
- Advanced DAM features (AI tagging, smart albums, face detection) are explicitly
  deferred.

**Accepted:**
- For a 3–5 person family, basic organization (spaces, folders, tags) is
  sufficient. Advanced features can be layered on once the archive foundation
  is proven reliable.

---

**Previous:** [ADR-003](./003-flat-tags.md) — Flat tags.
