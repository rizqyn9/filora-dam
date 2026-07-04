# Product Overview

## What is Filora?

Filora is a **multi-cloud Digital Asset Management (DAM)** platform for a family.
People upload and organize their photos, videos, and documents in one place.
Behind the scenes, Filora spreads those files across several cloud storage
accounts and keeps a cheap archived copy — but the user never has to think about
any of that.

> **The core promise:** you manage *your memories and files*, not *storage*.

Users never need to know:

- where a file is physically stored,
- which storage provider is used,
- which account holds their file.

Filora manages all of that automatically.

## The problem

Free-tier cloud media services (Cloudinary, ImageKit, etc.) are generous but
**limited per account**. A single family quickly outgrows one free account, and
juggling multiple accounts by hand is painful:

- Which account has space left?
- Where did I upload that video?
- How do I keep a safe backup without paying for premium tiers?
- How do I share a set of photos with a family member, but not everything?

Filora solves this by treating many storage accounts as **one transparent pool**,
adding a **cheap archive layer** for safety, and layering simple **sharing and
organization** (galleries, albums, tags) on top.

## Who it's for

A **single family** (private, invite-only). Not a public SaaS. This shapes every
decision: small user counts, simple security, low operational overhead, and a
bias toward shipping working features fast.

Typical people:

- **The owner** (superuser) — sets things up, manages storage accounts, invites family.
- **Family admins** — help manage content and members.
- **Members** — upload and organize their own photos/videos.
- **Viewers** — can look and download, but not change things.

See [roles.md](./roles.md) for the full model.

## What you can do (at a glance)

- Upload photos, videos, and documents into a **gallery**.
- Organize them into **albums** and with **tags**.
- **Share** a gallery or album with other family members by email invite.
- Have every file automatically stored on a **serving** layer (fast, viewable)
  and copied to an **archive** layer (cheap, safe backup).
- Manage everything from the **web app** or the **terminal (CLI)** — with multiple
  logged-in terminal sessions at once.

See [features.md](./features.md) for the full catalog.

## Value proposition

| For the user | How Filora delivers it |
|--------------|------------------------|
| "Never run out of space" | Pools multiple free-tier accounts as one; adds accounts as needed |
| "Never lose a file" | Every asset is also copied to a cheap archive layer |
| "Find things fast" | Galleries, albums, tags, search |
| "Share selectively" | Per-gallery / per-album membership (owner/editor/viewer) |
| "Don't make me think about storage" | Provider/account selection is fully automatic and hidden |

## Product principles

1. **Hide the storage complexity.** The user's mental model is galleries, albums,
   and files — never providers or accounts.
2. **Metadata is the truth.** Our database is authoritative; cloud providers are
   just where bytes happen to live.
3. **Safe by default.** Everything gets an archive copy; deletes go to a trash first.
4. **Simple, not enterprise.** Family-scale means we favor the simplest thing that
   works over elaborate, "scalable" machinery.
5. **Build first, refine later.** Ship working features; add abstractions only when
   a second real case appears.

## Non-goals (for now)

- Public sign-up / multi-tenant SaaS.
- Complex enterprise security and compliance.
- Real-time collaboration / editing.
- Microservices, event sourcing, CQRS, and similar heavy architecture.

## Related

- [Domain concepts & glossary](./concepts.md)
- [Feature catalog](./features.md)
- [Roles & permissions](./roles.md)
- [Roadmap & status](./roadmap.md)
- [System architecture](../architecture/README.md)
