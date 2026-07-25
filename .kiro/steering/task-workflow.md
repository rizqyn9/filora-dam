---
inclusion: manual
---

# Task-Driven Workflow

Structured feature/issue management via markdown files in `.ignore/tasks/`.

## Language Rule

All generated task file content (Contract, Progress Log, Meta, Manual Test Checklist) MUST be written in English regardless of the input language. The only exception is the `## Draft` section, which preserves the user's original input verbatim in whatever language they used.

## When to Activate

Use this workflow when the user:
- Sends a feature idea, bug report, or issue description
- References an existing task (e.g., "continue TASK-003")
- Says "grooming", "breakdown", or "plan this"

## File Conventions

- **Location:** `.ignore/tasks/`
- **Naming:** `TASK-{NNN}-{kebab-slug}.md` (scan existing files, increment highest ID)
- **Status flow:** `DRAFT` → `GROOMING` → `IN_PROGRESS` → `DONE` | `BLOCKED`
- **Permanence:** Never delete or reuse IDs

## Task File Template

```markdown
# TASK-{NNN}: [Short title]

## Meta

| Field   | Value              |
|---------|--------------------|
| Status  | DRAFT              |
| Created | YYYY-MM-DD         |
| Updated | YYYY-MM-DD         |
| Sessions| 1                  |
| Branch  | feat/task-NNN-slug |

---

## Draft

[Raw user input — preserved verbatim, never modified after grooming]

---

## Contract

### Problem Statement
[1-2 sentences]

### Acceptance Criteria
- [ ] Testable criterion

### Scope
- **In scope:** ...
- **Out of scope:** ...

### Technical Context
[Files, modules, dependencies discovered during research]

### Task Breakdown
1. Incremental task (each produces working code)

### Manual Test Checklist
**Happy Path:**
- [ ] Scenario — expected result

**Negative Cases:**
- [ ] Invalid input — expected error

**Edge Cases:**
- [ ] Boundary condition — expected behavior

---

## Progress Log

### Session N — YYYY-MM-DD

**Completed:**
- [x] What was done

**Changes:**
| File | Action | Description |
|------|--------|-------------|
| `path/file` | created/modified/deleted | Brief description |

**Decisions:**
- Decision + reasoning (the why)

**Next Steps:**
- Specific, actionable next items
```

## Workflow Phases

### Phase 1: DRAFT

1. Scan `.ignore/tasks/` to determine next ID.
2. Create `TASK-{NNN}-{slug}.md` from the template.
3. Copy user input verbatim into `## Draft`.
4. Set status = `DRAFT`.
5. Proceed immediately to Phase 2.

### Phase 2: GROOMING

Transform the draft into a structured contract:

1. **Comprehend** the user's intent from the draft.
2. **Research** the codebase — read files, grep patterns, check modules, read docs. Use any available tool.
3. **Fill the Contract:**
   - Problem Statement — concise, specific.
   - Acceptance Criteria — testable, checkbox-style.
   - Scope — explicit in/out boundaries to prevent creep.
   - Technical Context — file paths, module names, dependencies, architecture notes.
   - Task Breakdown — numbered, incremental. Each step builds on the previous and ends with integrated, working code.
   - Manual Test Checklist — happy path, negative, and edge cases.
4. Update status → `GROOMING`.
5. Present the contract to the user. Ask for confirmation or adjustments.
6. On user approval → status = `IN_PROGRESS`.

**Grooming rules:**
- Each task must be completable in a single session.
- No orphaned code — every task ends integrated.
- If a task is too large, split it.
- Follow Filora's design order: SQL migration → sqlc queries → repository → service → handler → UI.

### Phase 3: EXECUTION

1. Read the task file. Check the latest Progress Log for context.
2. Work through the Task Breakdown sequentially — never skip or reorder.
3. At session end, **always** append a Progress Log entry with: Completed items, Changes table (exhaustive), Decisions (with reasoning), Next Steps (specific).
4. Update Meta: increment `Sessions`, set `Updated` date.
5. When all Acceptance Criteria pass → status = `DONE`, check all AC items.

### Phase 4: RESUMPTION

When the user references an existing task in a new session:

1. Read the full task file.
2. Focus on the last Progress Log entry's "Next Steps".
3. Continue from where the previous session ended — do not redo completed work.
4. Increment session counter in Meta.

## Progress Log Standards

Write logs so a new session can understand full state without asking questions:
- Decisions include the *why*, not just what.
- Changes table lists every created/modified/deleted file.
- Next Steps are specific and actionable, never vague.
- Document blockers clearly when present.

## Constraints

- Never start coding without reading the task file first.
- Never skip grooming — even simple tasks get a contract.
- Never leave a session without updating the Progress Log.
- Never modify the `## Draft` section after grooming.
- Never work on tasks out of order from the Task Breakdown.
- Never create task files outside `.ignore/tasks/`.
