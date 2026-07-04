# Domain Concepts & Glossary

The precise meaning of every core term in Filora. These terms are used
consistently across product, architecture, database, and code. When in doubt,
this page is the shared vocabulary.

---

## People & access

**User**
A person in the family. Identity (login, profile) is owned by [Clerk](https://clerk.com);
Filora keeps a local mirror. See also [Auth](../architecture/auth.md).

**Superuser**
The most privileged role — full, unrestricted access. Typically the family owner.
Represented by the `superuser` role holding the wildcard permission.

**Role (global)**
A named set of capabilities: `superuser`, `admin`, `member`, `viewer`. A user can
hold several roles at once. See [roles.md](./roles.md).

**Permission**
A single capability expressed as `resource:action` (e.g. `asset:read`,
`gallery:invite`), granted to a role with a **scope**.

**Scope**
How far a permission reaches: `own` (only things the user owns/belongs to) or
`all` (the whole family workspace).

**Membership role (local)**
Access to a *specific* gallery or album: `owner`, `editor`, or `viewer`. This is
separate from global roles and governs day-to-day sharing.

**Invitation**
An offer (sent to an email address) to join a gallery or album at a given
membership role. The invitee may not have an account yet; they accept via a link
after signing in through Clerk.

---

## Organization

**Gallery**
The top-level space that holds assets. **Every user gets one default gallery**
automatically. A user can own several galleries and can be a member of other
people's galleries. Storage **quota is tracked per gallery**.

**Album**
A grouping of assets **inside a gallery** (e.g. "Summer 2024"). An album has an
owner who can invite other members. One asset can appear in **multiple albums**.

**Tag**
A short label used to filter, search, and organize assets. Tags are a shared
vocabulary **scoped to a gallery**; an asset can have many tags.

**Asset**
A logical file record — a photo, video, document, archive, or other file — plus
its metadata (name, type, size, checksum, tags, etc.). The asset is the "truth";
the actual bytes live in storage locations.

**Trash (soft delete)**
Deleting an asset moves it to trash (marks `deleted_at`) rather than destroying
it, so accidental deletes can be recovered.

---

## Storage

**Storage provider / account**
A concrete cloud account Filora uploads to (e.g. "ImageKit #1", "Cloudinary #2",
"GCS Archive"). Accounts are **global** and managed by admins — not owned by
end users. Many accounts can exist per layer to work around per-account limits.

**Layer**
Filora stores every asset in two layers:

- **Serving layer** — hot, fast, publicly viewable free-tier providers
  (Cloudinary, ImageKit). This is what the app serves and displays.
- **Archive layer** — cold, cheap, archive-class storage (Google Cloud Storage
  Archive, Cloudflare R2, etc.). This is the safety copy.

**Storage location**
A physical copy of an asset in one provider account. An asset has at least one
serving location and at least one archive location. Each location tracks its own
status (`pending` → `stored` / `failed`).

**Account election** *(backlog)*
The (not-yet-built) strategy that decides *which* account within a layer a new
upload lands on (e.g. most free space, round-robin).

**Archive sync job**
A background task that replicates an asset to the archive layer with retries.
Uploads reach the serving layer immediately; archiving happens asynchronously.

**Quota**
The storage limit, tracked **per gallery**. Physical capacity is spread across
multiple accounts per layer, so a gallery's quota is independent of any single
account's free-tier cap.

**Deduplication**
Identical files (same SHA-256 hash) within a gallery are stored once. Dedup is
scoped per gallery and ignores trashed assets.

---

## Clients & access methods

**Web app**
The React front-end. Login and sessions handled by Clerk.

**CLI (terminal)**
A command-line client that talks to the API over HTTP. Supports logging in from
a terminal and managing **multiple concurrent sessions** via opaque tokens.

**API**
The Go backend that holds all business logic. The web app and CLI are thin
clients over it.

---

## Data model note

For exact tables, columns, and constraints behind these concepts, see the
[Database schema reference](../database/schema.md) and [ERD](../database/erd.md).
