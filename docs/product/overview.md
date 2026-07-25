# Product Overview

Filora is a multi-cloud Digital Asset Management (DAM) platform for a family.

---

## What is Filora?

People upload and organize their files — photos, videos, documents, archives, anything — in one place. Behind the scenes, Filora spreads those files across several cloud storage accounts and keeps a cheap archived copy. The user never has to think about any of that.

> **The core promise:** you manage *your files*, not *storage*.

Users never need to know:

- where a file is physically stored,
- which storage provider is used,
- which account holds their file.

Filora manages all of that automatically.

## The problem

Free-tier cloud media services (Cloudinary, ImageKit, etc.) are generous but limited per account. A single family quickly outgrows one free account, and juggling multiple accounts by hand is painful:

- Which account has space left?
- Where did I upload that video?
- How do I keep a safe backup without paying for premium tiers?
- How do I share a set of files with a family member, but not everything?

Filora solves this by treating many storage accounts as one transparent pool, adding a cheap archive layer for safety, and layering simple sharing and organization (spaces, folders, tags) on top.

## Who it's for

A single family (private, invite-only). Not a public SaaS. This shapes every decision: small user counts, simple security, low operational overhead, and a bias toward shipping working features fast.

Typical people:

- **The owner** (superuser) — sets things up, manages storage accounts, invites family.
- **Family admins** — help manage content and members.
- **Members** — upload and organize their own files.
- **Viewers** — can look and download, but not change things.

See [roles.md](./roles.md) for the full model.

## What you can do (at a glance)

- Upload any file (photos, videos, documents, archives) into a **space**.
- Organize files into **folders** (hierarchical, nested) and with **tags**.
- **Share** a space with other family members by email invite.
- Have every file automatically stored on a **serving** layer (fast, viewable) and copied to an **archive** layer (cheap, safe backup).
- Browse media (photos/videos) in a grid view optimized for visual content.
- Manage everything from the **web app** or the **terminal (CLI)** — with multiple logged-in terminal sessions at once.

See [features.md](./features.md) for the full catalog.

## Domain model

```
Space (top-level container, one default per user)
├── Folders (hierarchical, nested via parent_id)
│   ├── Files (assets with folder_id set)
│   └── Sub-folders
├── Root files (assets with folder_id = NULL)
└── Tags (cross-cutting virtual grouping)
```

- **Space** — a workspace that groups files, folders, members, and storage quota together. Each user gets a default space on signup.
- **Folder** — hierarchical file organization within a space. Supports unlimited nesting.
- **Asset** — any uploaded file (image, video, document, archive, or generic file). Lives in one space, optionally in one folder.
- **Tag** — a label for cross-cutting grouping. A file can have many tags. Tags are scoped per space.

## Value proposition

| For the user | How Filora delivers it |
|--------------|------------------------|
| "Never run out of space" | Pools multiple free-tier accounts as one; adds accounts as needed |
| "Never lose a file" | Every asset is copied to a cheap archive layer |
| "Find things fast" | Spaces, folders, tags, search, type filters |
| "Share selectively" | Per-space membership (owner/editor/viewer) |
| "Don't make me think about storage" | Provider/account selection is fully automatic and hidden |

## Product principles

1. **Hide the storage complexity.** The user's mental model is spaces, folders, and files — never providers or accounts.
2. **Metadata is the truth.** Our database is authoritative; cloud providers are just where bytes happen to live.
3. **Safe by default.** Everything gets an archive copy; deletes go to a trash first.
4. **Simple, not enterprise.** Family-scale means we favor the simplest thing that works over elaborate machinery.
5. **Build first, refine later.** Ship working features; add abstractions only when a second real case appears.

## Non-goals (for now)

- Public sign-up / multi-tenant SaaS.
- Complex enterprise security and compliance.
- Real-time collaboration / editing.
- Microservices, event sourcing, CQRS, and similar heavy architecture.
- Album/collection features (can be added later if needed).

---

**Next:** [Domain concepts & glossary](./concepts.md) — Definitions and relationships.
