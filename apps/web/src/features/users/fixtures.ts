import type { ManagedUser } from "@/features/users/types";

// Demo data for slicing. Swap for the RBAC users query.
export const users: ManagedUser[] = [
  {
    id: 1,
    name: "Rizqy Nugroho",
    email: "rizqy@filora.app",
    role: "superuser",
    status: "active",
    lastSeen: "2026-07-04T10:20:00Z",
    createdAt: "2026-01-02T09:00:00Z",
  },
  {
    id: 2,
    name: "Sarah Lin",
    email: "sarah@filora.app",
    role: "admin",
    status: "active",
    lastSeen: "2026-07-04T08:02:00Z",
    createdAt: "2026-01-15T09:00:00Z",
  },
  {
    id: 3,
    name: "Diego Martins",
    email: "diego@filora.app",
    role: "member",
    status: "active",
    lastSeen: "2026-07-03T17:40:00Z",
    createdAt: "2026-02-20T09:00:00Z",
  },
  {
    id: 4,
    name: "Amelia Chen",
    email: "amelia@filora.app",
    role: "member",
    status: "invited",
    lastSeen: null,
    createdAt: "2026-06-28T09:00:00Z",
  },
  {
    id: 5,
    name: "Tom Baker",
    email: "tom@filora.app",
    role: "viewer",
    status: "active",
    lastSeen: "2026-07-01T11:12:00Z",
    createdAt: "2026-03-10T09:00:00Z",
  },
  {
    id: 6,
    name: "Priya Nair",
    email: "priya@filora.app",
    role: "viewer",
    status: "suspended",
    lastSeen: "2026-05-19T09:30:00Z",
    createdAt: "2026-04-05T09:00:00Z",
  },
];
