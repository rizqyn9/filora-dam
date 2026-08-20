# Filora Domain Glossary

The shared vocabulary for Filora. Every term here is agreed and used consistently
across product docs, architecture, database, and code.

---

## People & Access

**User** — A person in the family. Self-managed auth (email + bcrypt password).
Invite-only registration — superuser invites via opaque token link.

**Superuser** — The family owner. Full unrestricted access (wildcard permission).
Created via `cmd/seed` on first setup.

**Role (global)** — A named set of capabilities: `superuser`, `admin`, `member`,
`viewer`. A user can hold multiple roles. Roles are hardcoded in application
code; only assignments are stored in DB (`user_roles` table).

**Permission** — A single capability expressed as `resource:action` with a scope
(`own` or `all`).

**Membership** — Per-space access grant. An invited user gets one role on that
space: `editor` or `viewer`. The space creator is always `owner`.

**Session** — A server-side record for an authenticated client (web or CLI).
Opaque random token (32 bytes), stored as SHA-256 hash. Sliding window TTL:
web = 7 days idle, CLI = 90 days idle. Resolved via in-process LRU cache with
DB fallback (zero external deps, no Redis).

**Invitation** — An offer (email + role + space) to join Filora. Contains an
opaque token (hashed in DB). Delivered manually (superuser shares link via
WhatsApp/chat). Invitee visits registration URL, sets password, becomes a user.

---

## Organization

**Space** — Top-level container. Every user gets a default space on signup. A user
can own multiple spaces. Sharing = inviting another user to your space with a
membership role.

**Folder** — Hierarchical organizer within a space. Supports unlimited nesting
via `parent_id`. A file does not need to be in a folder (root-level is allowed).

**Asset** — A logical file record (image, video, document, archive, any file type)
plus its metadata. Represents one physical copy of bytes, shared by reference
across spaces and folders.

**Asset Reference** — The many-to-many link between an asset and a folder (or
space root). One asset can appear in multiple folders within the same space and
across multiple spaces. Physical bytes are stored once.

**Tag** — A flat label for cross-cutting grouping, scoped per space. An asset can
have many tags.

**Trash (soft delete)** — Deleting an asset reference marks `deleted_at` on that
reference row. Deleting a folder marks `deleted_at` on the folder (contents
hidden but recoverable). The physical asset is destroyed only when zero
references remain and 30 days have passed (retention period).

**Deduplication** — Global, by SHA-256 hash. If an uploaded file matches an
existing asset's checksum, no new bytes are stored — the system creates a new
reference to the existing asset instead.

---

## Storage

**Storage Provider** — A type of cloud service: `cloudinary`, `imagekit`, `r2`,
`gcs`. Defined as a port/interface; implementations are pluggable adapters.

**Storage Account** — One concrete cloud account (e.g. "Cloudinary #3",
"ImageKit #7"). Global, admin-managed. Many accounts per provider per layer.

**Serving Layer (L1)** — Hot, free-tier storage. Cloudinary, ImageKit, and similar
providers that can serve files directly and generate image previews via URL
transforms. Rawan banned — accounts are disposable and replaceable.

**Archive Layer (L2)** — Cold, cheap backup storage. GCS Archive class as first
implementation; pluggable for other providers. Every asset gets an archive copy
asynchronously.

**Storage Location** — A record that an asset's bytes exist at a specific account
and path. Tracks `layer` (serving/archive) and `status` (pending/stored/failed).
An asset has ≥1 serving location and (eventually) ≥1 archive location. Failed
locations are kept as audit trail; retries create new rows.

**Account Election** — The strategy that picks which storage account within a
layer receives a new upload. MVP: first-fit (first active account with remaining
capacity).

**Archive Sync Job** — A background task that replicates an asset from serving
layer to archive layer. Async, with exponential backoff retries. Dedicated table.

**Quota** — Storage limit tracked per space (denormalized counter:
`storage_quota_bytes` + `storage_used_bytes`). Incremented on upload, decremented
when last reference removed. Also tracked per storage account for election logic.

---

## Clients

**Web App** — React front-end. Thin client over the API. Auth via opaque session
token (stored in memory/cookie).

**CLI** — Go command-line client. Thin HTTP client, no business logic. Auth via
same opaque token (longer idle TTL: 90 days).

**API** — Go backend (Fiber). All business logic lives here. Source of truth for
rules and orchestration.

---

## Key Design Decisions (quick reference)

| Decision | Choice | ADR |
|----------|--------|-----|
| Storage model | Single physical copy, shared by reference | [ADR-001](./docs/adr/001-single-copy-storage.md) |
| Preview strategy | Provider-side URL transform, images only | [ADR-002](./docs/adr/002-provider-side-preview.md) |
| Tag model | Flat labels (not hierarchical) | [ADR-003](./docs/adr/003-flat-tags.md) |
| Product positioning | Archive-first DAM | [ADR-004](./docs/adr/004-archive-first-dam.md) |
| Auth | Self-managed, opaque tokens + in-process cache | [ADR-005](./docs/adr/005-self-managed-auth.md) |
