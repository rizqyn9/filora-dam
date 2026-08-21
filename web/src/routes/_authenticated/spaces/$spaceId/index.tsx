import { createFileRoute, useParams } from "@tanstack/react-router";

import { AssetBrowser } from "@/components/assets/asset-browser";

export const Route = createFileRoute("/_authenticated/spaces/$spaceId/")({
  component: SpaceRootPage,
});

function SpaceRootPage() {
  const { spaceId } = useParams({
    from: "/_authenticated/spaces/$spaceId/",
  });

  return <AssetBrowser spaceId={spaceId} />;
}
