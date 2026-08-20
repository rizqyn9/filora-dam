import { CheckCircle2, Clock, HardDrive, XCircle } from "lucide-react";
import { Pencil, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { RowActions } from "@/components/row-actions";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Switch } from "@/components/ui/switch";
import { providerLabels, type StorageAccount } from "@/features/storage/types";
import { formatBytes } from "@/lib/format";

function CountStat({
  icon: Icon,
  value,
  label,
  className,
}: {
  icon: typeof CheckCircle2;
  value: number;
  label: string;
  className?: string;
}) {
  return (
    <div className="flex items-center gap-1.5">
      <Icon className={className ?? "size-3.5 text-muted-foreground"} />
      <span className="font-medium tabular-nums">{value.toLocaleString()}</span>
      <span className="text-muted-foreground">{label}</span>
    </div>
  );
}

export function ProviderCard({ account }: { account: StorageAccount }) {
  const hasQuota = account.quota !== null;

  return (
    <Card>
      <CardHeader>
        <div className="flex size-9 items-center justify-center rounded-md bg-muted">
          <HardDrive className="size-4.5 text-muted-foreground" />
        </div>
        <CardTitle className="flex items-center gap-2 text-base">
          {account.name}
        </CardTitle>
        <div className="flex flex-wrap items-center gap-1.5">
          <Badge variant="outline">{providerLabels[account.type]}</Badge>
          <Badge
            variant={account.layer === "serving" ? "default" : "secondary"}
            className="capitalize"
          >
            {account.layer}
          </Badge>
        </div>
        <div className="ml-auto flex items-center gap-2">
          <Switch
            checked={account.isActive}
            onCheckedChange={(v) =>
              toast.success(`${account.name} ${v ? "activated" : "paused"}`)
            }
            aria-label="Toggle active"
          />
          <RowActions
            actions={[
              {
                label: "Edit",
                icon: Pencil,
                onSelect: () => toast.info("Edit provider"),
              },
              {
                label: "Remove",
                icon: Trash2,
                destructive: true,
                separatorBefore: true,
                onSelect: () => toast.success(`${account.name} removed`),
              },
            ]}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="space-y-1.5">
          <div className="flex items-baseline justify-between text-sm">
            <span className="font-medium tabular-nums">
              {formatBytes(account.used)}
            </span>
            <span className="text-xs text-muted-foreground">
              {hasQuota ? `of ${formatBytes(account.quota!)}` : "unlimited"}
            </span>
          </div>
          <Progress value={hasQuota ? account.usedPercent : 100} />
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs">
          <CountStat
            icon={CheckCircle2}
            value={account.storedCount}
            label="stored"
            className="size-3.5 text-emerald-600"
          />
          <CountStat
            icon={Clock}
            value={account.pendingCount}
            label="pending"
            className="size-3.5 text-amber-500"
          />
          <CountStat
            icon={XCircle}
            value={account.failedCount}
            label="failed"
            className="size-3.5 text-destructive"
          />
        </div>
      </CardContent>
    </Card>
  );
}
