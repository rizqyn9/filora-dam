import { createFileRoute } from "@tanstack/react-router";

import { DataTable } from "@/components/data-table";
import { Skeleton } from "@/components/ui/skeleton";
import { useGalleries, galleriesQueryOptions } from "@/features/galleries/api";
import { galleryColumns } from "@/features/galleries/components/gallery-columns";

export const Route = createFileRoute("/galleries/")({
  // Prefetch on navigation using the router's queryClient context.
  loader: ({ context }) =>
    context.queryClient.ensureQueryData(galleriesQueryOptions()),
  component: GalleriesPage,
});

function GalleriesPage() {
  const { data, isPending, isError, error } = useGalleries();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Galleries</h1>
        <p className="text-muted-foreground">
          All galleries you have access to.
        </p>
      </div>

      {isPending ? (
        <div className="space-y-2">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : isError ? (
        <p className="text-sm text-destructive">{error.message}</p>
      ) : (
        <DataTable
          columns={galleryColumns}
          data={data}
          emptyMessage="No galleries yet."
        />
      )}
    </div>
  );
}
