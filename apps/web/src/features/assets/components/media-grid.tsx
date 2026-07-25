import { Image, Play } from "lucide-react";

import type { Asset } from "@/features/assets/schemas";
import { formatBytes } from "@/lib/format";

interface MediaGridProps {
  assets: Asset[];
  onSelect?: (asset: Asset) => void;
}

/**
 * Grid/masonry-style view optimized for media (images & videos).
 * Displays thumbnails in a responsive grid with overlay metadata.
 * Used when filtering by type=image or type=video.
 */
export function MediaGrid({ assets, onSelect }: MediaGridProps) {
  if (assets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Image className="size-12 mb-3 opacity-50" />
        <p>No media files found.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {assets.map((asset) => (
        <MediaGridItem
          key={asset.id}
          asset={asset}
          onClick={() => onSelect?.(asset)}
        />
      ))}
    </div>
  );
}

interface MediaGridItemProps {
  asset: Asset;
  onClick?: () => void;
}

function MediaGridItem({ asset, onClick }: MediaGridItemProps) {
  const isVideo = asset.type === "video";

  return (
    <button
      type="button"
      className="group relative aspect-square overflow-hidden rounded-lg bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      onClick={onClick}
      aria-label={`View ${asset.name}`}
    >
      {/* Placeholder — real thumbnail URL comes from storage serving layer */}
      <div className="flex size-full items-center justify-center bg-muted">
        {isVideo ? (
          <Play className="size-8 text-muted-foreground" />
        ) : (
          <Image className="size-8 text-muted-foreground" />
        )}
      </div>

      {/* Overlay on hover */}
      <div className="absolute inset-x-0 bottom-0 translate-y-full bg-gradient-to-t from-black/70 to-transparent p-2 transition-transform group-hover:translate-y-0">
        <p className="truncate text-xs font-medium text-white">{asset.name}</p>
        <p className="text-xs text-white/70">{formatBytes(asset.size)}</p>
      </div>

      {/* Video badge */}
      {isVideo && (
        <div className="absolute right-1.5 top-1.5 rounded bg-black/60 px-1 py-0.5 text-[10px] font-medium text-white">
          VIDEO
        </div>
      )}
    </button>
  );
}
