import { createFileRoute, Link } from "@tanstack/react-router";
import {
  ArrowRight,
  CheckCircle2,
  Clock,
  Database,
  HardDrive,
  Loader2,
  Users,
  XCircle,
} from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { roles } from "@/features/roles/fixtures";
import { storageAccounts } from "@/features/storage/fixtures";
import { users } from "@/features/users/fixtures";

const archiveJobs = { pending: 60, running: 4, completed: 24_812, failed: 7 };

const jobStats = [
  {
    label: "Completed",
    value: archiveJobs.completed,
    icon: CheckCircle2,
    className: "text-emerald-600",
  },
  {
    label: "Running",
    value: archiveJobs.running,
    icon: Loader2,
    className: "text-blue-500",
  },
  {
    label: "Pending",
    value: archiveJobs.pending,
    icon: Clock,
    className: "text-amber-500",
  },
  {
    label: "Failed",
    value: archiveJobs.failed,
    icon: XCircle,
    className: "text-destructive",
  },
];

const shortcuts = [
  {
    title: "Users",
    description: "Members & invitations",
    to: "/admin/users",
    icon: Users,
  },
  {
    title: "Roles",
    description: "RBAC & permissions",
    to: "/admin/roles",
    icon: Database,
  },
  {
    title: "Storage",
    description: "Providers & usage",
    to: "/storage",
    icon: HardDrive,
  },
] as const;

export const Route = createFileRoute("/_app/admin/")({
  component: AdminPage,
});

function AdminPage() {
  return (
    <>
      <PageHeader
        title="Admin"
        description="System health and workspace administration."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Users" value={users.length} icon={Users} />
        <StatCard label="Roles" value={roles.length} icon={Database} />
        <StatCard
          label="Storage accounts"
          value={storageAccounts.length}
          icon={HardDrive}
        />
        <StatCard
          label="Failed jobs"
          value={archiveJobs.failed}
          icon={XCircle}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Archive replication</CardTitle>
          <CardDescription>
            Background jobs copying assets to the archive layer.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {jobStats.map((stat) => (
            <div key={stat.label} className="flex items-center gap-3">
              <div className="flex size-9 items-center justify-center rounded-md bg-muted">
                <stat.icon className={`size-4.5 ${stat.className}`} />
              </div>
              <div>
                <p className="text-lg font-semibold tabular-nums">
                  {stat.value.toLocaleString()}
                </p>
                <p className="text-xs text-muted-foreground">{stat.label}</p>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <div className="grid gap-4 sm:grid-cols-3">
        {shortcuts.map((item) => (
          <Link key={item.to} to={item.to}>
            <Card className="transition-colors hover:border-primary/40 hover:bg-accent/40">
              <CardContent className="flex items-center gap-3">
                <div className="flex size-10 items-center justify-center rounded-md bg-muted">
                  <item.icon className="size-5 text-muted-foreground" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium">{item.title}</p>
                  <p className="text-xs text-muted-foreground">
                    {item.description}
                  </p>
                </div>
                <ArrowRight className="size-4 text-muted-foreground" />
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </>
  );
}
