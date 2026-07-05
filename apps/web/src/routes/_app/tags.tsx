import { createFileRoute } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { tagColumns } from "@/features/tags/components/tag-columns";
import { TagFormDialog } from "@/features/tags/components/tag-form-dialog";
import { tags } from "@/features/tags/fixtures";

export const Route = createFileRoute("/_app/tags")({
  component: TagsPage,
});

function TagsPage() {
  return (
    <>
      <PageHeader
        title="Tags"
        description="Manage tags used to organize assets across galleries."
      />
      <DataTable
        columns={tagColumns}
        data={tags}
        searchColumn="name"
        searchPlaceholder="Search tags..."
        emptyMessage="No tags yet."
        toolbar={
          <TagFormDialog
            trigger={
              <Button size="sm">
                <Plus className="size-4" />
                New tag
              </Button>
            }
          />
        }
      />
    </>
  );
}
