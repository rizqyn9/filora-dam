# Roles & Permissions (in plain language)

Filora has **two layers of access control**. Together they answer "can this
person do this thing to this item?"

1. **Global roles (RBAC)** — what a person can do across the app.
2. **Membership roles** — what a person can do inside a *specific* gallery or album.

For the data model and exact grants, see [database/rbac.md](../database/rbac.md).

---

## 1. Global roles

A user can hold one or more of these. Access is checked as
`resource:action @ scope`, where **scope** is either `own` (things you own or
belong to) or `all` (the whole family workspace).

| Role | Think of it as | Reach |
|------|----------------|-------|
| **superuser** | The family owner / admin-of-admins | Everything, no limits (`*:*`) |
| **admin** | A trusted helper managing content & storage | Manage assets, galleries, albums, tags, storage accounts, users, sessions — across the whole workspace (`all`) |
| **member** | A normal family member | Create and manage **their own** galleries/albums/assets/tags (`own`) |
| **viewer** | A guest / read-only relative | View and download things they belong to (`own`, read-only) |

Notes:
- `admin` can do almost everything `superuser` can, **except** hold the raw
  wildcard and manage roles themselves (create/edit roles) — those stay with
  `superuser`.
- Storage accounts are managed centrally: only `superuser`/`admin` (or a user
  granted the `storage` permission) can add or edit them.

## 2. Membership roles (per gallery / album)

Sharing a specific gallery or album grants a **local** role to that person:

| Membership role | Can do |
|-----------------|--------|
| **owner** | Full control of the gallery/album; invite & remove members; delete it |
| **editor** | Add/remove/organize assets, create tags — but not manage members or delete it |
| **viewer** | View and download only |

The creator of a gallery/album is automatically its `owner`.

## How the two combine

To act on a gallery/album/asset, a user is allowed if **either**:

- their global grant has scope `all` (an admin/superuser overseeing everything), **or**
- their global grant has scope `own` **and** they have a sufficient **membership
  role** on that gallery/album.

`superuser` always bypasses these checks.

**Example** — deleting a photo in "Mom's Gallery":
- Superuser/admin: allowed (scope `all`).
- A member who is an `editor` of that gallery: allowed (own + editor).
- A member who is only a `viewer`: denied.
- A member with no membership there: denied.

## Personas (family context)

| Persona | Likely global role | Typical activity |
|---------|-------------------|------------------|
| Owner (you) | superuser | Set up storage accounts, invite family, everything |
| Partner / co-admin | admin | Help organize, manage members and content |
| Family member | member | Upload their own photos, make albums, invite for specific albums |
| Grandparent / guest | viewer | Browse and download shared galleries |

## Bootstrapping the superuser

Because identities come from Clerk, the owner's account only exists after their
first sign-in. After that, grant the superuser role once (SQL in
[`seed.sql`](../../apps/api/internal/database/seed.sql) / [rbac.md](../database/rbac.md#5-superuser-bootstrap)).
