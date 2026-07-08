import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { GalleryDeleteDialog } from "@/features/galleries/components/gallery-delete-dialog";
import { GalleryFormDialog } from "@/features/galleries/components/gallery-form-dialog";
import type { Gallery } from "@/features/galleries/schemas";
import { formatBytes, formatDate } from "@/lib/format";

function GalleryRowActions({ gallery }: { gallery: Gallery }) {
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  return (
    <>
      <RowActions
        actions={[
          {
            label: "Edit",
            icon: Pencil,
            onSelect: () => setEditOpen(true),
          },
          {
            label: "Delete",
            icon: Trash2,
            destructive: true,
            separatorBefore: true,
            onSelect: () =>
              gallery.is_default
                ? toast.error("The default gallery cannot be deleted")
                : setDeleteOpen(true),
          },
        ]}
      />
      <GalleryFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        gallery={gallery}
      />
      <GalleryDeleteDialog
        gallery={gallery}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </>
  );
}

export const galleryColumns: ColumnDef<Gallery>[] = [
  {
    accessorKey: "name",
    header: ({ column }) => (
      <Button
        variant="ghost"
        className="-ml-3"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Name
        <ArrowUpDown className="ml-2 size-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <div className="flex items-center gap-2 font-medium">
        {row.original.name}
        {row.original.is_default && <Badge variant="secondary">Default</Badge>}
      </div>
    ),
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.original.description ?? "—"}
      </span>
    ),
  },
  {
    id: "usage",
    header: "Usage",
    cell: ({ row }) => {
      const { storage_used, storage_quota } = row.original;
      return (
        <span className="tabular-nums">
          {formatBytes(storage_used)} / {formatBytes(storage_quota)}
        </span>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => formatDate(row.original.created_at),
  },
  {
    id: "actions",
    cell: ({ row }) => (
      <div className="text-right">
        <GalleryRowActions gallery={row.original} />
      </div>
    ),
  },
];
