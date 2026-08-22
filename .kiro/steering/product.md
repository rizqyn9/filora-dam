---
inclusion: always
---

# Filora Product Context

Filora is a private, archive-first Digital Asset Management platform for a family (3–5 people).

## Core Promise

Files are safe, organized, and findable. Users never manage storage.

## Positioning

Archive-first: data safety is the primary value. Organization (spaces, folders, tags, search) is the UX layer that makes the archive usable day-to-day.

## How It Works

- Users upload any file type into their **space**.
- Filora stores each file **once** physically (single-copy, shared by reference).
- Files land on a **serving layer (L1)** — free-tier providers pooled across many accounts.
- An async job copies files to an **archive layer (L2)** — cheap cold storage (pluggable).
- Users organize with **folders** (hierarchical, virtual references) and **tags** (flat labels per space).
- Sharing = inviting another user to your space (editor or viewer).

## Key Design Decisions

- Single physical copy, shared by reference across spaces/folders (ADR-001).
- Provider-side image preview via URL transforms; non-image files get icons (ADR-002).
- Flat tags, no hierarchy (ADR-003).
- Archive-first: safety > UX polish (ADR-004).
- Self-managed auth: opaque tokens + in-process cache (ADR-005).

Full domain vocabulary: `CONTEXT.md`. Full ADRs: `docs/adr/`.
