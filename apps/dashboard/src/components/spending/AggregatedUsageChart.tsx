/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { ModelsAggregatedUsage } from '@daytonaio/analytics-api-client'

interface AggregatedUsageChartProps {
  data: ModelsAggregatedUsage | undefined
  isLoading: boolean
}

function formatSeconds(seconds: number): string {
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`
  return `${(seconds / 3600).toFixed(1)}h`
}

function formatGBSeconds(gbSeconds: number): string {
  if (gbSeconds < 3600) return `${gbSeconds.toFixed(1)} GB-s`
  return `${(gbSeconds / 3600).toFixed(1)} GB-h`
}

function formatPrice(price: number): string {
  return `$${price.toFixed(2)}`
}

const chartConfig: ChartConfig = {
  cpu: {
    label: 'CPU',
    color: 'hsl(var(--chart-1))',
  },
  ram: {
    label: 'RAM',
    color: 'hsl(var(--chart-2))',
  },
  disk: {
    label: 'Disk',
    color: 'hsl(var(--chart-3))',
  },
}

export const AggregatedUsageChart: React.FC<AggregatedUsageChartProps> = ({ data, isLoading }) => {
  if (isLoading || !data) {
    return null
  }

  const totalPrice = data.totalPrice ?? 0
  const cpuSeconds = data.totalCPUSeconds ?? 0
  const ramGBSeconds = data.totalRAMGBSeconds ?? 0
  const diskGBSeconds = data.totalDiskGBSeconds ?? 0
  const sandboxCount = data.sandboxCount ?? 0

  const chartData = [
    {
      name: 'Resource Usage',
      cpu: cpuSeconds,
      ram: ramGBSeconds,
      disk: diskGBSeconds,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Price</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatPrice(totalPrice)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">CPU</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatSeconds(cpuSeconds)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">RAM</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatGBSeconds(ramGBSeconds)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Disk</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatGBSeconds(diskGBSeconds)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Sandboxes</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{sandboxCount}</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="flex items-center gap-2 space-y-0 border-b p-4 sm:flex-row">
          <div className="grid flex-1 gap-1 text-center sm:text-left">
            <CardTitle>Resource Usage Breakdown</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
          <ChartContainer config={chartConfig} className="aspect-auto h-[200px] w-full">
            <BarChart accessibilityLayer data={chartData} layout="vertical">
              <CartesianGrid horizontal={false} />
              <XAxis type="number" tickLine={false} axisLine={false} />
              <YAxis dataKey="name" type="category" tickLine={false} axisLine={false} hide />
              <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dot" />} />
              <ChartLegend content={<ChartLegendContent />} />
              <Bar dataKey="cpu" stackId="a" fill="var(--color-cpu)" radius={[0, 0, 0, 0]} />
              <Bar dataKey="ram" stackId="a" fill="var(--color-ram)" radius={[0, 0, 0, 0]} />
              <Bar dataKey="disk" stackId="a" fill="var(--color-disk)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  )
}
