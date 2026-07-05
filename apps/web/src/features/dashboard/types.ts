/** Presentational types for the dashboard slice (mirror the API dashboard module). */

export interface UsagePoint {
  month: string;
  images: number;
  videos: number;
  documents: number;
}

export interface TypeCount {
  type: string;
  count: number;
}

export interface RecentAsset {
  id: string;
  name: string;
  type: "image" | "video" | "document" | "archive" | "file";
  size: number;
  createdAt: string;
}

export interface DashboardSummary {
  totalAssets: number;
  totalSize: number;
  galleries: number;
  storageUsed: number;
  storageQuota: number;
}
