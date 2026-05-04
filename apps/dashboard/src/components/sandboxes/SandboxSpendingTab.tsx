/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { AnalyticsUsageParams, useSandboxUsagePeriods } from '@/hooks/queries/useAnalyticsUsage'
import { formatMoney } from '@/lib/utils'
import { format, subHours } from 'date-fns'
import { DollarSign, RefreshCw } from 'lucide-react'
import { useQueryStates } from 'nuqs'
import { useCallback, useMemo, useState } from 'react'
import { timeRangeSearchParams } from './SearchParams'

function formatTimestamp(timestamp: string) {
  try {
    return format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss')
  } catch {
    return timestamp
  }
}

function SpendingTableSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Start</TableHead>
          <TableHead>End</TableHead>
          <TableHead className="text-right">CPU</TableHead>
          <TableHead className="text-right">RAM (GB)</TableHead>
          <TableHead className="text-right">Disk (GB)</TableHead>
          <TableHead className="text-right">Price</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 8 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-4 w-36" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-36" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-4 w-8 ml-auto" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-4 w-8 ml-auto" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-4 w-8 ml-auto" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-4 w-14 ml-auto" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function SpendingErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyTitle>Failed to load spending</EmptyTitle>
        <EmptyDescription>Something went wrong while fetching usage periods.</EmptyDescription>
      </EmptyHeader>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="size-4" />
        Retry
      </Button>
    </Empty>
  )
}

function SpendingEmptyState({ hasFilters, onClearFilters }: { hasFilters: boolean; onClearFilters: () => void }) {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <DollarSign className="size-4" />
        </EmptyMedia>
        <EmptyTitle>{hasFilters ? 'No matching usage periods found' : 'No usage periods yet'}</EmptyTitle>
        <EmptyDescription>
          {hasFilters
            ? 'No usage periods matched your current filters.'
            : 'Usage periods will appear here after the sandbox has billable activity.'}
        </EmptyDescription>
      </EmptyHeader>
      {hasFilters ? (
        <Button variant="outline" size="sm" onClick={onClearFilters}>
          Clear filters
        </Button>
      ) : null}
    </Empty>
  )
}

export function SandboxSpendingTab({ sandboxId }: { sandboxId: string }) {
  const [timeRange, setTimeRange] = useQueryStates(timeRangeSearchParams)
  const [timeRangeSelectorKey, setTimeRangeSelectorKey] = useState(0)

  const resolvedFrom = useMemo(() => timeRange.from ?? subHours(new Date(), 24), [timeRange.from])
  const resolvedTo = useMemo(() => timeRange.to ?? new Date(), [timeRange.to])

  const queryParams: AnalyticsUsageParams = { from: resolvedFrom, to: resolvedTo }
  const { data, isLoading, isError, refetch } = useSandboxUsagePeriods(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback(
    (from: Date, to: Date) => {
      setTimeRange({ from, to })
    },
    [setTimeRange],
  )

  const handleTimeRangeClear = useCallback(() => {
    setTimeRange({ from: null, to: null })
  }, [setTimeRange])

  const hasFilters = Boolean(timeRange.from || timeRange.to)

  const handleClearFilters = useCallback(() => {
    setTimeRange({ from: null, to: null })
    setTimeRangeSelectorKey((key) => key + 1)
  }, [setTimeRange])

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3 shrink-0">
        <TimeRangeSelector
          key={timeRangeSelectorKey}
          onChange={handleTimeRangeChange}
          onClear={handleTimeRangeClear}
          defaultRange={timeRange.from && timeRange.to ? { from: timeRange.from, to: timeRange.to } : undefined}
          defaultSelectedQuickRange="Last 24 hours"
          className="w-auto"
        />
        <Button variant="ghost" size="icon-sm" onClick={() => refetch()} className="ml-auto">
          <RefreshCw className="size-4" />
        </Button>
      </div>

      {isLoading ? (
        <div className="flex-1 min-h-0 border rounded-md">
          <SpendingTableSkeleton />
        </div>
      ) : isError ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <SpendingErrorState onRetry={() => refetch()} />
        </div>
      ) : !data?.length ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <SpendingEmptyState hasFilters={hasFilters} onClearFilters={handleClearFilters} />
        </div>
      ) : (
        <ScrollArea
          fade="mask"
          fadeSide="end"
          horizontal
          className="flex-1 min-h-0 border rounded-md [&_[data-slot=scroll-area-viewport]>div]:!overflow-visible [&_[data-slot=scroll-area-viewport]>div>div]:!overflow-visible"
        >
          <Table>
            <TableHeader className="sticky top-0 z-10 bg-background after:absolute after:bottom-0 after:left-0 after:right-0 after:h-px after:bg-border">
              <TableRow>
                <TableHead>Start</TableHead>
                <TableHead>End</TableHead>
                <TableHead className="text-right">CPU</TableHead>
                <TableHead className="text-right">RAM (GB)</TableHead>
                <TableHead className="text-right">Disk (GB)</TableHead>
                <TableHead className="text-right">Price</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((period) => {
                const rowKey = `${period.startAt ?? 'unknown-start'}-${period.endAt ?? 'unknown-end'}`
                return (
                  <TableRow key={rowKey}>
                    <TableCell className="font-mono text-xs">
                      {period.startAt ? formatTimestamp(period.startAt) : '-'}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {period.endAt ? formatTimestamp(period.endAt) : '-'}
                    </TableCell>
                    <TableCell className="text-right">{period.cpu ?? 0}</TableCell>
                    <TableCell className="text-right">{period.ramGB ?? 0}</TableCell>
                    <TableCell className="text-right">{period.diskGB ?? 0}</TableCell>
                    <TableCell className="text-right font-mono text-xs">
                      {formatMoney(period.price ?? 0, {
                        maximumFractionDigits: 8,
                      })}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </ScrollArea>
      )}
    </div>
  )
}
