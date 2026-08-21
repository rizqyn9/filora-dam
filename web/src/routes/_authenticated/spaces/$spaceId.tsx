import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/spaces/$spaceId")({
  component: SpaceLayout,
});

function SpaceLayout() {
  return <Outlet />;
}
