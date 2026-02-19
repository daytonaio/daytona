/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback } from 'react'
import { useSandboxMetrics, MetricsQueryParams } from '@/hooks/useSandboxMetrics'
import { TimeRangeSelector } from './TimeRangeSelector'
import { Button } from '@/components/ui/button'
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartConfig } from '@/components/ui/chart'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, ResponsiveContainer, Legend } from 'recharts'
import { RefreshCw, BarChart3 } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'
import { format } from 'date-fns'
import { subHours } from 'date-fns'
import { MetricSeries } from '@daytonaio/api-client'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

interface MetricsTabProps {
  sandboxId: string
}

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
  series.forEach((s) => {
    s.dataPoints.forEach((point) => {
      timestampSet.add(point.timestamp)
    })
  })

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
      label: s.metricName.replace(/^daytona\.sandbox\./, ''),
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

interface MetricGroupChartProps {
  title: string
  series: MetricSeries[]
  convertToGiB: boolean
  viewMode?: ViewMode
  onViewModeChange?: (mode: ViewMode) => void
}

const MetricGroupChart: React.FC<MetricGroupChartProps> = ({
  title,
  series,
  convertToGiB,
  viewMode,
  onViewModeChange,
}) => {
  const chartData = React.useMemo(() => buildChartData(series, convertToGiB), [series, convertToGiB])
  const chartConfig = React.useMemo(() => buildChartConfig(series), [series])

  const displayTitle = viewMode ? `${title} (${viewMode})` : title

  return (
    <div className="min-h-[250px]">
      <div className="flex items-center gap-2 mb-2">
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
      <ChartContainer config={chartConfig} className="h-[220px] w-full">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData} margin={{ top: 10, right: 30, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="timestamp"
              tickFormatter={formatXAxis}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
            />
            <YAxis
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
            {series.map((s, index) => (
              <Line
                key={s.metricName}
                type="monotone"
                dataKey={s.metricName}
                stroke={CHART_COLORS[index % CHART_COLORS.length]}
                strokeWidth={2}
                dot={false}
                connectNulls
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>
    </div>
  )
}

export const MetricsTab: React.FC<MetricsTabProps> = ({ sandboxId }) => {
  const [timeRange, setTimeRange] = useState(() => {
    const now = new Date()
    return { from: subHours(now, 1), to: now }
  })

  const queryParams: MetricsQueryParams = {
    from: timeRange.from,
    to: timeRange.to,
  }

  const { data, isLoading, refetch } = useSandboxMetrics(sandboxId, queryParams)

  const [viewModes, setViewModes] = useState<Record<string, ViewMode>>({
    memory: '%',
    filesystem: '%',
  })

  const handleTimeRangeChange = useCallback((from: Date, to: Date) => {
    setTimeRange({ from, to })
  }, [])

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
      <div className="flex flex-wrap items-center gap-3">
        <TimeRangeSelector onChange={handleTimeRangeChange} className="w-auto" />

        <Button variant="outline" size="icon" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center h-full min-h-[400px]">
            <Spinner className="w-6 h-6" />
          </div>
        ) : !data?.series?.length ? (
          <div className="flex flex-col items-center justify-center h-full min-h-[400px] text-muted-foreground gap-2">
            <BarChart3 className="w-8 h-8" />
            <span className="text-sm">No metrics found</span>
          </div>
        ) : (
          <div className="flex flex-col gap-6">
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
        )}
      </div>
    </div>
  )
}
