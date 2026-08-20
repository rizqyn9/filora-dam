export type StorageLayer = "serving" | "archive";
export type StorageProviderType = "cloudinary" | "imagekit" | "r2" | "gcs";

export interface StorageAccount {
  id: number;
  name: string;
  layer: StorageLayer;
  type: StorageProviderType;
  isActive: boolean;
  quota: number | null;
  used: number;
  usedPercent: number;
  storedCount: number;
  pendingCount: number;
  failedCount: number;
}

export const providerLabels: Record<StorageProviderType, string> = {
  cloudinary: "Cloudinary",
  imagekit: "ImageKit",
  r2: "Cloudflare R2",
  gcs: "Google Cloud Storage",
};
