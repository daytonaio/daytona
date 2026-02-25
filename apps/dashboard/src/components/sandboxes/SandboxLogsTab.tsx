/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback, useMemo } from 'react'
import { useQueryStates } from 'nuqs'
import { useSandboxLogs, LogsQueryParams } from '@/hooks/useSandboxLogs'
import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { SeverityBadge } from '@/components/telemetry/SeverityBadge'
import { CopyButton } from '@/components/CopyButton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { ChevronLeft, ChevronRight, Search, FileText, RefreshCw, ChevronDown } from 'lucide-react'
import { format, subHours } from 'date-fns'
import { LogEntry } from '@daytonaio/api-client'
import { cn } from '@/lib/utils'
import { logsSearchParams, SEVERITY_OPTIONS, timeRangeSearchParams } from './SearchParams'

function formatTimestamp(timestamp: string) {
  try {
    return format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')
  } catch {
    return timestamp
  }
}

function LogsTableSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-10" />
          <TableHead className="w-48">Timestamp</TableHead>
          <TableHead className="w-24">Severity</TableHead>
          <TableHead>Message</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 10 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="size-4" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-36" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-5 w-14 rounded-full" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4" style={{ width: `${45 + (i % 4) * 12}%` }} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function LogsErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyTitle>Failed to load logs</EmptyTitle>
        <EmptyDescription>Something went wrong while fetching logs.</EmptyDescription>
      </EmptyHeader>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="size-4" />
        Retry
      </Button>
    </Empty>
  )
}

function LogsEmptyState() {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <FileText className="size-4" />
        </EmptyMedia>
        <EmptyTitle>No logs found</EmptyTitle>
        <EmptyDescription>
          Try adjusting your time range or filters.{' '}
          <a href={`${DAYTONA_DOCS_URL}/en/experimental/otel-collection`} target="_blank" rel="noopener noreferrer">
            Learn more about observability
          </a>
          .
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function SandboxLogsTab({ sandboxId }: { sandboxId: string }) {
  const [params, setParams] = useQueryStates(logsSearchParams)
  const [timeRange, setTimeRange] = useQueryStates(timeRangeSearchParams)
  const [searchInput, setSearchInput] = useState(params.search)
  const [expandedRow, setExpandedRow] = useState<number | null>(null)
  const limit = 50

  const resolvedFrom = useMemo(() => timeRange.from ?? subHours(new Date(), 1), [timeRange.from])
  const resolvedTo = useMemo(() => timeRange.to ?? new Date(), [timeRange.to])

  const queryParams: LogsQueryParams = useMemo(
    () => ({
      from: resolvedFrom,
      to: resolvedTo,
      page: params.logsPage,
      limit,
      severities: params.severity.length > 0 ? [...params.severity] : undefined,
      search: params.search || undefined,
    }),
    [resolvedFrom, resolvedTo, params.logsPage, params.severity, params.search],
  )

  const { data, isLoading, isError, refetch } = useSandboxLogs(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback(
    (from: Date, to: Date) => {
      setTimeRange({ from, to })
      setParams({ logsPage: 1 })
    },
    [setTimeRange, setParams],
  )

  const handleSearch = useCallback(() => {
    setParams({ search: searchInput, logsPage: 1 })
  }, [searchInput, setParams])

  const handleSeverityChange = useCallback(
    (value: string) => {
      if (value === 'all' || !value) {
        setParams({ severity: [], logsPage: 1 })
      } else {
        setParams({ severity: [value as (typeof SEVERITY_OPTIONS)[number]], logsPage: 1 })
      }
    },
    [setParams],
  )

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3 shrink-0">
        <TimeRangeSelector
          onChange={handleTimeRangeChange}
          defaultRange={timeRange.from && timeRange.to ? { from: timeRange.from, to: timeRange.to } : undefined}
          className="w-auto"
        />

        <div className="flex items-center gap-2">
          <Input
            placeholder="Search logs..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="w-48"
          />
          <Button variant="outline" size="icon-sm" onClick={handleSearch}>
            <Search className="size-4" />
          </Button>
        </div>

        <Select value={params.severity.length === 1 ? params.severity[0] : ''} onValueChange={handleSeverityChange}>
          <SelectTrigger className="w-32" size="sm">
            <SelectValue placeholder="Severity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            {SEVERITY_OPTIONS.map((sev) => (
              <SelectItem key={sev} value={sev}>
                {sev}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button variant="ghost" size="icon-sm" onClick={() => refetch()} className="ml-auto">
          <RefreshCw className="size-4" />
        </Button>
      </div>

      {isLoading ? (
        <div className="flex-1 min-h-0 border rounded-md">
          <LogsTableSkeleton />
        </div>
      ) : isError ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <LogsErrorState onRetry={() => refetch()} />
        </div>
      ) : !data?.items?.length ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <LogsEmptyState />
        </div>
      ) : (
        <ScrollArea
          fade="mask"
          horizontal
          className="flex-1 min-h-0 border rounded-md [&_[data-slot=scroll-area-viewport]>div]:!overflow-visible [&_[data-slot=scroll-area-viewport]>div>div]:!overflow-visible"
        >
          <Table>
            <TableHeader className="sticky top-0 z-10 bg-background after:absolute after:bottom-0 after:left-0 after:right-0 after:h-px after:bg-border">
              <TableRow>
                <TableHead className="w-10" />
                <TableHead className="w-48">Timestamp</TableHead>
                <TableHead className="w-24">Severity</TableHead>
                <TableHead>Message</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((log: LogEntry, index: number) => (
                <React.Fragment key={index}>
                  <TableRow
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => setExpandedRow(expandedRow === index ? null : index)}
                  >
                    <TableCell>
                      <ChevronDown
                        className={cn(
                          'size-4 transition-transform duration-200',
                          expandedRow === index && 'rotate-180',
                        )}
                      />
                    </TableCell>
                    <TableCell className="font-mono text-xs">{formatTimestamp(log.timestamp)}</TableCell>
                    <TableCell>
                      <SeverityBadge severity={log.severityText} />
                    </TableCell>
                    <TableCell className="max-w-md truncate font-mono text-xs">{log.body}</TableCell>
                  </TableRow>
                  {expandedRow === index && (
                    <TableRow>
                      <TableCell colSpan={4} className="bg-muted/30 p-4">
                        <div className="space-y-3">
                          <div>
                            <h4 className="text-sm font-medium mb-1">Full Message</h4>
                            <pre className="text-xs bg-background p-2 rounded overflow-x-auto whitespace-pre-wrap">
                              {log.body}
                            </pre>
                          </div>
                          {log.traceId && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Trace ID</h4>
                              <code className="text-xs bg-background p-1 rounded">{log.traceId}</code>
                            </div>
                          )}
                          {log.spanId && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Span ID</h4>
                              <code className="text-xs bg-background p-1 rounded">{log.spanId}</code>
                            </div>
                          )}
                          {Object.keys(log.logAttributes || {}).length > 0 && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Attributes</h4>
                              <div className="relative">
                                <CopyButton
                                  value={JSON.stringify(log.logAttributes, null, 2)}
                                  tooltipText="Copy"
                                  size="icon-xs"
                                  className="absolute top-1.5 right-1.5"
                                />
                                <pre className="text-xs bg-background p-2 rounded overflow-x-auto">
                                  {JSON.stringify(log.logAttributes, null, 2)}
                                </pre>
                              </div>
                            </div>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </React.Fragment>
              ))}
            </TableBody>
          </Table>
        </ScrollArea>
      )}

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between shrink-0">
          <span className="text-sm text-muted-foreground">
            Page {params.logsPage} of {data.totalPages} ({data.total} total)
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={params.logsPage <= 1}
              onClick={() => setParams({ logsPage: params.logsPage - 1 })}
            >
              <ChevronLeft className="size-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={params.logsPage >= data.totalPages}
              onClick={() => setParams({ logsPage: params.logsPage + 1 })}
            >
              Next
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
