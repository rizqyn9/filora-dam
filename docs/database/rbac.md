# Auth & RBAC Model

How Filora handles identity, sessions, and authorization at the data layer.

- **Web identity/sessions** → [Clerk](https://clerk.com) (external).
- **Terminal sessions** → our own `cli_sessions` tokens.
- **Authorization** → two tiers:
  1. **Global RBAC** — roles → permissions with `own`/`all` scope (app-wide capability).
  2. **Per-resource membership** — a local `owner`/`editor`/`viewer` role on each
     gallery and album (day-to-day sharing).

This is a private, family-scale app: the model is intentionally simple, not enterprise-hardened.

---

## 1. Identity (Clerk)

Clerk owns web login and session lifecycle. We do **not** store passwords or web
session tokens. We keep a local mirror of the user in `users`:

- `clerk_user_id` — the stable Clerk id, our link back to Clerk.
- `email`, `name`, `avatar_url` — mirrored profile fields.

### Provisioning: webhook + JIT
- **Webhook sync** — Clerk `user.created` / `user.updated` / `user.deleted` events
  upsert/deactivate the local `users` row.
- **JIT fallback** — on the first authenticated API request, if no local row matches
  the token's `clerk_user_id`, create one on the fly.

Deleting a user cascades to their roles, sessions, providers, assets, and locations
(all owning FKs are `ON DELETE CASCADE`).

---

## 2. Terminal sessions (CLI)

Web sessions are Clerk's; the CLI uses **our own** opaque tokens so it can work
outside a browser and manage multiple independent logins.

- Each terminal login creates a `cli_sessions` row.
- The raw token is generated server-side and shown to the CLI **once**; only its
  **SHA-256 hash** is stored in `token_hash`.
- A user may hold **many concurrent sessions** (one per device/terminal), each with
  an optional `label`, `ip_address`, `user_agent`.
- Sessions are independently **revocable** (`revoked_at`) and optionally expiring
  (`expires_at`, NULL = never).
- Active-session lookups use the partial index on `WHERE revoked_at IS NULL`.

### Token lifecycle
1. User authenticates (via Clerk) and requests a CLI token.
2. API generates a random token, stores `sha256(token)` + metadata, returns the raw token once.
3. CLI stores the raw token locally and sends it on each request.
4. API hashes the incoming token and looks it up; rejects if missing, revoked, or expired.
5. `last_used_at` is updated on use; `revoke` sets `revoked_at`.

> The token itself is never persisted in plaintext anywhere server-side.

---

## 3. Authorization (RBAC)

Access is decided by roles and permissions, never by ad-hoc user checks in code.

```
users ──< user_roles >── roles ──< role_permissions >── permissions
                                         │
                                       scope (own | all)
```

- A user has **many roles** (`user_roles`). That set is the user's "role group".
- A role has **many permissions** (`role_permissions`); each grant carries a **scope**.
- A permission is a `(resource, action)` pair (`permissions`).

### Permission = resource + action
Examples: `asset:read`, `asset:delete`, `storage:create`, `role:assign`, `session:revoke`.
Wildcards use `*`: `('*','*')` means full access.

### Scope qualifier
Each grant in `role_permissions` has a `scope`:

| Scope | Meaning |
|-------|---------|
| `own` | Applies only to rows the user owns (`row.user_id = current_user.id`) |
| `all` | Applies to every row in the family workspace |

So "can this user delete this asset?" resolves to:
1. Does any of the user's roles grant `asset:delete` (or `*:*` / `asset:*`)?
2. If the grant's scope is `all` → allow. If `own` → allow only when the asset's
   `user_id` matches the user.

### Evaluation order (recommended for the service layer)
1. Resolve the user's roles → their granted permissions (with scopes).
2. Match the requested `(resource, action)` against grants, honoring wildcards
   (`*` resource and/or `*` action).
3. Take the **widest** matching scope (`all` beats `own`).
4. If scope is `own`, enforce ownership on the target row (see membership below).
5. Deny by default if nothing matches.

---

## 3a. Gallery & album membership (per-resource)

Global RBAC decides *capability* (can this user create galleries, manage storage,
etc.). Access to a **specific** gallery or album is decided by local membership.

- `gallery_members` and `album_members` link a user to the resource with a
  `member_role`: `owner`, `editor`, or `viewer`.
- The resource `owner_id` also has a membership row with role `owner`, so an
  access check is a single membership lookup (no `owner_id` special-case needed).
- Suggested capability per membership role:

  | Local role | Gallery / album abilities |
  |------------|---------------------------|
  | `owner` | full control; invite/remove members; delete the resource |
  | `editor` | add/remove/organize assets; create tags; cannot manage members or delete the resource |
  | `viewer` | read + download only |

### How the two tiers combine
For an action on a gallery/album/asset, a user is allowed if **either**:
- their global grant scope is `all` (admin/superuser oversight), **or**
- their global grant scope is `own` **and** they have a sufficient membership
  role on the target gallery/album (or its parent gallery, for assets).

Resolve "own" for assets via `assets.gallery_id` → membership in that gallery
(and/or `assets.uploaded_by = user`). Superuser (`*:*`) always bypasses.

---

## 4. Seeded roles

Defined in [`seed.sql`](../../apps/api/internal/database/seed.sql). All are `is_system = true`.

| Role | Grants (summary) | Scope |
|------|------------------|-------|
| `superuser` | `*:*` (everything) | `all` |
| `admin` | asset/gallery/album/tag full, storage full, user read/update, role read/assign, session read/revoke, dashboard read | `all` |
| `member` | asset/gallery/album/tag full, storage read, dashboard read | `own` |
| `viewer` | asset read/download, gallery/album/tag read, dashboard read | `own` |

Notes:
- `admin` deliberately does **not** hold the raw `*:*` wildcard, nor `role:manage` — those are reserved for `superuser`.
- `admin` can `role:assign` (assign existing roles) but not `role:manage` (create/edit roles).
- `member`/`viewer` grants are scope `own`; actual reach into a gallery/album is
  further gated by the local membership role (see §3a).

### Permission catalog (seeded)

| Resource | Actions |
|----------|---------|
| `*` | `*` |
| `asset` | `read`, `create`, `update`, `delete`, `download` |
| `gallery` | `read`, `create`, `update`, `delete`, `invite` |
| `album` | `read`, `create`, `update`, `delete`, `invite` |
| `tag` | `read`, `create`, `update`, `delete` |
| `storage` | `read`, `create`, `update`, `delete` |
| `user` | `read`, `update`, `delete` |
| `role` | `read`, `assign`, `manage` |
| `session` | `read`, `revoke` |
| `dashboard` | `read` |

---

## 5. Superuser bootstrap

Because users come from Clerk, the owner's row only exists after their first sign-in.
Grant the superuser role afterwards (replace the email):

```sql
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'owner@example.com' AND r.slug = 'superuser'
ON CONFLICT DO NOTHING;
```

---

## 6. Boundaries & rules

- Never store passwords or Clerk session tokens in our DB.
- Never store CLI tokens in plaintext — only `sha256(token)`.
- Authorization decisions go through the RBAC resolution above, not hard-coded checks.
- New capabilities are modeled as new `(resource, action)` permission rows + grants,
  not as new boolean flags on `users`.
