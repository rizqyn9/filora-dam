---
inclusion: manual
---

# Documentation Standards

Rules for writing and maintaining documentation across this project.

## Structure & Organization

- **One topic per file**, max ~300 lines. If it's longer, split it.
- Every folder with multiple docs gets a `readme.md` as index/TOC.
- Use `#[[file:path]]` references instead of duplicating content across files.
- Filenames: `kebab-case.md`. Use numeric prefix for ordered content (`01-`, `02-`, `03-`).
- Headings max 3 levels deep (`#`, `##`, `###`). Deeper nesting means the doc should be split.

## Content Formatting

- Start every doc with a **one-sentence summary** of what it covers (first line after the title).
- Use **tables** for structured data (features, comparisons, status lists, mappings).
- Use **ASCII diagrams** for flows and architecture — not images. AI agents can read text; images are opaque.
- Use fenced code blocks with language tags for code/config/commands.
- Short paragraphs: 3–4 sentences max. One idea per paragraph.

## Language & Tone

- **English** (plain, accessible) for all documentation.
- Active voice, present tense: "The system processes..." not "The system will process..."
- No jargon without definition — every domain term must be in the glossary (`docs/qualix/08-glossary.md`) or defined inline on first use.
- Be direct. Avoid hedging ("basically", "kind of", "should probably").
- No filler intros ("In this document we will discuss..."). Start with the substance.

## Cross-Referencing

- Use **relative links** (`./other-file.md`) not absolute paths.
- Every doc ends with a `**Next:**` link to the logical next document.
- The glossary is the single source of truth for terminology — reference it, don't redefine terms.

## Maintenance

- Docs are updated in the **same PR** as the code they describe. No orphan docs.
- Outdated docs are deleted, not left with "TODO: update" markers.
- If a feature changes significantly, the doc is rewritten — not patched with addendums.

## Doc Locations

| Type | Location |
|------|----------|
| Product docs (what the product does, workflows, features) | `docs/qualix/` |
| Feature behavior specs (API contracts, status transitions, data models) | `docs/{feature}/` |
| App-internal architecture (layers, module structure, setup, deploy) | `bedrock-app/docs/` |
| System-level docs (traffic flow, infra cost, cross-cutting) | `docs/` root |
| Agent instructions (coding rules, conventions, registries) | `.kiro/steering/` |

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
