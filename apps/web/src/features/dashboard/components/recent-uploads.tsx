import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { RecentAsset } from "@/features/dashboard/types";
import { assetTypeMeta } from "@/lib/asset-type";
import { formatBytes, formatDate } from "@/lib/format";
import { cn } from "@/lib/utils";

export function RecentUploads({ assets }: { assets: RecentAsset[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent uploads</CardTitle>
        <CardDescription>Latest assets across your workspace.</CardDescription>
        <CardAction>
          <Button variant="ghost" size="sm">
            View all
          </Button>
        </CardAction>
      </CardHeader>
      <div className="divide-y">
        {assets.map((asset) => {
          const meta = assetTypeMeta(asset.type);
          return (
            <div key={asset.id} className="flex items-center gap-3 px-6 py-2.5">
              <div
                className={cn(
                  "flex size-9 shrink-0 items-center justify-center rounded-md",
                  meta.className,
                )}
              >
                <meta.icon className="size-4.5" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{asset.name}</p>
                <p className="text-xs text-muted-foreground">
                  {formatBytes(asset.size)} · {formatDate(asset.createdAt)}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
