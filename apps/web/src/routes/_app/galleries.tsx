import { createFileRoute } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { galleryColumns } from "@/features/galleries/components/gallery-columns";
import { galleries } from "@/features/galleries/fixtures";

export const Route = createFileRoute("/_app/galleries")({
  component: GalleriesPage,
});

function GalleriesPage() {
  return (
    <>
      <PageHeader
        title="Galleries"
        description="All galleries you have access to."
      />
      <DataTable
        columns={galleryColumns}
        data={galleries}
        searchColumn="name"
        searchPlaceholder="Search galleries..."
        emptyMessage="No galleries yet."
        toolbar={
          <Button size="sm">
            <Plus className="size-4" />
            New gallery
          </Button>
        }
      />
    </>
  );
}
