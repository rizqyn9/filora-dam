# Web App Design (v1)

The first shippable version of the Filora web app: a file manager that connects to the existing Go API.

---

## Scope

**Must-have (v1):**
- Login (email + password)
- Spaces (list, switch, create)
- Folder browser (tree in sidebar, create, rename, move, delete/trash)
- Asset browser (grid/list toggle, upload, rename, delete)
- Breadcrumb navigation

**Not in v1:**
- Tags
- Storage admin panel
- Session management
- Change password UI
- Register flow (accessed via invite link, not from login page)
- File previews (API doesn't expose file URLs yet — mime-type icons only)
- Multi-select / bulk actions (no bulk API endpoints)

---

## Layout

Sidebar + content area. Sidebar contains:
1. Space switcher (dropdown) at the top
2. Folder tree for the active space below

Content area contains:
1. Breadcrumbs + view toggle (grid/list) + Upload button in a toolbar
2. File/folder grid or list below

Mobile: sidebar collapses into a Sheet (slide-over) triggered by a hamburger button.

```
┌─────────────────────────────────────────────────┐
│ [≡]  Breadcrumb > Path > Here    [grid|list] [↑]│
├──────────┬──────────────────────────────────────┤
│ [Space▾] │                                      │
│          │   ┌────┐ ┌────┐ ┌────┐ ┌────┐       │
│ 📁 Photos│   │ 📄 │ │ 📄 │ │ 📄 │ │ 📄 │       │
│  📁 2024 │   └────┘ └────┘ └────┘ └────┘       │
│  📁 2023 │                                      │
│ 📁 Docs  │   ┌────┐ ┌────┐ ┌────┐              │
│ 📁 School│   │ 📄 │ │ 📄 │ │ 📄 │              │
│          │   └────┘ └────┘ └────┘              │
├──────────┴──────────────────────────────────────┤
│ 👤 User Name                                    │
└─────────────────────────────────────────────────┘
```

---

## Routing

TanStack Router file-based routing:

```
src/routes/
├── __root.tsx                          ← TooltipProvider, Toaster, devtools
├── login.tsx                           ← /login (public)
├── register.tsx                        ← /register?token=xyz (public, invite-only)
├── _authenticated.tsx                  ← layout: auth guard + sidebar+content shell
└── _authenticated/
    ├── index.tsx                       ← / → redirect to first space
    └── spaces/
        ├── $spaceId.tsx                ← layout: loads space + folder tree
        └── $spaceId/
            ├── index.tsx               ← /spaces/:spaceId (root-level files)
            └── folders/
                └── $folderId.tsx       ← /spaces/:spaceId/folders/:folderId
```

**Auth guard:** `_authenticated.tsx` checks for token presence in auth store. If missing → redirect to `/login`. On token present → fetch `/auth/me` to validate; on 401 → clear token, redirect to login.

**After login:** Redirect to `/spaces/:firstSpaceId` (fetch spaces, pick first).

---

## State Management

| Concern | Where | Why |
|---------|-------|-----|
| Auth token | Zustand store (persisted to localStorage) | Write-once-read-many, survives tab close |
| User profile | TanStack Query (`/auth/me`) | Server state, benefits from cache/refetch |
| Active space | URL param (`:spaceId`) | Bookmarkable, back-button works |
| Folder tree | TanStack Query (all folders for active space) | Server state, cached per space |
| Asset list | TanStack Query (paginated per folder) | Server state, paginated |
| View mode (grid/list) | Zustand store (persisted) | User preference |
| Theme (light/dark/system) | Zustand store (persisted) — already built | User preference |
| Sidebar open/closed | Zustand store (not persisted) | Resets on page load, fine |

---

## Auth Flow

1. User visits any protected route → `_authenticated.tsx` checks auth store for token
2. No token → redirect to `/login`
3. Token exists → call `GET /api/v1/auth/me` (via TanStack Query)
4. 401 response → clear token from store, redirect to `/login`
5. Success → render authenticated layout

**Login page:** Email + password form. On success, store token in zustand (→ localStorage), redirect to first space.

**Register page:** Only reachable via invite link (`/register?token=xyz`). Shows name + password form. On success, store token, redirect to first space.

---

## Folder Tree

- **Data loading:** One `GET /api/v1/folders?space_id=X` call (no parent_id) fetches ALL folders for the space. Build tree structure client-side from `parent_id` relationships.
- **Interaction:** Arrow icon expands/collapses children (no navigation). Folder name click navigates to that folder's route.
- **Active state:** Current folder (from URL) highlighted in tree with auto-expand of ancestors.
- **Mutations:** Create/rename/move/delete invalidate the folder query for that space.

---

## Asset Browser

- **View modes:** Grid (thumbnail cards with mime-type icon + filename) and List (table with name, size, date). Toggle persisted in zustand.
- **Loading:** Skeleton cards/rows while fetching. Pagination via offset/limit (load more or infinite scroll — start with "Load more" button).
- **Actions per item:** Kebab menu (three-dot) with: Rename, Delete (with confirmation dialog).
- **No multi-select in v1.**

---

## Upload

- **Trigger:** Drag-and-drop anywhere in content area OR "Upload" button (file picker).
- **Drag visual:** Subtle overlay on the content area when files are dragged over.
- **Progress:** Single persistent toast showing aggregate progress ("Uploading 3/12...").
- **Completion:** Toast transitions to "12 files uploaded" (auto-dismiss after 5s). On partial failure: "11 uploaded, 1 failed" (persists until dismissed).
- **During upload:** User can keep browsing — upload is non-blocking (mutations run in background).
- **API call:** `POST /api/v1/assets/upload` (multipart/form-data) per file, sequential or with concurrency cap.

---

## Delete Behavior

- **Always confirm** with a shadcn Dialog: "Move to trash?" for assets, "Move folder to trash? This includes all files inside." for folders.
- API does soft-delete (`deleted_at`). No undo toast (restore endpoint exists for folders only, not assets).

---

## Components

### Shadcn (install upfront)

button, dialog, dropdown-menu, input, label, separator, skeleton, scroll-area, avatar, select, breadcrumb, progress, sheet, toggle-group (+ already installed: sonner, tooltip)

### Custom (build during feature work)

| Component | Purpose |
|-----------|---------|
| `FolderTree` | Recursive tree with expand/collapse + active highlight |
| `FileGrid` | Grid of asset cards (icon + name + kebab) |
| `FileList` | Table view of assets |
| `UploadDropzone` | Invisible overlay detecting drag events |
| `UploadToast` | Persistent toast with aggregate progress |
| `SpaceSwitcher` | Dropdown showing user's spaces |

---

## Feature Directory Structure

```
src/features/
├── auth/
│   ├── api.ts           ← login/logout mutations, useMe query
│   ├── schemas.ts       ← Zod: AuthResponseSchema, UserSchema
│   └── auth-store.ts    ← zustand: token + login/logout actions
├── spaces/
│   ├── api.ts           ← useSpaces, useCreateSpace
│   └── schemas.ts       ← Zod: SpaceSchema
├── folders/
│   ├── api.ts           ← useFolders, useCreateFolder, useRenameFolder, etc.
│   └── schemas.ts       ← Zod: FolderSchema
└── assets/
    ├── api.ts           ← useAssets, useUpload, useRenameAsset, useDeleteRef
    └── schemas.ts       ← Zod: AssetSchema
```

---

## Loading & Empty States

- **File grid:** Skeleton cards/rows while loading.
- **Sidebar folder tree:** No skeleton (loads fast for <500 folders). Shows tree immediately.
- **Empty folder:** Centered message "This folder is empty" + Upload button.
- **Empty space (no folders):** "Get started — create a folder or upload files."

---

## Key Constraints

1. **No file previews in v1** — API doesn't expose file URLs. Show mime-type icons (lucide-react has icons for image, video, file-text, file-archive, etc.).
2. **No bulk operations** — API has no bulk endpoints. Single-item actions only.
3. **No tags** — feature exists in API but deferred from web v1 to keep scope tight.
4. **No dashboard** — land directly in the file browser after login.
5. **Register via invite link only** — no link from login page.

---

**Next:** Implementation tickets derived from this design.
