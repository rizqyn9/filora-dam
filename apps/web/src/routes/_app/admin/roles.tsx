import { createFileRoute } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { RoleCard } from "@/features/roles/components/role-card";
import { roles } from "@/features/roles/fixtures";

export const Route = createFileRoute("/_app/admin/roles")({
  component: RolesPage,
});

function RolesPage() {
  return (
    <>
      <PageHeader
        title="Roles"
        description="Define roles and the permissions granted to each."
        actions={
          <Button size="sm">
            <Plus className="size-4" />
            New role
          </Button>
        }
      />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {roles.map((role) => (
          <RoleCard key={role.id} role={role} />
        ))}
      </div>
    </>
  );
}
