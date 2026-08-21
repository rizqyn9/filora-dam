import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/spaces/$spaceId/",
)({
  component: SpaceRootPage,
});

function SpaceRootPage() {
  // Asset browser wired in ticket 06
  return (
    <div className="flex h-full items-center justify-center">
      <p className="text-muted-foreground">Space root — assets will appear here</p>
    </div>
  );
}
