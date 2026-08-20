# Product Overview

Filora is a private, archive-first Digital Asset Management platform for a family.

---

## What is Filora?

People upload any file — photos, videos, documents, archives — into their own
space. Behind the scenes, Filora stores each file once across a pool of free-tier
cloud accounts and keeps a cheap archived copy on a separate layer. The user
never has to think about any of that.

> **The core promise:** your files are safe, organized, and always findable —
> you never manage storage.

Users never need to know:

- where a file is physically stored,
- which storage provider holds it,
- which account among dozens it landed on.

Filora manages all of that automatically.

## The problem

Free-tier cloud media services (Cloudinary, ImageKit, etc.) are generous but
limited per account. A single family quickly outgrows one free account, and
juggling multiple accounts by hand is painful:

- Which account has space left?
- Where did I upload that video?
- How do I keep a safe backup without paying for premium tiers?
- How do I share a set of files with a family member, but not everything?

Filora solves this by pooling many storage accounts as one transparent layer,
adding a cheap archive layer for safety, and layering simple organization
(spaces, folders, tags) on top.

## Who it's for

A single family (private, invite-only). Not a public SaaS. This shapes every
decision: small user counts (3–5 people), simple security, low operational
overhead, and a bias toward shipping working features fast.

Typical people:

- **The owner** (superuser) — sets things up, manages storage accounts, invites family.
- **Family admins** — help manage content and members.
- **Members** — upload and organize their own files.
- **Viewers** — can look and download, but not change things.

See [roles.md](./roles.md) for the full model.

## What you can do (at a glance)

- Upload any file into your **space**.
- Organize files into **folders** (hierarchical, nested) and with **tags** (flat labels).
- Place one file in **multiple folders** (virtual reference, not a copy).
- **Share** a space with family members by invite (editor or viewer access).
- Have every file stored on a **serving layer** (fast, free-tier) and copied to an
  **archive layer** (cheap, safe backup) asynchronously.
- Browse images with provider-generated **previews**; non-image files show type icons.
- Manage everything from the **web app** or the **CLI**.

See [features.md](./features.md) for the full catalog.

## Domain model

```
Space (top-level container, ≥1 per user)
├── Folders (hierarchical, nested via parent_id)
│   └── Asset References (many-to-many; same asset in multiple folders)
├── Root assets (references with no folder)
└── Tags (flat labels, scoped per space)

Asset (single physical copy)
├── Storage Location: Serving Layer (L1)
└── Storage Location: Archive Layer (L2)
```

- **Space** — a workspace that groups assets, folders, members, and quota.
  Each user gets a default space on signup. A user can own multiple spaces.
- **Folder** — hierarchical organization within a space. Unlimited nesting.
- **Asset** — a logical file record plus metadata. One physical copy, shared by
  reference across spaces and folders.
- **Asset Reference** — the link between an asset and a folder/space. Many-to-many.
- **Tag** — a flat label for cross-cutting grouping, scoped per space.

See [concepts.md](./concepts.md) for the full glossary.

## Value proposition

| For the user | How Filora delivers it |
|--------------|------------------------|
| "Never lose a file" | Every asset is archived to a cheap backup layer automatically |
| "Never run out of space" | Pools many free-tier accounts; single-copy storage maximizes capacity |
| "Find things fast" | Spaces, folders, tags, search, type filters |
| "Share selectively" | Per-space membership (owner/editor/viewer) |
| "Don't make me think about storage" | Provider/account selection is fully automatic and hidden |

## Product principles

1. **Archive-first.** Data safety is the core value. Every asset reaches both
   layers. Organization is the UX that makes the archive usable.
2. **Hide the storage complexity.** The user's mental model is spaces, folders,
   and files — never providers or accounts.
3. **Metadata is the truth.** PostgreSQL is authoritative; cloud providers are
   just where bytes happen to live.
4. **Safe by default.** Soft deletes, trash retention, archive redundancy.
5. **Simple, not enterprise.** Family-scale means the simplest thing that works.
6. **Build first, refine later.** Ship working features; add abstractions only
   when a second real case appears.

## Non-goals (for now)

- Public sign-up / multi-tenant SaaS.
- Complex enterprise security and compliance.
- Real-time collaboration / editing.
- Mobile app or auto-sync from devices.
- AI-powered tagging, smart albums, face detection.
- Video/PDF preview generation (server-side compute).
- Microservices, event sourcing, CQRS.

---

**Next:** [Domain concepts & glossary](./concepts.md) — Definitions and relationships.
