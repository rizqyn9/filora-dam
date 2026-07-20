---
inclusion: manual
---

# Git Commit

Format: `type(scope): subject`

- Imperative mood, max 72 chars, no period
- One logical change per commit (atomic)
- Subject-only — no body unless the "why" is non-obvious

Types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `perf`

Scopes: `agents`, `workers`, `infrastructure`, `services`, `ui`, `api`, `models`, `terraform`, `config`, `tooling`, `workflows`, `prompts`

Multi-scope: `feat(agents,ui): Add test case editing`

Examples:
```
feat(workers): Add Playwright script retry logic
fix(api): Return 404 for missing scenario
refactor(infrastructure): Extract S3 upload to shared helper
chore(tooling): Update ruff to 0.8
```
