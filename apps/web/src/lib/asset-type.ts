import {
  Archive,
  File,
  FileText,
  Image,
  Video,
  type LucideIcon,
} from "lucide-react";

export type AssetType = "image" | "video" | "document" | "archive" | "file";

interface AssetTypeMeta {
  label: string;
  icon: LucideIcon;
  /** Tailwind classes for a subtle tinted icon chip. */
  className: string;
}

const META: Record<AssetType, AssetTypeMeta> = {
  image: {
    label: "Image",
    icon: Image,
    className: "bg-chart-1/10 text-chart-1",
  },
  video: {
    label: "Video",
    icon: Video,
    className: "bg-chart-2/10 text-chart-2",
  },
  document: {
    label: "Document",
    icon: FileText,
    className: "bg-chart-3/10 text-chart-3",
  },
  archive: {
    label: "Archive",
    icon: Archive,
    className: "bg-chart-4/10 text-chart-4",
  },
  file: { label: "File", icon: File, className: "bg-muted text-foreground" },
};

export function assetTypeMeta(type: string): AssetTypeMeta {
  return META[type as AssetType] ?? META.file;
}
