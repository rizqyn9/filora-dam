import { createFileRoute, Navigate } from "@tanstack/react-router";

import { useSpaces } from "@/features/spaces/api";

export const Route = createFileRoute("/_authenticated/")({
  component: RedirectToFirstSpace,
});

function RedirectToFirstSpace() {
  const { data: spaces, isLoading } = useSpaces();

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  if (spaces?.length) {
    return (
      <Navigate to="/spaces/$spaceId" params={{ spaceId: spaces[0].id }} />
    );
  }

  return (
    <div className="flex h-full items-center justify-center">
      <p className="text-muted-foreground">
        No spaces yet. Create one to get started.
      </p>
    </div>
  );
}
