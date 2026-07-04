import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { Gallery } from "@/features/galleries/schemas";
import { formatBytes, formatDate } from "@/lib/format";

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
];
