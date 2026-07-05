import { useMemo } from "react";
import { Cell, Label, Pie, PieChart } from "recharts";

import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type { TypeCount } from "@/features/dashboard/types";

const chartConfig = {
  count: { label: "Assets" },
  image: { label: "Images", color: "var(--chart-1)" },
  video: { label: "Videos", color: "var(--chart-2)" },
  document: { label: "Documents", color: "var(--chart-3)" },
  archive: { label: "Archives", color: "var(--chart-4)" },
  file: { label: "Files", color: "var(--chart-5)" },
} satisfies ChartConfig;

export function AssetTypeChart({ data }: { data: TypeCount[] }) {
  const total = useMemo(
    () => data.reduce((acc, d) => acc + d.count, 0),
    [data],
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Asset types</CardTitle>
        <CardDescription>Distribution across all galleries.</CardDescription>
      </CardHeader>
      <ChartContainer
        config={chartConfig}
        className="mx-auto aspect-square max-h-64"
      >
        <PieChart>
          <ChartTooltip
            cursor={false}
            content={<ChartTooltipContent hideLabel />}
          />
          <Pie
            data={data}
            dataKey="count"
            nameKey="type"
            innerRadius={64}
            strokeWidth={4}
          >
            {data.map((d) => (
              <Cell key={d.type} fill={`var(--color-${d.type})`} />
            ))}
            {/* recharts render prop */}
            <Label
              content={({ viewBox }) => {
                if (!viewBox || !("cx" in viewBox)) return null;
                return (
                  <text
                    x={viewBox.cx}
                    y={viewBox.cy}
                    textAnchor="middle"
                    dominantBaseline="middle"
                  >
                    <tspan
                      x={viewBox.cx}
                      y={viewBox.cy}
                      className="fill-foreground text-2xl font-semibold"
                    >
                      {total.toLocaleString()}
                    </tspan>
                    <tspan
                      x={viewBox.cx}
                      y={(viewBox.cy ?? 0) + 20}
                      className="fill-muted-foreground text-xs"
                    >
                      Total assets
                    </tspan>
                  </text>
                );
              }}
            />
          </Pie>
        </PieChart>
      </ChartContainer>
    </Card>
  );
}
