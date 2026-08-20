import { createFileRoute } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useSpaces } from "@/features/spaces/api";
import { spaceColumns } from "@/features/spaces/components/space-columns";
import { SpaceFormDialog } from "@/features/spaces/components/space-form-dialog";
import { ApiError } from "@/lib/api-client";

export const Route = createFileRoute("/_app/spaces")({
  component: SpacesPage,
});

function SpacesPage() {
  const { data, isPending, isError, error } = useSpaces();

  const newButton = (
    <SpaceFormDialog
      trigger={
        <Button size="sm">
          <Plus className="size-4" />
          New space
        </Button>
      }
    />
  );

  return (
    <>
      <PageHeader
        title="Spaces"
        description="All spaces you have access to."
        actions={newButton}
      />

      {isPending ? (
        <div className="space-y-3">
          <Skeleton className="h-9 w-full max-w-xs" />
          <Skeleton className="h-64 w-full rounded-lg" />
        </div>
      ) : isError ? (
        <Alert variant="destructive">
          <AlertTitle>Failed to load spaces</AlertTitle>
          <AlertDescription>
            {error instanceof ApiError
              ? error.message
              : "Something went wrong. Please try again."}
          </AlertDescription>
        </Alert>
      ) : (
        <DataTable
          columns={spaceColumns}
          data={data}
          searchColumn="name"
          searchPlaceholder="Search spaces..."
          emptyMessage="No spaces yet."
        />
      )}
    </>
  );
}
