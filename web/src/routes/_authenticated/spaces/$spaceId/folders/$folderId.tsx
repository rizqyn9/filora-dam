import { createFileRoute, useParams } from "@tanstack/react-router";

import { AssetBrowser } from "@/components/assets/asset-browser";

export const Route = createFileRoute(
  "/_authenticated/spaces/$spaceId/folders/$folderId",
)({
  component: FolderPage,
});

function FolderPage() {
  const { spaceId, folderId } = useParams({
    from: "/_authenticated/spaces/$spaceId/folders/$folderId",
  });

  return <AssetBrowser spaceId={spaceId} folderId={folderId} />;
}
