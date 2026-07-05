import { Lock, Pencil, Trash2, Users } from "lucide-react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Role } from "@/features/roles/types";

export function RoleCard({ role }: { role: Role }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          {role.name}
          {role.isSystem && (
            <Badge variant="outline" className="gap-1 font-normal">
              <Lock className="size-3" />
              System
            </Badge>
          )}
        </CardTitle>
        <CardDescription>{role.description ?? "—"}</CardDescription>
        <div className="ml-auto">
          <RowActions
            actions={[
              {
                label: "Edit",
                icon: Pencil,
                onSelect: () => toast.info(`Edit ${role.name}`),
              },
              {
                label: "Delete",
                icon: Trash2,
                destructive: true,
                separatorBefore: true,
                onSelect: () =>
                  role.isSystem
                    ? toast.error("System roles cannot be deleted")
                    : toast.success(`${role.name} deleted`),
              },
            ]}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex flex-wrap gap-1.5">
          {role.permissions.map((perm) => (
            <Badge
              key={perm}
              variant="secondary"
              className="font-mono text-xs font-normal"
            >
              {perm}
            </Badge>
          ))}
        </div>
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <Users className="size-3.5" />
          {role.memberCount} {role.memberCount === 1 ? "member" : "members"}
        </div>
      </CardContent>
    </Card>
  );
}
