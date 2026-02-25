/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback, useMemo } from 'react'
import { useQueryStates } from 'nuqs'
import { useSandboxMetrics, MetricsQueryParams } from '@/hooks/useSandboxMetrics'
import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { Button } from '@/components/ui/button'
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartConfig } from '@/components/ui/chart'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Legend } from 'recharts'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { RefreshCw, BarChart3 } from 'lucide-react'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { format, subHours } from 'date-fns'
import { MetricSeries } from '@daytonaio/api-client'
import { getMetricDisplayName } from '@/constants/metrics'
import { timeRangeSearchParams } from './SearchParams'

const CHART_COLORS = [
  'hsl(var(--chart-1))',
  'hsl(var(--chart-2))',
  'hsl(var(--chart-3))',
  'hsl(var(--chart-4))',
  'hsl(var(--chart-5))',
]

const BYTES_TO_GIB = 1024 * 1024 * 1024
type ViewMode = '%' | 'GiB'

const METRIC_GROUPS = [
  { key: 'cpu', title: 'CPU', prefix: '.cpu.', hasToggle: false },
  { key: 'memory', title: 'Memory', prefix: '.memory.', hasToggle: true },
  { key: 'filesystem', title: 'Filesystem', prefix: '.filesystem.', hasToggle: true },
]

function isByteMetric(metricName: string): boolean {
  return !metricName.endsWith('.utilization')
}

function buildChartData(series: MetricSeries[], convertToGiB: boolean): Record<string, unknown>[] {
  const timestampSet = new Set<string>()
  series.forEach((s) => s.dataPoints.forEach((p) => timestampSet.add(p.timestamp)))
  const timestamps = Array.from(timestampSet).sort()

  return timestamps.map((timestamp) => {
    const point: Record<string, unknown> = { timestamp }
    series.forEach((s) => {
      const dp = s.dataPoints.find((p) => p.timestamp === timestamp)
      if (dp?.value == null) {
        point[s.metricName] = null
      } else if (convertToGiB && isByteMetric(s.metricName)) {
        point[s.metricName] = Math.round((dp.value / BYTES_TO_GIB) * 100) / 100
      } else {
        point[s.metricName] = dp.value
      }
    })
    return point
  })
}

function buildChartConfig(series: MetricSeries[]): ChartConfig {
  const config: ChartConfig = {}
  series.forEach((s, index) => {
    config[s.metricName] = {
      label: getMetricDisplayName(s.metricName),
      color: CHART_COLORS[index % CHART_COLORS.length],
    }
  })
  return config
}

const formatXAxis = (timestamp: string) => {
  try {
    return format(new Date(timestamp), 'HH:mm')
  } catch {
    return timestamp
  }
}

function MetricGroupChart({
  title,
  series,
  convertToGiB,
  viewMode,
  onViewModeChange,
}: {
  title: string
  series: MetricSeries[]
  convertToGiB: boolean
  viewMode?: ViewMode
  onViewModeChange?: (mode: ViewMode) => void
}) {
  const chartData = React.useMemo(() => buildChartData(series, convertToGiB), [series, convertToGiB])
  const chartConfig = React.useMemo(() => buildChartConfig(series), [series])
  const displayTitle = viewMode ? `${title} (${viewMode})` : title

  return (
    <div className="rounded-md border border-border">
      <div className="flex items-center gap-2 px-4 py-3">
        <h3 className="text-sm font-medium">{displayTitle}</h3>
        {viewMode && onViewModeChange && (
          <ToggleGroup
            type="single"
            value={viewMode}
            onValueChange={(value) => {
              if (value) onViewModeChange(value as ViewMode)
            }}
            variant="outline"
            size="sm"
          >
            <ToggleGroupItem value="%" className="text-xs px-2 h-6">
              %
            </ToggleGroupItem>
            <ToggleGroupItem value="GiB" className="text-xs px-2 h-6">
              GiB
            </ToggleGroupItem>
          </ToggleGroup>
        )}
      </div>
      <div className="border-t border-border px-4 py-3">
        <ChartContainer config={chartConfig} className="h-[220px] w-full">
          <LineChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="timestamp"
              tickFormatter={formatXAxis}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
            />
            <YAxis
              width={35}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              tickFormatter={convertToGiB ? (value: number) => value.toFixed(2) : undefined}
              domain={viewMode === '%' ? [0, 100] : undefined}
            />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  labelFormatter={(label) => {
                    try {
                      return format(new Date(label as string), 'yyyy-MM-dd HH:mm:ss')
                    } catch {
                      return String(label)
                    }
                  }}
                />
              }
            />
            <Legend />
            {series.map((s, index) => {
              const isLimit = s.metricName.endsWith('.limit') || s.metricName.endsWith('.total')
              return (
                <Line
                  key={s.metricName}
                  type="monotone"
                  dataKey={s.metricName}
                  name={getMetricDisplayName(s.metricName)}
                  stroke={CHART_COLORS[index % CHART_COLORS.length]}
                  strokeWidth={isLimit ? 1.5 : 2}
                  strokeDasharray={isLimit ? '6 3' : undefined}
                  dot={false}
                  connectNulls
                />
              )
            })}
          </LineChart>
        </ChartContainer>
      </div>
    </div>
  )
}

function MetricsChartsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      {['CPU', 'Memory', 'Filesystem'].map((title) => (
        <div key={title} className="min-h-[250px]">
          <div className="flex items-center gap-2 mb-2">
            <Skeleton className="h-4 w-20" />
          </div>
          <Skeleton className="h-[220px] w-full rounded-md" />
        </div>
      ))}
    </div>
  )
}

function MetricsErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyTitle>Failed to load metrics</EmptyTitle>
        <EmptyDescription>Something went wrong while fetching metrics.</EmptyDescription>
      </EmptyHeader>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="size-4" />
        Retry
      </Button>
    </Empty>
  )
}

function MetricsEmptyState() {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <BarChart3 className="size-4" />
        </EmptyMedia>
        <EmptyTitle>No metrics available</EmptyTitle>
        <EmptyDescription>
          Metrics may take a moment to appear after the sandbox starts.{' '}
          <a href={`${DAYTONA_DOCS_URL}/en/experimental/otel-collection`} target="_blank" rel="noopener noreferrer">
            Learn more about observability
          </a>
          .
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function SandboxMetricsTab({ sandboxId }: { sandboxId: string }) {
  const [timeRange, setTimeRange] = useQueryStates(timeRangeSearchParams)
  const [viewModes, setViewModes] = useState<Record<string, ViewMode>>({ memory: '%', filesystem: '%' })

  const resolvedFrom = useMemo(() => timeRange.from ?? subHours(new Date(), 1), [timeRange.from])
  const resolvedTo = useMemo(() => timeRange.to ?? new Date(), [timeRange.to])

  const queryParams: MetricsQueryParams = { from: resolvedFrom, to: resolvedTo }
  const { data, isLoading, isError, refetch } = useSandboxMetrics(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback(
    (from: Date, to: Date) => {
      setTimeRange({ from, to })
    },
    [setTimeRange],
  )

  const handleViewModeChange = useCallback((groupKey: string, mode: ViewMode) => {
    setViewModes((prev) => ({ ...prev, [groupKey]: mode }))
  }, [])

  const groupedSeries = React.useMemo(() => {
    if (!data?.series?.length) return []

    return METRIC_GROUPS.map((group) => {
      const allSeries = data.series.filter((s) => s.metricName.includes(group.prefix))
      const mode = group.hasToggle ? viewModes[group.key] : undefined

      let filteredSeries = allSeries.filter((s) => s.metricName !== 'system.memory.utilization')
      if (mode === '%') {
        filteredSeries = filteredSeries.filter((s) => s.metricName.endsWith('.utilization'))
      } else if (mode === 'GiB') {
        filteredSeries = filteredSeries.filter((s) => !s.metricName.endsWith('.utilization'))
      }

      return {
        key: group.key,
        title: group.title,
        series: filteredSeries,
        convertToGiB: mode === 'GiB',
        hasToggle: group.hasToggle,
        viewMode: mode,
      }
    }).filter((group) => group.series.length > 0)
  }, [data, viewModes])

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3 shrink-0">
        <TimeRangeSelector
          onChange={handleTimeRangeChange}
          defaultRange={timeRange.from && timeRange.to ? { from: timeRange.from, to: timeRange.to } : undefined}
          className="w-auto"
        />
        <Button variant="ghost" size="icon-sm" onClick={() => refetch()} className="ml-auto">
          <RefreshCw className="size-4" />
        </Button>
      </div>

      {isLoading ? (
        <div className="flex-1 min-h-0 rounded-md border border-border p-4">
          <MetricsChartsSkeleton />
        </div>
      ) : isError ? (
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <MetricsErrorState onRetry={() => refetch()} />
        </div>
      ) : !data?.series?.length ? (
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <MetricsEmptyState />
        </div>
      ) : (
        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="flex flex-col gap-4">
            {groupedSeries.map((group) => (
              <MetricGroupChart
                key={group.key}
                title={group.title}
                series={group.series}
                convertToGiB={group.convertToGiB}
                viewMode={group.hasToggle ? group.viewMode : undefined}
                onViewModeChange={group.hasToggle ? (mode) => handleViewModeChange(group.key, mode) : undefined}
              />
            ))}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}
