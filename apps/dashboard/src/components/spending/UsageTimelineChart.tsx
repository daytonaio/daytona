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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { useRegions } from '@/hooks/useRegions'
import { formatMoney } from '@/lib/utils'
import { ModelsUsageChartPoint } from '@daytona/analytics-api-client'
import type { RegionUsageOverview } from '@daytona/api-client'
import { differenceInCalendarDays } from 'date-fns'
import { useMemo, useState } from 'react'
import { Area, AreaChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from 'recharts'

type UsageTimelineChartProps = {
  data: ModelsUsageChartPoint[] | undefined
  isLoading: boolean
  regionUsage: RegionUsageOverview[] | undefined
  selectedRegion: string | undefined
  onRegionChange: (region: string) => void
  dateRange: { from: Date; to: Date }
}

type ChartMode = 'resources' | 'cost'
type ResourceFilter = 'all' | 'cpu' | 'ram' | 'disk'

const RESOURCE_COLORS = {
  cpu: 'hsl(var(--chart-1))',
  ram: 'hsl(var(--chart-2))',
  disk: 'hsl(var(--chart-3))',
}

const LIMIT_COLOR = '#ef4444'

const allResourceChartConfig: ChartConfig = {
  cpuPercent: { label: 'Compute', color: RESOURCE_COLORS.cpu },
  ramPercent: { label: 'Memory', color: RESOURCE_COLORS.ram },
  diskPercent: { label: 'Storage', color: RESOURCE_COLORS.disk },
}

const cpuChartConfig: ChartConfig = {
  cpu: { label: 'CPU (vCPU)', color: RESOURCE_COLORS.cpu },
}

const ramChartConfig: ChartConfig = {
  ramGB: { label: 'RAM (GiB)', color: RESOURCE_COLORS.ram },
}

const diskChartConfig: ChartConfig = {
  diskGB: { label: 'Disk (GiB)', color: RESOURCE_COLORS.disk },
}

const costChartConfig: ChartConfig = {
  cpuPrice: { label: 'CPU', color: RESOURCE_COLORS.cpu },
  ramPrice: { label: 'RAM', color: RESOURCE_COLORS.ram },
  diskPrice: { label: 'Disk', color: RESOURCE_COLORS.disk },
}

function formatTime(value: string) {
  const date = new Date(value)
  if (isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

function formatDateAndTime(value: string) {
  const date = new Date(value)
  if (isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

function formatTooltipLabel(value: string) {
  const date = new Date(value)
  if (isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

export function UsageTimelineChart({
  data,
  isLoading,
  regionUsage,
  selectedRegion,
  onRegionChange,
  dateRange,
}: UsageTimelineChartProps) {
  const { getRegionName } = useRegions()

  const [mode, setMode] = useState<ChartMode>('resources')
  const [resourceFilter, setResourceFilter] = useState<ResourceFilter>('all')

  const isMultiDay = useMemo(
    () => differenceInCalendarDays(dateRange.to, dateRange.from) >= 1,
    [dateRange.from, dateRange.to],
  )

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

  const chartData = useMemo(() => {
    if (!data?.length) return []
    return data.map((point) => {
      const cpu = point.cpu ?? 0
      const ramGB = point.ramGB ?? 0
      const diskGB = point.diskGB ?? 0

      const cpuPercent = limits?.cpu ? (cpu / limits.cpu) * 100 : 0
      const ramPercent = limits?.ram ? (ramGB / limits.ram) * 100 : 0
      const diskPercent = limits?.disk ? (diskGB / limits.disk) * 100 : 0

      return {
        time: point.time ?? '',
        cpu,
        ramGB,
        diskGB,
        cpuPrice: point.cpuPrice ?? 0,
        ramPrice: point.ramPrice ?? 0,
        diskPrice: point.diskPrice ?? 0,
        cpuPercent,
        ramPercent,
        diskPercent,
      }
    })
  }, [data, limits])

  const isResourceMode = mode === 'resources'

  const { chartConfig, dataKeys } = useMemo(() => {
    if (!isResourceMode) {
      return {
        chartConfig: costChartConfig,
        dataKeys: ['cpuPrice', 'ramPrice', 'diskPrice'] as string[],
      }
    }
    switch (resourceFilter) {
      case 'all':
        return {
          chartConfig: allResourceChartConfig,
          dataKeys: ['cpuPercent', 'ramPercent', 'diskPercent'] as string[],
        }
      case 'cpu':
        return { chartConfig: cpuChartConfig, dataKeys: ['cpu'] as string[] }
      case 'ram':
        return { chartConfig: ramChartConfig, dataKeys: ['ramGB'] as string[] }
      case 'disk':
        return { chartConfig: diskChartConfig, dataKeys: ['diskGB'] as string[] }
    }
  }, [isResourceMode, resourceFilter])

  const yAxisDomain = useMemo<[number, number] | undefined>(() => {
    if (!isResourceMode) return undefined
    if (resourceFilter === 'all') {
      return [0, 120]
    }
    if (!limits) return undefined
    const limitValue = resourceFilter === 'cpu' ? limits.cpu : resourceFilter === 'ram' ? limits.ram : limits.disk
    if (limitValue <= 0) return undefined
    // Round up to a nice number so ticks are evenly spaced
    const raw = limitValue * 1.2
    const tickCount = 4
    const interval = Math.ceil(raw / tickCount)
    const niceMax = interval * tickCount
    return [0, niceMax]
  }, [isResourceMode, resourceFilter, limits])

  const yAxisFormatter = useMemo(() => {
    if (!isResourceMode) return (value: number) => formatMoney(value)
    if (resourceFilter === 'all') return (value: number) => `${value}%`
    return (value: number) => value.toLocaleString()
  }, [isResourceMode, resourceFilter])

  const tooltipValueFormatter = useMemo(() => {
    if (!isResourceMode) return (value: number) => formatMoney(value)
    if (resourceFilter === 'all') return (value: number) => `${value.toFixed(1)}%`
    return (value: number) => value.toLocaleString()
  }, [isResourceMode, resourceFilter])

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <p className="text-xl font-semibold leading-none tracking-tight">Usage Timeline</p>
        <div className="flex items-center gap-3 flex-wrap">
          <Select
            value={selectedRegion ?? ''}
            onValueChange={(value) => onRegionChange(value)}
            disabled={!regionUsage?.length}
          >
            <SelectTrigger size="sm" className="w-[160px] rounded-lg" aria-label="Select region">
              <SelectValue placeholder="Select region" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              {regionUsage?.map((usage) => (
                <SelectItem key={usage.regionId} value={usage.regionId}>
                  {getRegionName(usage.regionId) ?? usage.regionId}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {isResourceMode && (
            <Select value={resourceFilter} onValueChange={(value) => setResourceFilter(value as ResourceFilter)}>
              <SelectTrigger size="sm" className="w-[150px] rounded-lg" aria-label="Select resource">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                <SelectItem value="all">All Resources</SelectItem>
                <SelectItem value="cpu">Compute</SelectItem>
                <SelectItem value="ram">Memory</SelectItem>
                <SelectItem value="disk">Storage</SelectItem>
              </SelectContent>
            </Select>
          )}
          <ToggleGroup
            type="single"
            value={mode}
            onValueChange={(value) => {
              if (value) setMode(value as ChartMode)
            }}
            variant="outline"
            size="sm"
          >
            <ToggleGroupItem value="resources">Resources</ToggleGroupItem>
            <ToggleGroupItem value="cost">Cost</ToggleGroupItem>
          </ToggleGroup>
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
              tickFormatter={isMultiDay ? formatDateAndTime : formatTime}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={4}
              tickCount={5}
              width={50}
              domain={yAxisDomain}
              tickFormatter={yAxisFormatter}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  indicator="dot"
                  labelFormatter={(label) => formatTooltipLabel(label)}
                  valueFormatter={(value) => tooltipValueFormatter(Number(value))}
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
                stackId={resourceFilter === 'all' && isResourceMode ? 'a' : undefined}
              />
            ))}
            {isResourceMode && resourceFilter === 'all' && (
              <ReferenceLine
                y={100}
                stroke={LIMIT_COLOR}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: 'Limit (100%)',
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLOR,
                }}
              />
            )}
            {isResourceMode && resourceFilter === 'cpu' && limits && limits.cpu > 0 && (
              <ReferenceLine
                y={limits.cpu}
                stroke={LIMIT_COLOR}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `CPU limit (${limits.cpu} vCPU)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLOR,
                }}
              />
            )}
            {isResourceMode && resourceFilter === 'ram' && limits && limits.ram > 0 && (
              <ReferenceLine
                y={limits.ram}
                stroke={LIMIT_COLOR}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `RAM limit (${limits.ram} GiB)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLOR,
                }}
              />
            )}
            {isResourceMode && resourceFilter === 'disk' && limits && limits.disk > 0 && (
              <ReferenceLine
                y={limits.disk}
                stroke={LIMIT_COLOR}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `Disk limit (${limits.disk} GiB)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLOR,
                }}
              />
            )}
            <ChartLegend content={<ChartLegendContent />} />
          </AreaChart>
        </ChartContainer>
      </div>
    </div>
  )
}
