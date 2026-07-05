import { createFileRoute } from "@tanstack/react-router";
import { UserPlus } from "lucide-react";

import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { userColumns } from "@/features/users/components/user-columns";
import { users } from "@/features/users/fixtures";

export const Route = createFileRoute("/_app/admin/users")({
  component: UsersPage,
});

function UsersPage() {
  return (
    <>
      <PageHeader
        title="Users"
        description="Manage workspace members and their roles."
        actions={
          <Button size="sm">
            <UserPlus className="size-4" />
            Invite user
          </Button>
        }
      />
      <DataTable
        columns={userColumns}
        data={users}
        searchColumn="name"
        searchPlaceholder="Search users..."
        emptyMessage="No users found."
      />
    </>
  );
}
