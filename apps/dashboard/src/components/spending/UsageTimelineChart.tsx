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
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { formatMoney } from '@/lib/utils'
import { ModelsUsageChartPoint } from '@daytonaio/analytics-api-client'
import type { RegionUsageOverview } from '@daytonaio/api-client'
import { useEffect, useMemo, useState } from 'react'
import { Area, AreaChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from 'recharts'

type UsageTimelineChartProps = {
  data: ModelsUsageChartPoint[] | undefined
  isLoading: boolean
  regionUsage: RegionUsageOverview[] | undefined
}

type ChartMode = 'resources' | 'cost'

const LIMIT_COLORS = {
  cpu: '#3b82f6',
  ram: '#22c55e',
  disk: '#a855f7',
}

const resourceChartConfig: ChartConfig = {
  cpu: { label: 'CPU (vCPU)', color: 'hsl(var(--chart-1))' },
  ramGB: { label: 'RAM (GiB)', color: 'hsl(var(--chart-2))' },
  diskGB: { label: 'Disk (GiB)', color: 'hsl(var(--chart-3))' },
}

const costChartConfig: ChartConfig = {
  cpuPrice: { label: 'CPU', color: 'hsl(var(--chart-1))' },
  ramPrice: { label: 'RAM', color: 'hsl(var(--chart-2))' },
  diskPrice: { label: 'Disk', color: 'hsl(var(--chart-3))' },
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

function formatDateTime(value: string) {
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

const ALL_REGIONS_VALUE = '__all__'

export function UsageTimelineChart({ data, isLoading, regionUsage }: UsageTimelineChartProps) {
  const { getRegionName } = useRegions()
  const { selectedOrganization } = useSelectedOrganization()

  const [mode, setMode] = useState<ChartMode>('resources')
  const [selectedRegion, setSelectedRegion] = useState<string | undefined>(undefined)

  // Default to the organization's default region once regionUsage is available
  useEffect(() => {
    if (selectedRegion) return
    if (!regionUsage?.length) return
    const defaultRegionId = selectedOrganization?.defaultRegionId
    if (defaultRegionId && regionUsage.some((r) => r.regionId === defaultRegionId)) {
      setSelectedRegion(defaultRegionId)
    } else {
      setSelectedRegion(regionUsage[0].regionId)
    }
  }, [regionUsage, selectedOrganization?.defaultRegionId, selectedRegion])

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

  // Expand the Y axis domain to include reference line values
  const yAxisDomain = useMemo<[number, number] | undefined>(() => {
    if (!isResourceMode || !limits) return undefined
    const maxLimit = Math.max(limits.cpu, limits.ram, limits.disk)
    // Add 10% padding above the highest limit
    return [0, Math.ceil(maxLimit * 1.1)]
  }, [isResourceMode, limits])

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <p className="text-xl font-semibold leading-none tracking-tight">Usage Timeline</p>
        <div className="flex items-center gap-3 flex-wrap">
          <Select
            value={selectedRegion ?? ALL_REGIONS_VALUE}
            onValueChange={(value) => setSelectedRegion(value === ALL_REGIONS_VALUE ? undefined : value)}
            disabled={!regionUsage?.length}
          >
            <SelectTrigger size="sm" className="w-[160px] rounded-lg" aria-label="Select region">
              <SelectValue placeholder="All Regions" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value={ALL_REGIONS_VALUE}>All Regions</SelectItem>
              {regionUsage?.map((usage) => (
                <SelectItem key={usage.regionId} value={usage.regionId}>
                  {getRegionName(usage.regionId) ?? usage.regionId}
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
              domain={yAxisDomain}
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
            {isResourceMode && limits && limits.cpu > 0 && (
              <ReferenceLine
                y={limits.cpu}
                stroke={LIMIT_COLORS.cpu}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `CPU limit (${limits.cpu} vCPU)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLORS.cpu,
                }}
              />
            )}
            {isResourceMode && limits && limits.ram > 0 && (
              <ReferenceLine
                y={limits.ram}
                stroke={LIMIT_COLORS.ram}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `RAM limit (${limits.ram} GiB)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLORS.ram,
                }}
              />
            )}
            {isResourceMode && limits && limits.disk > 0 && (
              <ReferenceLine
                y={limits.disk}
                stroke={LIMIT_COLORS.disk}
                strokeDasharray="6 3"
                strokeWidth={1.5}
                label={{
                  value: `Disk limit (${limits.disk} GiB)`,
                  position: 'insideTopRight',
                  fontSize: 11,
                  fill: LIMIT_COLORS.disk,
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
