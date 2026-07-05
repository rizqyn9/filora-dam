import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Pencil, Tag as TagIcon, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { TagFormDialog } from "@/features/tags/components/tag-form-dialog";
import type { Tag } from "@/features/tags/types";
import { formatDate } from "@/lib/format";

function TagRowActions({ tag }: { tag: Tag }) {
  const [editOpen, setEditOpen] = useState(false);

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
            onSelect: () => toast.success(`Tag "${tag.name}" deleted`),
          },
        ]}
      />
      <TagFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        initialName={tag.name}
      />
    </>
  );
}

export const tagColumns: ColumnDef<Tag>[] = [
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
      <Badge variant="secondary" className="gap-1 font-normal">
        <TagIcon className="size-3" />
        {row.original.name}
      </Badge>
    ),
  },
  {
    accessorKey: "galleryName",
    header: "Gallery",
    cell: ({ row }) => (
      <span className="text-muted-foreground">{row.original.galleryName}</span>
    ),
  },
  {
    accessorKey: "assetCount",
    header: () => <div className="text-right">Assets</div>,
    cell: ({ row }) => (
      <div className="text-right tabular-nums">
        {row.original.assetCount.toLocaleString()}
      </div>
    ),
  },
  {
    accessorKey: "createdAt",
    header: "Created",
    cell: ({ row }) => formatDate(row.original.createdAt),
  },
  {
    id: "actions",
    cell: ({ row }) => (
      <div className="text-right">
        <TagRowActions tag={row.original} />
      </div>
    ),
  },
];
