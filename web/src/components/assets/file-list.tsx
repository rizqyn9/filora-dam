import {
  File,
  FileArchive,
  FileAudio,
  FileImage,
  FileText,
  FileVideo,
  MoreVertical,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import type { Asset } from "@/features/assets/schemas";
import { formatBytes, formatDate } from "@/lib/format";

function getMimeIcon(mime: string) {
  if (mime.startsWith("image/")) return FileImage;
  if (mime.startsWith("video/")) return FileVideo;
  if (mime.startsWith("audio/")) return FileAudio;
  if (mime.startsWith("text/") || mime.includes("pdf") || mime.includes("document"))
    return FileText;
  if (mime.includes("zip") || mime.includes("tar") || mime.includes("archive"))
    return FileArchive;
  return File;
}

export function FileList({ assets }: { assets: Asset[] }) {
  return (
    <div className="divide-y">
      {assets.map((asset) => {
        const Icon = getMimeIcon(asset.mime_type);
        return (
          <div
            key={asset.id}
            className="group flex items-center gap-3 px-4 py-2 transition-colors hover:bg-accent/50"
          >
            <Icon className="size-5 shrink-0 text-muted-foreground" />
            <span className="min-w-0 flex-1 truncate text-sm">{asset.name}</span>
            <span className="hidden shrink-0 text-xs text-muted-foreground sm:block">
              {formatBytes(asset.size_bytes)}
            </span>
            <span className="hidden shrink-0 text-xs text-muted-foreground md:block">
              {formatDate(asset.created_at)}
            </span>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 shrink-0 opacity-0 group-hover:opacity-100"
              aria-label="Actions"
            >
              <MoreVertical className="size-4" />
            </Button>
          </div>
        );
      })}
    </div>
  );
}

export function FileListSkeleton() {
  return (
    <div className="divide-y">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 px-4 py-2">
          <Skeleton className="size-5 rounded" />
          <Skeleton className="h-4 flex-1" />
          <Skeleton className="hidden h-3 w-16 sm:block" />
        </div>
      ))}
    </div>
  );
}
