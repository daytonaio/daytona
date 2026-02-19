/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback } from 'react'
import { useSandboxUsagePeriods, AnalyticsUsageParams } from '@/hooks/useAnalyticsUsage'
import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { RefreshCw, DollarSign } from 'lucide-react'
import { format } from 'date-fns'
import { subDays } from 'date-fns'

interface SandboxSpendingTabProps {
  sandboxId: string
}

export const SandboxSpendingTab: React.FC<SandboxSpendingTabProps> = ({ sandboxId }) => {
  const [timeRange, setTimeRange] = useState(() => {
    const now = new Date()
    return { from: subDays(now, 30), to: now }
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
        <TimeRangeSelector onChange={handleTimeRangeChange} defaultRange={defaultRange} className="w-auto" />

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
              {data.map((period, index) => (
                <TableRow key={index}>
                  <TableCell className="text-sm">
                    {period.startAt ? format(new Date(period.startAt), 'yyyy-MM-dd HH:mm:ss') : '-'}
                  </TableCell>
                  <TableCell className="text-sm">
                    {period.endAt ? format(new Date(period.endAt), 'yyyy-MM-dd HH:mm:ss') : '-'}
                  </TableCell>
                  <TableCell className="text-right">{period.cpu ?? 0}</TableCell>
                  <TableCell className="text-right">{period.ramGB ?? 0}</TableCell>
                  <TableCell className="text-right">{period.diskGB ?? 0}</TableCell>
                  <TableCell className="text-right">${(period.price ?? 0).toFixed(2)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  )
}
