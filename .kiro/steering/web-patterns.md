---
inclusion: fileMatch
fileMatchPattern: "web/**"
---

# Web App Patterns

Conventions for the Filora React frontend. Follow these — don't reinvent.

## State Management

| Concern | Where | Why |
|---------|-------|-----|
| Auth token | `features/auth/auth-store.ts` (zustand + persist to localStorage) | Write-once-read-many |
| User profile | TanStack Query (`/auth/me`) | Server state, cached |
| Active space/folder | URL params (`:spaceId`, `:folderId`) | Bookmarkable |
| Folder tree | TanStack Query (`/folders?space_id=X`) | All folders loaded at once, tree built client-side |
| Asset list | TanStack Query (paginated) | Server state |
| View mode (grid/list) | `stores/ui-store.ts` (zustand + persist) | User preference |
| Theme | `stores/ui-store.ts` (zustand + persist) | User preference |
| Sidebar open/closed | `stores/ui-store.ts` (not persisted) | Transient UI state |

## API Client Pattern

- `src/lib/api.ts` is the single fetch wrapper. It handles the response envelope and auth header.
- Public auth routes (`/auth/login`, `/auth/register`) skip the `/api/v1` prefix.
- All protected routes go through `/api/v1/...`.
- Multipart uploads bypass the JSON wrapper — use raw `fetch` with FormData.

## Feature Directory Convention

```
src/features/<domain>/
├── api.ts        ← TanStack Query hooks (useX) + mutations (useCreateX, etc.)
├── schemas.ts    ← Zod schemas + inferred types (no separate types.ts)
└── *-store.ts    ← Zustand store (only if needed, e.g. auth-store.ts)
```

- Zod schemas ARE the types (`z.infer<>`). No generated types, no separate interfaces.
- One `api.ts` per feature — contains both queries and mutations.
- Query keys follow: `["domain", ...params]` (e.g. `["assets", spaceId, folderId]`).

## Component Organization

- `src/components/ui/` — Shadcn primitives (don't edit these manually)
- `src/components/layout/` — Shell components (sidebar, toolbar, space-switcher)
- `src/components/assets/` — Asset browser, grid, list, upload, dialogs
- `src/components/folders/` — Folder tree, mutation dialogs
- `src/components/spaces/` — Create space dialog

## Routing Rules

- File-based routing via TanStack Router (`src/routes/`)
- `_authenticated.tsx` layout guards all protected routes
- `$spaceId.tsx` layout wraps all space-scoped routes
- Route params are the source of truth for "current location" — never store in zustand

## Dialog Pattern

Every mutation dialog follows:
1. Receives `open` + `onOpenChange` props (controlled from parent)
2. Calls the mutation hook from `features/<domain>/api.ts`
3. Shows toast on success, inline error on failure
4. Invalidates the relevant query on success

## Shadcn v4 (base-ui) Gotchas

- `BreadcrumbLink` uses `render` prop, not `asChild`
- `ToggleGroup` uses `value` as an array (not `type="single"` + string)
- `TooltipProvider` has no `delayDuration` prop
- `Select.onValueChange` can receive null — guard with `if (v)`
