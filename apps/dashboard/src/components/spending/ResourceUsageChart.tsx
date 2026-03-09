/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { formatMoney } from '@/lib/utils'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'

export type UsageChartData = {
  date: string
  ramGB: number
  cpu: number
  diskGB: number
  // gpu: number
}

type UsageChartProps = {
  data: Record<string, unknown>[]
  chartConfig: Record<string, { label: string; color: string }>
  chartType: 'bar' | 'area'
  showTotal?: boolean
}

const getShortDate = (value: string) => {
  const date = new Date(value)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    year: 'numeric',
  })
}

export function ResourceUsageChart({ data, chartConfig, chartType, showTotal }: UsageChartProps) {
  if (chartType === 'bar') {
    return (
      <ChartContainer config={chartConfig as ChartConfig} className="aspect-auto h-[300px] w-full">
        <BarChart accessibilityLayer data={data}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="date" tickLine={false} tickMargin={10} axisLine={false} tickFormatter={getShortDate} />
          <YAxis
            tickLine={false}
            axisLine={false}
            tickMargin={4}
            tickCount={4}
            width={45}
            tickFormatter={(value) => {
              if (!showTotal) {
                return value
              }

              return formatMoney(value)
            }}
          />
          <ChartTooltip
            cursor={false}
            content={
              <ChartTooltipContent
                indicator="dot"
                hideLabel={!showTotal}
                labelFormatter={(label, payload) => {
                  return `${getShortDate(label)}: ${formatMoney(payload.reduce((acc, curr) => acc + (curr.value as number), 0))}`
                }}
              />
            }
          />
          <ChartLegend content={<ChartLegendContent />} />
          <Bar dataKey="cpu" stackId="a" fill="var(--color-cpu)" radius={[0, 0, 0, 0]} />
          <Bar dataKey="ramGB" stackId="a" fill="var(--color-ramGB)" radius={[0, 0, 0, 0]} />
          <Bar dataKey="diskGB" stackId="a" fill="var(--color-diskGB)" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ChartContainer>
    )
  }

  return (
    <ChartContainer config={chartConfig as ChartConfig} className="aspect-auto h-[300px] w-full">
      <AreaChart data={data}>
        <defs>
          <linearGradient id="fillDiskGB" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-diskGB)" stopOpacity={0.8} />
            <stop offset="95%" stopColor="var(--color-diskGB)" stopOpacity={0.1} />
          </linearGradient>
          <linearGradient id="fillRamGB" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-ramGB)" stopOpacity={0.8} />
            <stop offset="95%" stopColor="var(--color-ramGB)" stopOpacity={0.1} />
          </linearGradient>
          <linearGradient id="fillCpu" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-cpu)" stopOpacity={0.8} />
            <stop offset="95%" stopColor="var(--color-cpu)" stopOpacity={0.1} />
          </linearGradient>
        </defs>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="date"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          minTickGap={32}
          tickFormatter={getShortDate}
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickMargin={4}
          tickCount={4}
          width={45}
          tickFormatter={(value) => formatMoney(value)}
        />
        <ChartTooltip
          cursor={false}
          content={
            <ChartTooltipContent
              indicator="dot"
              labelFormatter={(label, payload) => {
                return `${getShortDate(label)}: ${formatMoney(payload.reduce((acc, curr) => acc + (curr.value as number), 0))}`
              }}
            />
          }
        />
        <Area dataKey="diskGB" type="monotoneX" fill="url(#fillDiskGB)" stroke="var(--color-diskGB)" stackId="a" />
        <Area dataKey="ramGB" type="monotoneX" fill="url(#fillRamGB)" stroke="var(--color-ramGB)" stackId="a" />
        <Area dataKey="cpu" type="monotoneX" fill="url(#fillCpu)" stroke="var(--color-cpu)" stackId="a" />
        <ChartLegend content={<ChartLegendContent />} />
      </AreaChart>
    </ChartContainer>
  )
}
