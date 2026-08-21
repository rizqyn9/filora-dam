import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/")({
  component: HomePage,
});

function HomePage() {
  // Ticket 04 will replace this with a redirect to the first space.
  return (
    <div className="flex h-full items-center justify-center">
      <p className="text-muted-foreground">Loading spaces...</p>
    </div>
  );
}
