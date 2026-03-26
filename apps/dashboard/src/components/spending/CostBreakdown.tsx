/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { subMonths } from 'date-fns'
import { AlertCircle, BarChart3, RefreshCw } from 'lucide-react'
import * as React from 'react'
import { useCallback, useMemo } from 'react'
import { FacetFilter } from '../ui/facet-filter'
import { ResourceUsageChart, UsageChartData } from './ResourceUsageChart'

type CostBreakdownProps = {
  usageData: UsageChartData[]
  showTotal?: boolean
  isLoading?: boolean
  isError?: boolean
  onRetry?: () => void
}

export function CostBreakdown({ usageData, showTotal, isLoading, isError, onRetry }: CostBreakdownProps) {
  const [timeRange, setTimeRange] = React.useState(12)
  const [chartType, setChartType] = React.useState<'bar' | 'area'>('bar')
  const [filters, setFilters] = React.useState<Set<string>>(new Set(['ramGB', 'cpu', 'diskGB']))

  const filterFromObject = useCallback(
    (obj: any) => {
      return Object.fromEntries(Object.entries(obj).filter(([key]) => key === 'date' || filters.has(key)))
    },
    [filters],
  )

  const data = useMemo(() => {
    const referenceDate = new Date()

    let monthsToSubtract = 12
    if (timeRange === 6) {
      monthsToSubtract = 6
    } else if (timeRange === 3) {
      monthsToSubtract = 3
    }
    const startDate = subMonths(new Date(referenceDate), monthsToSubtract)

    const filteredData = usageData.filter((usage) => {
      const date = new Date(usage.date)
      return date >= startDate
    })

    if (filteredData.length < timeRange) {
      const missingData: UsageChartData[] = []
      for (let i = 0; i < timeRange - filteredData.length; i++) {
        const refDate = filteredData.length > 0 ? filteredData[0].date : new Date()
        missingData.push({
          date: subMonths(new Date(refDate), i + 1).toISOString(),
          diskGB: 0,
          ramGB: 0,
          cpu: 0,
        })
      }
      return [...missingData.reverse(), ...filteredData].map(filterFromObject)
    }

    return filteredData.map(filterFromObject)
  }, [usageData, timeRange, filterFromObject])

  const chartConfig = useMemo(() => {
    const mapped: Record<string, { label: string; color: string }> = {
      cpu: {
        label: 'CPU',
        color: 'hsl(var(--chart-1))',
      },
      ramGB: {
        label: 'RAM',
        color: 'hsl(var(--chart-2))',
      },
      diskGB: {
        label: 'Disk',
        color: 'hsl(var(--chart-3))',
      },
    }

    return filterFromObject(mapped) as Record<string, { label: string; color: string }>
  }, [filterFromObject])

  const noData = !isLoading && usageData.length === 0

  return (
    <Card>
      <CardHeader className="flex flex-col sm:flex-row sm:items-center gap-2 space-y-0 border-b p-4">
        <div className="flex-1">
          <CardTitle>Monthly Cost Breakdown</CardTitle>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <FacetFilter
            title="Filters"
            className="h-8 pr-1"
            options={[
              { label: 'CPU', value: 'cpu' },
              { label: 'RAM', value: 'ramGB' },
              { label: 'Disk', value: 'diskGB' },
            ]}
            selectedValues={filters}
            setSelectedValues={setFilters}
          />
          <Select value={chartType} onValueChange={(value) => setChartType(value as 'bar' | 'area')}>
            <SelectTrigger size="sm" className="w-[80px] rounded-lg" aria-label="Select a chart type">
              <SelectValue placeholder="Bar" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="bar">Bar</SelectItem>
              <SelectItem value="area">Area</SelectItem>
            </SelectContent>
          </Select>
          <Select value={timeRange.toString()} onValueChange={(value) => setTimeRange(Number(value))}>
            <SelectTrigger size="sm" className="w-[150px] rounded-lg" aria-label="Select the range of months">
              <SelectValue placeholder="Last 12 months" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="12">Last 12 months</SelectItem>
              <SelectItem value="6">Last 6 months</SelectItem>
              <SelectItem value="3">Last 3 months</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      {isError ? (
        <Empty className="py-12">
          <EmptyHeader>
            <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
              <AlertCircle />
            </EmptyMedia>
            <EmptyTitle className="text-destructive">Failed to load billing data</EmptyTitle>
            <EmptyDescription>Something went wrong while fetching billing data. Please try again.</EmptyDescription>
          </EmptyHeader>
          {onRetry && (
            <EmptyContent>
              <Button variant="secondary" size="sm" onClick={onRetry}>
                <RefreshCw />
                Retry
              </Button>
            </EmptyContent>
          )}
        </Empty>
      ) : noData ? (
        <Empty className="py-12">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <BarChart3 />
            </EmptyMedia>
            <EmptyTitle>No billing data yet</EmptyTitle>
            <EmptyDescription>
              Monthly cost data will show up here once your organization has usage to report.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <CardContent className="relative px-2 pt-4 sm:px-6 sm:pt-6">
          {isLoading && (
            <div className="absolute inset-0 z-10 flex items-center justify-center backdrop-grayscale backdrop-blur-[2px]">
              <Spinner className="size-6" />
            </div>
          )}
          <ResourceUsageChart data={data} chartConfig={chartConfig} chartType={chartType} showTotal={showTotal} />
        </CardContent>
      )}
    </Card>
  )
}
