import type { Role } from "@/features/roles/types";

// Demo data for slicing. Swap for the `/rbac/roles` query.
export const roles: Role[] = [
  {
    id: 1,
    slug: "superuser",
    name: "Superuser",
    description: "Full, unrestricted access to everything.",
    isSystem: true,
    memberCount: 1,
    permissions: ["*:*"],
  },
  {
    id: 2,
    slug: "admin",
    name: "Admin",
    description: "Workspace-wide management across all galleries.",
    isSystem: true,
    memberCount: 2,
    permissions: [
      "gallery:*",
      "asset:*",
      "storage:*",
      "role:read",
      "user:manage",
    ],
  },
  {
    id: 3,
    slug: "member",
    name: "Member",
    description: "Create and manage their own galleries and assets.",
    isSystem: true,
    memberCount: 8,
    permissions: ["gallery:create", "asset:create", "asset:read", "tag:*"],
  },
  {
    id: 4,
    slug: "viewer",
    name: "Viewer",
    description: "Read-only access to shared galleries.",
    isSystem: true,
    memberCount: 14,
    permissions: ["gallery:read", "asset:read"],
  },
  {
    id: 5,
    slug: "editor",
    name: "Content Editor",
    description: "Custom role for the marketing team.",
    isSystem: false,
    memberCount: 5,
    permissions: ["asset:create", "asset:update", "tag:create", "gallery:read"],
  },
];
