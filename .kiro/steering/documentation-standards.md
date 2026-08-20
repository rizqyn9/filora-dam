---
inclusion: fileMatch
fileMatchPattern: ['**/*.md', 'docs/**/*']
---

# Documentation Standards

Rules for writing and maintaining documentation in the Filora project.

## Structure & Organization

- One topic per file, max ~300 lines. Split if longer.
- Every folder with multiple docs gets a `README.md` as index/TOC.
- Use `#[[file:path]]` references instead of duplicating content across files.
- Filenames: `kebab-case.md`. Use numeric prefix for ordered content (`01-`, `02-`).
- Headings max 3 levels deep (`#`, `##`, `###`). Deeper nesting means the doc should be split.

## Content Formatting

- Start every doc with a one-sentence summary after the title.
- Use tables for structured data (features, comparisons, status, mappings).
- Use ASCII diagrams for flows and architecture — not images.
- Use fenced code blocks with language tags for code/config/commands.
- Short paragraphs: 3–4 sentences max. One idea per paragraph.

## Language & Tone

- English, plain and accessible.
- Active voice, present tense: "The system processes..." not "The system will process..."
- No jargon without definition — define domain terms inline on first use or reference the product concepts doc (`docs/product/concepts.md`).
- Be direct. No hedging ("basically", "kind of", "should probably").
- No filler intros. Start with substance.

## Cross-Referencing

- Use relative links (`./other-file.md`) not absolute paths.
- End each doc with a `**Next:**` link to the logical next document when part of a sequence.
- If docs and code/SQL disagree, the code/SQL wins — fix the docs.

## Maintenance

- Docs update in the same PR as the code they describe. No orphan docs.
- Outdated docs are deleted, not left with "TODO: update" markers.
- If a feature changes significantly, rewrite the doc — don't patch with addendums.

## Doc Locations

| Type | Location |
|------|----------|
| Product docs (domain, features, roles, roadmap) | `docs/product/` |
| Architecture (apps, layers, flows, storage, auth) | `docs/architecture/` |
| Database (schema reference, ERD, design rules, RBAC) | `docs/database/` |
| Per-app docs (setup, internal architecture, API spec) | `<name>/README.md`, `<name>/API.md` |
| Agent instructions (coding rules, conventions) | `.kiro/steering/` |

## Documentation Lookup Strategy

To find relevant documentation efficiently, use the index-first approach:

1. Start with the **topic index** (README.md) in the relevant folder — never load all docs at once.
2. Read only the specific file that matches your need from the index table.
3. If the topic spans multiple areas, read each area's index first, then drill into the matching file.

### Topic-to-Index Map

Use this table to resolve a knowledge need to the correct index file:

| When you need to know about... | Read this index first |
|---|----|
| What Filora is, domain terms, features, user roles | `docs/product/README.md` |
| Database design, schema, tables, columns, ERD, naming conventions | `docs/database/README.md` |
| Architecture, system design, request flows, storage layers, auth design | `docs/architecture/README.md` (if exists) or `api/README.md` |
| API endpoints, HTTP contracts, route structure | `api/API.md` |
| How to build, run, test, deploy the API | `api/README.md` |
| How to build, run the web app | `web/README.md` |
| Coding rules, commit conventions, agent behavior | `AGENTS.md` |
| All documentation entry point | `docs/README.md` |

### Detailed Index Contents

**`docs/product/README.md`** resolves to:

| File | Covers |
|------|--------|
| `docs/product/overview.md` | Vision, problem, target users, value proposition |
| `docs/product/concepts.md` | Domain glossary — meaning of every core term |
| `docs/product/features.md` | Feature catalog |
| `docs/product/roles.md` | Personas and roles/permissions model |
| `docs/product/roadmap.md` | MVP scope, implementation status, backlog |

**`docs/database/README.md`** resolves to:

| File | Covers |
|------|--------|
| `docs/database/erd.md` | Entity-relationship diagram, table groupings, relationships |
| `docs/database/design-standards.md` | Naming, types, ID strategy, indexing, enum, migration conventions |
| `docs/database/schema-reference.md` | Per-table column details, constraints, indexes, domain notes |

**Canonical schema source:** `api/internal/database/schema.sql`
**sqlc queries:** `api/internal/database/queries/*.sql`

### Code Location Quick Reference

| Area | Path |
|------|------|
| API modules (vertical slices) | `api/internal/modules/{auth,asset,dashboard,folder,rbac,space,storage,tag}/` |
| Auth & middleware | `api/internal/auth/` |
| Config & env | `api/internal/config/` |
| Database layer (pgx pool) | `api/internal/database/` |
| sqlc generated code | `api/internal/database/db/` |
| Shared utilities | `api/internal/lib/` |
| Server setup & compose root | `api/cmd/server/main.go` |
| Worker (archive replication) | `api/cmd/worker/main.go` |

## Template for New Docs

```markdown
# Title

One-sentence summary of what this document covers.

---

## Section 1

Content here.

## Section 2

Content here.

---

**Next:** [Next Document](./next-doc.md) — Brief description.
```
