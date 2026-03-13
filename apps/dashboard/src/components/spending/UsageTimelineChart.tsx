/*
 * Copyright Daytona Platforms Inc.
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
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'
import { formatMoney } from '@/lib/utils'
import { ModelsUsageChartPoint } from '@daytonaio/analytics-api-client'
import type { Region, RegionUsageOverview } from '@daytonaio/api-client'
import { useMemo, useState } from 'react'
import { Area, AreaChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from 'recharts'

type UsageTimelineChartProps = {
  data: ModelsUsageChartPoint[] | undefined
  isLoading: boolean
  regions: Region[]
  regionsLoading: boolean
  selectedRegion: string | undefined
  onRegionChange: (regionId: string | undefined) => void
  getRegionName: (regionId: string) => string | undefined
  regionUsage: RegionUsageOverview[] | undefined
}

type ChartMode = 'resources' | 'cost'

const LIMIT_COLORS = {
  cpu: 'hsl(var(--chart-1))',
  ram: 'hsl(var(--chart-2))',
  disk: 'hsl(var(--chart-3))',
}

const resourceChartConfig: ChartConfig = {
  cpu: { label: 'CPU (cores)', color: 'hsl(var(--chart-1))' },
  ramGB: { label: 'RAM (GB)', color: 'hsl(var(--chart-2))' },
  diskGB: { label: 'Disk (GB)', color: 'hsl(var(--chart-3))' },
}

const costChartConfig: ChartConfig = {
  cpuPrice: { label: 'CPU', color: 'hsl(var(--chart-1))' },
  ramPrice: { label: 'RAM', color: 'hsl(var(--chart-2))' },
  diskPrice: { label: 'Disk', color: 'hsl(var(--chart-3))' },
}

function formatTime(value: string) {
  const date = new Date(value)
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

const ALL_REGIONS_VALUE = '__all__'

export function UsageTimelineChart({
  data,
  isLoading,
  regions,
  regionsLoading,
  selectedRegion,
  onRegionChange,
  getRegionName,
  regionUsage,
}: UsageTimelineChartProps) {
  const [mode, setMode] = useState<ChartMode>('resources')

  const chartData = useMemo(() => {
    if (!data?.length) return []
    return data.map((point) => ({
      time: point.time ?? '',
      cpu: point.cpu ?? 0,
      ramGB: point.ramGB ?? 0,
      diskGB: point.diskGB ?? 0,
      cpuPrice: point.cpuPrice ?? 0,
      ramPrice: point.ramPrice ?? 0,
      diskPrice: point.diskPrice ?? 0,
    }))
  }, [data])

  const limits = useMemo(() => {
    if (!selectedRegion || !regionUsage) return null
    const region = regionUsage.find((r) => r.regionId === selectedRegion)
    if (!region) return null
    return {
      cpu: region.totalCpuQuota,
      ram: region.totalMemoryQuota,
      disk: region.totalDiskQuota,
    }
  }, [selectedRegion, regionUsage])

  const isResourceMode = mode === 'resources'
  const chartConfig = isResourceMode ? resourceChartConfig : costChartConfig
  const dataKeys = isResourceMode
    ? (['cpu', 'ramGB', 'diskGB'] as const)
    : (['cpuPrice', 'ramPrice', 'diskPrice'] as const)

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <p className="text-xl font-semibold leading-none tracking-tight">Usage Timeline</p>
        <div className="flex items-center gap-3">
          <Select
            value={selectedRegion ?? ALL_REGIONS_VALUE}
            onValueChange={(value) => onRegionChange(value === ALL_REGIONS_VALUE ? undefined : value)}
            disabled={regionsLoading}
          >
            <SelectTrigger size="sm" className="w-[160px] rounded-lg" aria-label="Select region">
              <SelectValue placeholder="All Regions" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value={ALL_REGIONS_VALUE}>All Regions</SelectItem>
              {regions.map((region) => (
                <SelectItem key={region.id} value={region.id}>
                  {getRegionName(region.id) ?? region.id}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className="flex items-center gap-2">
            <Label htmlFor="usage-mode-switch" className="text-sm text-muted-foreground">
              Resources
            </Label>
            <Switch
              id="usage-mode-switch"
              size="sm"
              checked={mode === 'cost'}
              onCheckedChange={(checked) => setMode(checked ? 'cost' : 'resources')}
            />
            <Label htmlFor="usage-mode-switch" className="text-sm text-muted-foreground">
              Cost
            </Label>
          </div>
        </div>
      </div>
      <div className="relative">
        {isLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center backdrop-grayscale backdrop-blur-[2px]">
            <Spinner className="size-6" />
          </div>
        )}
        <ChartContainer config={chartConfig} className="aspect-auto h-[300px] w-full">
          <AreaChart data={chartData}>
            <defs>
              {dataKeys.map((key) => (
                <linearGradient key={key} id={`fill-${key}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={`var(--color-${key})`} stopOpacity={0.8} />
                  <stop offset="95%" stopColor={`var(--color-${key})`} stopOpacity={0.1} />
                </linearGradient>
              ))}
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={48}
              tickFormatter={formatTime}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={4}
              tickCount={5}
              width={50}
              tickFormatter={(value) => (isResourceMode ? value.toLocaleString() : formatMoney(value))}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  indicator="dot"
                  labelFormatter={(label) => formatDateTime(label)}
                  valueFormatter={(value) =>
                    isResourceMode ? Number(value).toLocaleString() : formatMoney(Number(value))
                  }
                />
              }
            />
            {dataKeys.map((key) => (
              <Area
                key={key}
                dataKey={key}
                type="monotoneX"
                fill={`url(#fill-${key})`}
                stroke={`var(--color-${key})`}
                stackId="a"
              />
            ))}
            {isResourceMode && limits && (
              <>
                <ReferenceLine
                  y={limits.cpu}
                  stroke={LIMIT_COLORS.cpu}
                  strokeDasharray="6 3"
                  strokeWidth={1.5}
                  label={{
                    value: `CPU limit (${limits.cpu})`,
                    position: 'insideTopRight',
                    fontSize: 11,
                    fill: LIMIT_COLORS.cpu,
                  }}
                />
                <ReferenceLine
                  y={limits.ram}
                  stroke={LIMIT_COLORS.ram}
                  strokeDasharray="6 3"
                  strokeWidth={1.5}
                  label={{
                    value: `RAM limit (${limits.ram} GB)`,
                    position: 'insideTopRight',
                    fontSize: 11,
                    fill: LIMIT_COLORS.ram,
                  }}
                />
                <ReferenceLine
                  y={limits.disk}
                  stroke={LIMIT_COLORS.disk}
                  strokeDasharray="6 3"
                  strokeWidth={1.5}
                  label={{
                    value: `Disk limit (${limits.disk} GB)`,
                    position: 'insideTopRight',
                    fontSize: 11,
                    fill: LIMIT_COLORS.disk,
                  }}
                />
              </>
            )}
            <ChartLegend content={<ChartLegendContent />} />
          </AreaChart>
        </ChartContainer>
      </div>
    </div>
  )
}
