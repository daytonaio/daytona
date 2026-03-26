/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback } from 'react'
import { useSandboxUsagePeriods, AnalyticsUsageParams } from '@/hooks/queries/useAnalyticsUsage'
import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { RefreshCw, DollarSign } from 'lucide-react'
import { format, subHours } from 'date-fns'

function formatPrice(price: number): string {
  if (price >= 0.01) return `$${price.toFixed(2)}`
  if (price === 0) return '$0.00'
  return `$${parseFloat(price.toPrecision(2))}`
}

interface SandboxSpendingTabProps {
  sandboxId: string
}

export const SandboxSpendingTab: React.FC<SandboxSpendingTabProps> = ({ sandboxId }) => {
  const [timeRange, setTimeRange] = useState(() => {
    const now = new Date()
    return { from: subHours(now, 24), to: now }
  })

  const queryParams: AnalyticsUsageParams = {
    from: timeRange.from,
    to: timeRange.to,
  }

  const { data, isLoading, refetch } = useSandboxUsagePeriods(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback((from: Date, to: Date) => {
    setTimeRange({ from, to })
  }, [])

  const defaultRange = { from: timeRange.from, to: timeRange.to }

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3">
        <TimeRangeSelector
          onChange={handleTimeRangeChange}
          defaultRange={defaultRange}
          defaultSelectedQuickRange="Last 24 hours"
          className="w-auto"
        />

        <Button variant="outline" size="icon" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center h-full min-h-[400px]">
            <Spinner className="w-6 h-6" />
          </div>
        ) : !data?.length ? (
          <div className="flex flex-col items-center justify-center h-full min-h-[400px] text-muted-foreground gap-2">
            <DollarSign className="w-8 h-8" />
            <span className="text-sm">No usage periods found</span>
          </div>
        ) : (
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
              {data.map((period) => {
                const rowKey = `${period.startAt ?? 'unknown-start'}-${period.endAt ?? 'unknown-end'}`
                return (
                  <TableRow key={rowKey}>
                    <TableCell className="text-sm">
                      {period.startAt ? format(new Date(period.startAt), 'yyyy-MM-dd HH:mm:ss') : '-'}
                    </TableCell>
                    <TableCell className="text-sm">
                      {period.endAt ? format(new Date(period.endAt), 'yyyy-MM-dd HH:mm:ss') : '-'}
                    </TableCell>
                    <TableCell className="text-right">{period.cpu ?? 0}</TableCell>
                    <TableCell className="text-right">{period.ramGB ?? 0}</TableCell>
                    <TableCell className="text-right">{period.diskGB ?? 0}</TableCell>
                    <TableCell className="text-right">{formatPrice(period.price ?? 0)}</TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  )
}
