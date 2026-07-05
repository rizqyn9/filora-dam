import type { LucideIcon } from "lucide-react";
import { ArrowDownRight, ArrowUpRight } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatCardProps {
  label: string;
  value: string | number;
  icon?: LucideIcon;
  /** e.g. "+12.5%" — sign drives the up/down styling. */
  delta?: string;
  hint?: string;
  className?: string;
}

/**
 * Compact metric card for dashboards and summaries.
 */
export function StatCard({
  label,
  value,
  icon: Icon,
  delta,
  hint,
  className,
}: StatCardProps) {
  const isNegative = delta?.trim().startsWith("-");

  return (
    <Card className={cn("gap-0 py-0", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 px-4 pt-4 pb-2">
        <CardTitle className="text-xs font-medium text-muted-foreground">
          {label}
        </CardTitle>
        {Icon && <Icon className="size-4 text-muted-foreground" />}
      </CardHeader>
      <CardContent className="px-4 pb-4">
        <div className="text-2xl font-semibold tracking-tight tabular-nums">
          {value}
        </div>
        <div className="mt-1 flex items-center gap-2 text-xs">
          {delta && (
            <span
              className={cn(
                "inline-flex items-center gap-0.5 font-medium",
                isNegative ? "text-destructive" : "text-emerald-600",
              )}
            >
              {isNegative ? (
                <ArrowDownRight className="size-3" />
              ) : (
                <ArrowUpRight className="size-3" />
              )}
              {delta}
            </span>
          )}
          {hint && <span className="text-muted-foreground">{hint}</span>}
        </div>
      </CardContent>
    </Card>
  );
}
