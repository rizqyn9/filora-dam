import { createFileRoute } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useGalleries } from "@/features/galleries/api";
import { galleryColumns } from "@/features/galleries/components/gallery-columns";
import { GalleryFormDialog } from "@/features/galleries/components/gallery-form-dialog";
import { ApiError } from "@/lib/api-client";

export const Route = createFileRoute("/_app/galleries")({
  component: GalleriesPage,
});

function GalleriesPage() {
  const { data, isPending, isError, error } = useGalleries();

  const newButton = (
    <GalleryFormDialog
      trigger={
        <Button size="sm">
          <Plus className="size-4" />
          New gallery
        </Button>
      }
    />
  );

  return (
    <>
      <PageHeader
        title="Galleries"
        description="All galleries you have access to."
        actions={newButton}
      />

      {isPending ? (
        <div className="space-y-3">
          <Skeleton className="h-9 w-full max-w-xs" />
          <Skeleton className="h-64 w-full rounded-lg" />
        </div>
      ) : isError ? (
        <Alert variant="destructive">
          <AlertTitle>Failed to load galleries</AlertTitle>
          <AlertDescription>
            {error instanceof ApiError
              ? error.message
              : "Something went wrong. Please try again."}
          </AlertDescription>
        </Alert>
      ) : (
        <DataTable
          columns={galleryColumns}
          data={data}
          searchColumn="name"
          searchPlaceholder="Search galleries..."
          emptyMessage="No galleries yet."
        />
      )}
    </>
  );
}
