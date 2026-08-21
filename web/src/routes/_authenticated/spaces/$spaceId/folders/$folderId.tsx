import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/spaces/$spaceId/folders/$folderId",
)({
  component: FolderPage,
});

function FolderPage() {
  // Asset browser wired in ticket 06
  return (
    <div className="flex h-full items-center justify-center">
      <p className="text-muted-foreground">Folder contents — assets will appear here</p>
    </div>
  );
}
