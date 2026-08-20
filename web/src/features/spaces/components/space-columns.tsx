import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SpaceDeleteDialog } from "@/features/spaces/components/space-delete-dialog";
import { SpaceFormDialog } from "@/features/spaces/components/space-form-dialog";
import type { Space } from "@/features/spaces/schemas";
import { formatBytes, formatDate } from "@/lib/format";

function SpaceRowActions({ space }: { space: Space }) {
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
              space.is_default
                ? toast.error("The default space cannot be deleted")
                : setDeleteOpen(true),
          },
        ]}
      />
      <SpaceFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        space={space}
      />
      <SpaceDeleteDialog
        space={space}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </>
  );
}

export const spaceColumns: ColumnDef<Space>[] = [
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
        {row.original.description ?? "\u2014"}
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
        <SpaceRowActions space={row.original} />
      </div>
    ),
  },
];
