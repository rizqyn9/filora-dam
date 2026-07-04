import { createFileRoute } from "@tanstack/react-router";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export const Route = createFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Welcome to Filora</h1>
        <p className="text-muted-foreground">
          Multi-cloud digital asset management.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Galleries</CardTitle>
            <CardDescription>Organize assets into galleries.</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Browse and manage your galleries from the top navigation.
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
