import type { Asset } from "@/features/assets/schemas";
import { AssetList } from "@/features/assets/components/asset-list";
import { MediaGrid } from "@/features/assets/components/media-grid";

interface AssetViewProps {
  assets: Asset[];
  /** Active type filter — when "image" or "video", uses grid view. */
  typeFilter?: string | null;
  onSelect?: (asset: Asset) => void;
}

/**
 * Smart asset view that switches between grid (media) and list (files) layout
 * based on the active filter or content composition.
 *
 * - If typeFilter is "image" or "video" → MediaGrid (masonry-like)
 * - Otherwise → AssetList (table-like with icons)
 */
export function AssetView({ assets, typeFilter, onSelect }: AssetViewProps) {
  const useMediaView = typeFilter === "image" || typeFilter === "video";

  if (useMediaView) {
    return <MediaGrid assets={assets} onSelect={onSelect} />;
  }

  return <AssetList assets={assets} onSelect={onSelect} />;
}
