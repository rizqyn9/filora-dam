import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Shield, Trash2, UserCog } from "lucide-react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { ManagedUser, UserStatus } from "@/features/users/types";
import { formatDate } from "@/lib/format";
import { cn, getInitials } from "@/lib/utils";

const statusStyles: Record<UserStatus, string> = {
  active: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
  invited: "bg-amber-500/10 text-amber-600 border-amber-500/20",
  suspended: "bg-destructive/10 text-destructive border-destructive/20",
};

export const userColumns: ColumnDef<ManagedUser>[] = [
  {
    accessorKey: "name",
    header: ({ column }) => (
      <Button
        variant="ghost"
        className="-ml-3"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        User
        <ArrowUpDown className="ml-2 size-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <Avatar className="size-8">
          <AvatarFallback className="text-xs">
            {getInitials(row.original.name)}
          </AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <p className="truncate text-sm font-medium">{row.original.name}</p>
          <p className="truncate text-xs text-muted-foreground">
            {row.original.email}
          </p>
        </div>
      </div>
    ),
  },
  {
    accessorKey: "role",
    header: "Role",
    cell: ({ row }) => (
      <Badge variant="secondary" className="gap-1 font-normal capitalize">
        <Shield className="size-3" />
        {row.original.role}
      </Badge>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge
        variant="outline"
        className={cn("capitalize", statusStyles[row.original.status])}
      >
        {row.original.status}
      </Badge>
    ),
  },
  {
    accessorKey: "lastSeen",
    header: "Last seen",
    cell: ({ row }) =>
      row.original.lastSeen ? (
        formatDate(row.original.lastSeen)
      ) : (
        <span className="text-muted-foreground">—</span>
      ),
  },
  {
    id: "actions",
    cell: ({ row }) => (
      <div className="text-right">
        <RowActions
          actions={[
            {
              label: "Change role",
              icon: UserCog,
              onSelect: () =>
                toast.info(`Change role for ${row.original.name}`),
            },
            {
              label: "Remove",
              icon: Trash2,
              destructive: true,
              separatorBefore: true,
              onSelect: () => toast.success(`${row.original.name} removed`),
            },
          ]}
        />
      </div>
    ),
  },
];
