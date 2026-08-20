import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type { UsagePoint } from "@/features/dashboard/types";

const chartConfig = {
  images: { label: "Images", color: "var(--chart-1)" },
  videos: { label: "Videos", color: "var(--chart-2)" },
  documents: { label: "Documents", color: "var(--chart-3)" },
} satisfies ChartConfig;

export function UsageChart({ data }: { data: UsagePoint[] }) {
  return (
    <Card className="col-span-full lg:col-span-2">
      <CardHeader>
        <CardTitle>Uploads over time</CardTitle>
        <CardDescription>Assets added per month, by type.</CardDescription>
      </CardHeader>
      <ChartContainer
        config={chartConfig}
        className="aspect-auto h-64 w-full px-2"
      >
        <AreaChart data={data} margin={{ left: 8, right: 8, top: 8 }}>
          <defs>
            {Object.entries(chartConfig).map(([key, cfg]) => (
              <linearGradient
                key={key}
                id={`fill-${key}`}
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop offset="5%" stopColor={cfg.color} stopOpacity={0.7} />
                <stop offset="95%" stopColor={cfg.color} stopOpacity={0.05} />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid vertical={false} strokeDasharray="3 3" />
          <XAxis
            dataKey="month"
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            fontSize={12}
          />
          <ChartTooltip content={<ChartTooltipContent />} />
          <ChartLegend content={<ChartLegendContent />} />
          {Object.keys(chartConfig).map((key) => (
            <Area
              key={key}
              dataKey={key}
              type="natural"
              stackId="1"
              stroke={`var(--color-${key})`}
              fill={`url(#fill-${key})`}
            />
          ))}
        </AreaChart>
      </ChartContainer>
    </Card>
  );
}
