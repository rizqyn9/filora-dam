---
inclusion: manual
---

# Git Commit Conventions

Rules for writing commit messages in the Filora project.

## Format

```
type(scope): subject
```

- Imperative mood, lowercase subject, no trailing period.
- Max 72 characters for the subject line.
- One logical change per commit (atomic). Don't mix refactors with features.
- Subject-only by default. Add a body only when the "why" is non-obvious.
- No emoji prefixes.

## Types

| Type | Use for |
|------|---------|
| `feat` | New feature or user-facing behavior |
| `fix` | Bug fix |
| `refactor` | Code restructuring with no behavior change |
| `chore` | Tooling, deps, config, CI — no production logic change |
| `docs` | Documentation only |
| `test` | Adding or updating tests |
| `perf` | Performance improvement |

## Scopes

Scope reflects the area of code changed. Use the most specific applicable scope.

| Scope | Applies to |
|-------|-----------|
| `api` | `api/` — handlers, services, middleware, server config |
| `cli` | `cli/` — CLI client code |
| `web` | `web/` — React frontend |
| `db` | Database migrations, sqlc queries, schema changes |
| `auth` | Authentication/authorization (Clerk integration, RBAC) |
| `storage` | Storage adapters (Cloudinary, ImageKit, R2) |
| `assets` | Asset module (upload, dedup, trash, download) |
| `spaces` | Space module (members, invitations) |
| `folders` | Folder module (hierarchy, navigation) |
| `tags` | Tag module |
| `dashboard` | Dashboard module |
| `docs` | `docs/` directory (use `docs` type + omit scope for doc-only changes) |
| `config` | App configuration, env vars |

Multi-scope is allowed when a change spans two areas: `feat(api,web): ...`

Omit scope for truly cross-cutting changes: `chore: update Go dependencies`

## Body (when needed)

Separate from subject with a blank line. Wrap at 72 characters. Explain **why**, not what — the diff shows what changed.

```
fix(assets): prevent duplicate upload when hash matches trashed asset

The dedup check was not considering soft-deleted assets, causing a
constraint violation on re-upload after trash.
```

## Staging

- Stage specific files, not `git add .` or `git add -A`.
- Never commit `.env`, secrets, or generated files (`bin/`, `node_modules/`).
- Run linters/formatters before committing — don't create fixup commits for lint.

## Examples

```
feat(spaces): add member invitation endpoint
fix(api): return 404 for missing space instead of 500
refactor(storage): extract upload validation to shared helper
chore(web): update TanStack Router to v1.120
docs: add storage adapter architecture doc
test(auth): add RBAC permission boundary tests
perf(db): add index on assets.space_id for listing queries
```
