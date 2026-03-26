/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback } from 'react'
import { useSandboxTraces, TracesQueryParams } from '@/hooks/useSandboxTraces'
import { TimeRangeSelector } from './TimeRangeSelector'
import { TraceDetailsSheet } from './TraceDetailsSheet'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChevronLeft, ChevronRight, RefreshCw, Activity } from 'lucide-react'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { CopyButton } from '@/components/CopyButton'
import { Spinner } from '@/components/ui/spinner'
import { format } from 'date-fns'
import { subHours } from 'date-fns'
import { TraceSummary } from '@daytonaio/api-client'

interface TracesTabProps {
  sandboxId: string
}

export const TracesTab: React.FC<TracesTabProps> = ({ sandboxId }) => {
  const [timeRange, setTimeRange] = useState(() => {
    const now = new Date()
    return { from: subHours(now, 1), to: now }
  })
  const [page, setPage] = useState(1)
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null)
  const limit = 50

  const queryParams: TracesQueryParams = {
    from: timeRange.from,
    to: timeRange.to,
    page,
    limit,
  }

  const { data, isLoading, refetch } = useSandboxTraces(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback((from: Date, to: Date) => {
    setTimeRange({ from, to })
    setPage(1)
  }, [])

  const formatTimestamp = (timestamp: string) => {
    try {
      return format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')
    } catch {
      return timestamp
    }
  }

  const formatDuration = (durationMs: number) => {
    if (durationMs < 1) {
      return `${(durationMs * 1000).toFixed(2)}us`
    }
    if (durationMs < 1000) {
      return `${durationMs.toFixed(2)}ms`
    }
    return `${(durationMs / 1000).toFixed(2)}s`
  }

  const truncateTraceId = (traceId: string) => {
    if (traceId.length > 16) {
      return `${traceId.slice(0, 8)}...${traceId.slice(-8)}`
    }
    return traceId
  }

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3">
        <TimeRangeSelector onChange={handleTimeRangeChange} className="w-auto" />

        <Button variant="outline" size="icon" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <ScrollArea fade="mask" className="flex-1 min-h-0 border rounded-md">
        {isLoading ? (
          <div className="flex items-center justify-center h-40">
            <Spinner className="w-6 h-6" />
          </div>
        ) : !data?.items?.length ? (
          <div className="flex flex-col items-center justify-center h-40 text-muted-foreground gap-2">
            <Activity className="w-8 h-8" />
            <span className="text-sm">No traces found</span>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Trace ID</TableHead>
                <TableHead>Root Span</TableHead>
                <TableHead>Start Time</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead className="text-center">Spans</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((trace: TraceSummary) => (
                <TableRow
                  key={trace.traceId}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => setSelectedTraceId(trace.traceId)}
                >
                  <TableCell className="font-mono text-xs">
                    <div className="flex items-center gap-1">
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span>{truncateTraceId(trace.traceId)}</span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <code className="font-mono text-xs">{trace.traceId}</code>
                        </TooltipContent>
                      </Tooltip>
                      <CopyButton
                        value={trace.traceId}
                        tooltipText="Copy Trace ID"
                        size="icon-xs"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </div>
                  </TableCell>
                  <TableCell className="max-w-xs truncate">{trace.rootSpanName}</TableCell>
                  <TableCell className="font-mono text-xs">{formatTimestamp(trace.startTime)}</TableCell>
                  <TableCell className="font-mono text-xs">{formatDuration(trace.durationMs)}</TableCell>
                  <TableCell className="text-center">{trace.spanCount}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </ScrollArea>

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Page {page} of {data.totalPages} ({data.total} total traces)
          </span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= data.totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      <TraceDetailsSheet
        sandboxId={sandboxId}
        traceId={selectedTraceId}
        open={!!selectedTraceId}
        onOpenChange={(open) => !open && setSelectedTraceId(null)}
      />
    </div>
  )
}
