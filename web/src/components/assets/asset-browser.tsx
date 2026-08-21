import { LayoutGrid, List, Upload } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useAssets } from "@/features/assets/api";
import { useUiStore } from "@/stores/ui-store";

import { FileGrid, FileGridSkeleton } from "./file-grid";
import { FileList, FileListSkeleton } from "./file-list";

const PAGE_SIZE = 50;

interface AssetBrowserProps {
  spaceId: string;
  folderId?: string;
}

export function AssetBrowser({ spaceId, folderId }: AssetBrowserProps) {
  const viewMode = useUiStore((s) => s.viewMode);
  const setViewMode = useUiStore((s) => s.setViewMode);
  const [offset, setOffset] = useState(0);

  const { data: assets, isLoading } = useAssets({
    spaceId,
    folderId,
    limit: PAGE_SIZE,
    offset,
  });

  return (
    <div className="flex h-full flex-col">
      {/* Toolbar */}
      <div className="flex items-center gap-2 border-b px-4 py-2">
        <div className="flex-1" />
        <ToggleGroup
          value={[viewMode]}
          onValueChange={(v) => {
            if (v.length) setViewMode(v[0] as "grid" | "list");
          }}
          size="sm"
        >
          <ToggleGroupItem value="grid" aria-label="Grid view">
            <LayoutGrid className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem value="list" aria-label="List view">
            <List className="size-4" />
          </ToggleGroupItem>
        </ToggleGroup>
        <Button size="sm" variant="outline" disabled>
          <Upload className="mr-1.5 size-4" />
          Upload
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {isLoading ? (
          viewMode === "grid" ? (
            <FileGridSkeleton />
          ) : (
            <FileListSkeleton />
          )
        ) : !assets?.length ? (
          <EmptyState />
        ) : viewMode === "grid" ? (
          <FileGrid assets={assets} />
        ) : (
          <FileList assets={assets} />
        )}

        {/* Load more */}
        {assets && assets.length === PAGE_SIZE && (
          <div className="flex justify-center pb-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setOffset((o) => o + PAGE_SIZE)}
            >
              Load more
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 p-8">
      <p className="text-muted-foreground">This folder is empty</p>
      <p className="text-sm text-muted-foreground">
        Upload files or create a folder to get started.
      </p>
    </div>
  );
}
