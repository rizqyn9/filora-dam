import { createFileRoute } from "@tanstack/react-router";
import { Download, HardDrive, Images, Layers, Upload } from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { AssetTypeChart } from "@/features/dashboard/components/asset-type-chart";
import { RecentUploads } from "@/features/dashboard/components/recent-uploads";
import { UsageChart } from "@/features/dashboard/components/usage-chart";
import {
  recentAssets,
  summary,
  typeCounts,
  usageSeries,
} from "@/features/dashboard/fixtures";
import { formatBytes } from "@/lib/format";

export const Route = createFileRoute("/_app/")({
  component: DashboardPage,
});

function DashboardPage() {
  const usedPct = Math.round(
    (summary.storageUsed / summary.storageQuota) * 100,
  );

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="Overview of your assets and storage."
        actions={
          <Button size="sm">
            <Upload className="size-4" />
            Upload
          </Button>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label="Total assets"
          value={summary.totalAssets.toLocaleString()}
          icon={Layers}
          delta="+8.2%"
          hint="vs last month"
        />
        <StatCard
          label="Storage used"
          value={formatBytes(summary.totalSize)}
          icon={HardDrive}
          delta="+3.1%"
          hint="vs last month"
        />
        <StatCard
          label="Galleries"
          value={summary.galleries}
          icon={Images}
          delta="+2"
          hint="new this month"
        />
        <StatCard
          label="Downloads"
          value="2,140"
          icon={Download}
          delta="-1.4%"
          hint="vs last month"
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <UsageChart data={usageSeries} />
        <AssetTypeChart data={typeCounts} />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <RecentUploads assets={recentAssets} />
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Storage</CardTitle>
            <CardDescription>Across all storage accounts.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-baseline justify-between">
              <span className="text-2xl font-semibold tabular-nums">
                {formatBytes(summary.storageUsed)}
              </span>
              <span className="text-sm text-muted-foreground">
                of {formatBytes(summary.storageQuota)}
              </span>
            </div>
            <Progress value={usedPct} />
            <p className="text-xs text-muted-foreground">
              {usedPct}% used ·{" "}
              {formatBytes(summary.storageQuota - summary.storageUsed)} free
            </p>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
