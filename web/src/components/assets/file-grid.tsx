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
import { formatBytes } from "@/lib/format";

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

export function FileGrid({ assets }: { assets: Asset[] }) {
  return (
    <div className="grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
      {assets.map((asset) => {
        const Icon = getMimeIcon(asset.mime_type);
        return (
          <div
            key={asset.id}
            className="group relative flex flex-col items-center gap-2 rounded-lg border p-3 transition-colors hover:bg-accent/50"
          >
            <Icon className="size-10 text-muted-foreground" />
            <div className="w-full min-w-0 text-center">
              <p className="truncate text-sm font-medium">{asset.name}</p>
              <p className="text-xs text-muted-foreground">
                {formatBytes(asset.size_bytes)}
              </p>
            </div>
            {/* Kebab menu placeholder — wired in ticket 13/14 */}
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-1 top-1 size-7 opacity-0 group-hover:opacity-100"
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

export function FileGridSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex flex-col items-center gap-2 rounded-lg border p-3">
          <Skeleton className="size-10 rounded" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
        </div>
      ))}
    </div>
  );
}
