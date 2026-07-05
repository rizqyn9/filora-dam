import type { Gallery } from "@/features/galleries/schemas";

// Demo data for slicing. Swap for `useGalleries()` once the API is wired.
export const galleries: Gallery[] = [
  {
    id: 1,
    owner_id: 1,
    name: "Marketing 2026",
    description: "Campaign creative and brand assets",
    is_default: true,
    storage_quota: 107_374_182_400,
    storage_used: 64_424_509_440,
    created_at: "2026-01-12T09:00:00Z",
    updated_at: "2026-06-30T12:00:00Z",
  },
  {
    id: 2,
    owner_id: 1,
    name: "Product Shots",
    description: "E-commerce product photography",
    is_default: false,
    storage_quota: 53_687_091_200,
    storage_used: 41_231_686_042,
    created_at: "2026-02-03T09:00:00Z",
    updated_at: "2026-06-28T12:00:00Z",
  },
  {
    id: 3,
    owner_id: 2,
    name: "Events",
    description: "Conferences, offsites and team photos",
    is_default: false,
    storage_quota: 26_843_545_600,
    storage_used: 9_663_676_416,
    created_at: "2026-03-21T09:00:00Z",
    updated_at: "2026-06-25T12:00:00Z",
  },
  {
    id: 4,
    owner_id: 3,
    name: "Legal & Docs",
    description: null,
    is_default: false,
    storage_quota: 10_737_418_240,
    storage_used: 1_073_741_824,
    created_at: "2026-04-08T09:00:00Z",
    updated_at: "2026-06-20T12:00:00Z",
  },
];
