import type {
  DashboardSummary,
  RecentAsset,
  TypeCount,
  UsagePoint,
} from "@/features/dashboard/types";

// Demo data for slicing. Replace with the `/galleries/:id/dashboard` query.

export const summary: DashboardSummary = {
  totalAssets: 12_480,
  totalSize: 348_205_871_104, // ~324 GB
  galleries: 18,
  storageUsed: 348_205_871_104,
  storageQuota: 549_755_813_888, // 512 GB
};

export const usageSeries: UsagePoint[] = [
  { month: "Jan", images: 820, videos: 210, documents: 140 },
  { month: "Feb", images: 932, videos: 268, documents: 180 },
  { month: "Mar", images: 1010, videos: 290, documents: 176 },
  { month: "Apr", images: 1180, videos: 340, documents: 220 },
  { month: "May", images: 1290, videos: 410, documents: 260 },
  { month: "Jun", images: 1480, videos: 520, documents: 310 },
  { month: "Jul", images: 1610, videos: 610, documents: 352 },
];

export const typeCounts: TypeCount[] = [
  { type: "image", count: 8420 },
  { type: "video", count: 2360 },
  { type: "document", count: 1240 },
  { type: "archive", count: 320 },
  { type: "file", count: 140 },
];

export const recentAssets: RecentAsset[] = [
  {
    id: "1",
    name: "campaign-hero-2024.jpg",
    type: "image",
    size: 4_582_400,
    createdAt: "2026-07-04T10:22:00Z",
  },
  {
    id: "2",
    name: "product-launch.mp4",
    type: "video",
    size: 128_450_000,
    createdAt: "2026-07-04T09:10:00Z",
  },
  {
    id: "3",
    name: "brand-guidelines.pdf",
    type: "document",
    size: 2_340_000,
    createdAt: "2026-07-03T18:44:00Z",
  },
  {
    id: "4",
    name: "team-offsite-batch.zip",
    type: "archive",
    size: 512_000_000,
    createdAt: "2026-07-03T14:02:00Z",
  },
  {
    id: "5",
    name: "logo-mark-white.png",
    type: "image",
    size: 128_400,
    createdAt: "2026-07-03T11:30:00Z",
  },
];
