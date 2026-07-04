# Auth Architecture

How Filora authenticates users and enforces access. The data model behind this
lives in [database/rbac.md](../database/rbac.md); this page is the runtime /
enforcement view.

Three moving parts:

1. **Clerk** — web identity & sessions.
2. **CLI tokens** — terminal sessions we issue ourselves.
3. **Authorization** — global RBAC + per-resource membership.

---

## 1. Web identity (Clerk)

[Clerk](https://clerk.com) owns web login and session lifecycle. Filora stores
**no passwords** and **no web session tokens** — only a local mirror of the user.

**Provisioning (webhook + JIT):**
- **Webhook sync** — Clerk `user.created` / `user.updated` / `user.deleted`
  events upsert/deactivate the local `users` row. Deliveries are made idempotent
  via `clerk_webhook_events` (dedupe on the event id).
- **JIT fallback** — on the first authenticated API request, if no local row
  matches the token's `clerk_user_id`, create one on the fly.

**On first provisioning:** create the user's **default gallery** and its
`owner` membership row.

**Request path:**
```
Web request with Clerk session token
  → auth middleware: verify token with Clerk → clerk_user_id
  → resolve (or JIT-create) local users row → current user
  → handler / service
```

## 2. Terminal sessions (CLI)

The CLI can't use a browser session, and a user may be logged in from several
machines, so Filora issues its **own opaque tokens**.

- Each terminal login creates a `cli_sessions` row.
- The raw token is generated server-side and shown **once**; only its
  **SHA-256 hash** is stored (`token_hash`).
- A user may hold **many concurrent sessions**, each with an optional `label`
  (device name), and each independently **revocable** (`revoked_at`) and
  optionally expiring (`expires_at`).

**Token lifecycle:**
```
1. User authenticates (via Clerk) and requests a CLI token.
2. API generates a random token, stores sha256(token) + metadata, returns raw token once.
3. CLI stores the raw token locally, sends it on each request.
4. API hashes the incoming token, looks it up; rejects if missing/revoked/expired.
5. last_used_at is updated on use; revoke sets revoked_at.
```

The raw token is never persisted server-side in plaintext.

## 3. Authorization (two tiers)

Once the middleware has resolved the **current user** (via Clerk or a CLI token),
the **service layer** decides what they may do. There are two tiers, checked
together.

### Tier 1 — Global RBAC
`users → user_roles → roles → role_permissions → permissions`, each grant carrying
a **scope** (`own` | `all`). Permissions are `resource:action` pairs; wildcards
use `*`. Resolution:

```
1. Load the user's roles → granted permissions (with scopes).
2. Match requested (resource, action), honoring '*' wildcards.
3. Take the widest matching scope (all > own).
4. If scope = own, enforce ownership/membership (Tier 2).
5. Deny by default.
```

Roles: `superuser` (`*:*`), `admin` (workspace-wide), `member` (own), `viewer`
(own, read-only). See [roles.md](../product/roles.md).

### Tier 2 — Per-resource membership
Access to a **specific** gallery/album uses a local `member_role`
(`owner`/`editor`/`viewer`) in `gallery_members` / `album_members`. The resource
owner also holds an `owner` membership row, so checks are a single lookup.

### Combining them
Allowed if **either**:
- global grant scope is `all` (admin/superuser oversight), **or**
- global grant scope is `own` **and** the user has a sufficient membership role
  on the target gallery/album (for an asset, via its `gallery_id` and/or
  `uploaded_by`).

`superuser` (`*:*`) always bypasses.

## Invitations

Inviting someone to a gallery/album (`invitations`) targets an **email**, which
may not yet be a user:
```
1. Owner/admin creates an invitation (email, target gallery/album, role, token).
2. Invitee opens the link → signs in via Clerk (creating their users row if new).
3. On accept: create the membership row, set status=accepted + accepted_user_id.
```

## Enforcement rules (for implementers)

- Do the auth/permission check in the **service layer**, not scattered in handlers.
- Never hard-code user-id comparisons; go through the RBAC + membership resolution.
- Store only hashes of CLI tokens; never log raw tokens.
- Treat every external input as untrusted; validate with validator v10.

## See also

- [Database: RBAC & auth model](../database/rbac.md)
- [Product: roles & permissions](../product/roles.md)
- [Architecture overview](./overview.md)
