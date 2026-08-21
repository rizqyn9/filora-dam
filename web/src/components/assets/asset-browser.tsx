import { useCallback, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { useAssets } from "@/features/assets/api";
import { useUploadFiles } from "@/features/assets/use-upload-files";
import { useSpaces } from "@/features/spaces/api";
import { useUiStore } from "@/stores/ui-store";

import { ContentToolbar } from "./content-toolbar";
import { FileGrid, FileGridSkeleton } from "./file-grid";
import { FileList, FileListSkeleton } from "./file-list";
import { UploadDropzone } from "./upload-dropzone";

const PAGE_SIZE = 50;

interface AssetBrowserProps {
  spaceId: string;
  folderId?: string;
}

export function AssetBrowser({ spaceId, folderId }: AssetBrowserProps) {
  const viewMode = useUiStore((s) => s.viewMode);
  const [offset, setOffset] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data: spaces } = useSpaces();
  const spaceName = spaces?.find((s) => s.id === spaceId)?.name;

  const { data: assets, isLoading } = useAssets({
    spaceId,
    folderId,
    limit: PAGE_SIZE,
    offset,
  });

  const uploadFiles = useUploadFiles({ spaceId, folderId });

  const handleUploadClick = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files ?? []);
      if (files.length) uploadFiles(files);
      e.target.value = ""; // reset so same file can be re-selected
    },
    [uploadFiles],
  );

  return (
    <div className="flex h-full flex-col">
      <ContentToolbar
        spaceId={spaceId}
        spaceName={spaceName}
        folderId={folderId}
        onUploadClick={handleUploadClick}
      />

      <UploadDropzone onDrop={uploadFiles}>
        {/* Content */}
        <div className="h-full overflow-auto">
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
      </UploadDropzone>

      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={handleFileChange}
      />
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
