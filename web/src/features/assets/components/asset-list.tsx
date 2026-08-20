import { FileText, Archive, File } from "lucide-react";

import type { Asset } from "@/features/assets/schemas";
import { assetTypeMeta } from "@/lib/asset-type";
import { formatBytes, formatDate } from "@/lib/format";

interface AssetListProps {
  assets: Asset[];
  onSelect?: (asset: Asset) => void;
}

/**
 * Table/list view for non-media files (documents, archives, generic files).
 * Used as the default view or when browsing folders with mixed content.
 */
export function AssetList({ assets, onSelect }: AssetListProps) {
  if (assets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <File className="size-12 mb-3 opacity-50" />
        <p>No files here yet.</p>
      </div>
    );
  }

  return (
    <div className="divide-y rounded-lg border">
      {assets.map((asset) => {
        const meta = assetTypeMeta(asset.type);
        const Icon = meta.icon;

        return (
          <button
            key={asset.id}
            type="button"
            className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            onClick={() => onSelect?.(asset)}
            aria-label={`View ${asset.name}`}
          >
            <div className={`flex size-9 items-center justify-center rounded-md ${meta.className}`}>
              <Icon className="size-4" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{asset.name}</p>
              <p className="text-xs text-muted-foreground">
                {meta.label} &middot; {formatBytes(asset.size)}
              </p>
            </div>
            <span className="text-xs text-muted-foreground">
              {formatDate(asset.created_at)}
            </span>
          </button>
        );
      })}
    </div>
  );
}
