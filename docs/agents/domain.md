# Domain Docs

How the engineering skills consume this repo's domain documentation.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root — authoritative domain glossary.
- **`docs/adr/`** — architecture decision records. Read ADRs that touch the area you're about to work in.

If any of these files don't exist, proceed silently. The `/domain-modeling` skill (via `/grill-with-docs`) creates them lazily when terms or decisions get resolved.

## File structure

Single-context repo:

```
/
├── CONTEXT.md
├── docs/adr/
│   ├── 001-single-copy-storage.md
│   ├── 002-provider-side-preview.md
│   ├── 003-flat-tags.md
│   └── 004-archive-first-dam.md
└── api/
```

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in `CONTEXT.md`. Don't drift to synonyms the glossary explicitly avoids.

Key terms: Space (not Gallery), Asset, Asset Reference, Tag (flat), Serving Layer (L1), Archive Layer (L2), Storage Account.

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly:

> _Contradicts ADR-001 (single-copy storage), but worth reopening because…_
