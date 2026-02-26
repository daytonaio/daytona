/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback, useMemo } from 'react'
import { useQueryStates } from 'nuqs'
import { useSandboxTraces, TracesQueryParams } from '@/hooks/useSandboxTraces'
import { useSandboxTraceSpans } from '@/hooks/useSandboxTraceSpans'
import { TimeRangeSelector } from '@/components/telemetry/TimeRangeSelector'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { CopyButton } from '@/components/CopyButton'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { ChevronLeft, ChevronRight, RefreshCw, Activity, ChevronDown } from 'lucide-react'
import { format, subHours } from 'date-fns'
import { TraceSummary, TraceSpan } from '@daytonaio/api-client'
import { cn } from '@/lib/utils'
import { tracesSearchParams, timeRangeSearchParams } from './SearchParams'

interface SpanNode extends TraceSpan {
  depth: number
  children: SpanNode[]
}

function buildSpanTree(spans: TraceSpan[]): SpanNode[] {
  const map = new Map<string, SpanNode>()
  const roots: SpanNode[] = []

  for (const span of spans) {
    map.set(span.spanId, { ...span, depth: 0, children: [] })
  }
  for (const span of spans) {
    const node = map.get(span.spanId)!
    if (span.parentSpanId && map.has(span.parentSpanId)) {
      map.get(span.parentSpanId)!.children.push(node)
    } else {
      roots.push(node)
    }
  }

  const assignDepths = (nodes: SpanNode[], depth: number) => {
    for (const n of nodes) {
      n.depth = depth
      assignDepths(n.children, depth + 1)
    }
  }
  assignDepths(roots, 0)

  const flat: SpanNode[] = []
  const walk = (nodes: SpanNode[]) => {
    nodes.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
    for (const n of nodes) {
      flat.push(n)
      walk(n.children)
    }
  }
  walk(roots)
  return flat
}

function formatNsDuration(ns: number) {
  const ms = ns / 1_000_000
  if (ms < 1) return `${(ms * 1000).toFixed(0)}µs`
  if (ms < 1000) return `${ms.toFixed(2)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatMsDuration(ms: number) {
  if (ms < 1) return `${(ms * 1000).toFixed(2)}µs`
  if (ms < 1000) return `${ms.toFixed(2)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatTimestamp(timestamp: string) {
  try {
    return format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')
  } catch {
    return timestamp
  }
}

function truncateId(id: string) {
  return id.length > 16 ? `${id.slice(0, 8)}…${id.slice(-8)}` : id
}

function statusColor(code?: string): string {
  if (!code) return 'bg-blue-500/70'
  const c = code.toUpperCase()
  if (c === 'ERROR' || c === 'STATUS_CODE_ERROR') return 'bg-destructive/70'
  if (c === 'OK' || c === 'STATUS_CODE_OK') return 'bg-emerald-500/70'
  return 'bg-blue-500/70'
}

function TracesTableSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-10" />
          <TableHead>Trace ID</TableHead>
          <TableHead>Root Span</TableHead>
          <TableHead>Start Time</TableHead>
          <TableHead>Duration</TableHead>
          <TableHead className="text-center">Spans</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 8 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="size-4" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-28" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4" style={{ width: `${40 + (i % 3) * 15}%` }} />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-36" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-16" />
            </TableCell>
            <TableCell className="text-center">
              <Skeleton className="h-4 w-8 mx-auto" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function TracesErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyTitle>Failed to load traces</EmptyTitle>
        <EmptyDescription>Something went wrong while fetching traces.</EmptyDescription>
      </EmptyHeader>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="size-4" />
        Retry
      </Button>
    </Empty>
  )
}

function TracesEmptyState() {
  return (
    <Empty className="flex-1 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Activity className="size-4" />
        </EmptyMedia>
        <EmptyTitle>No traces found</EmptyTitle>
        <EmptyDescription>
          Try adjusting your time range.{' '}
          <a href={`${DAYTONA_DOCS_URL}/en/experimental/otel-collection`} target="_blank" rel="noopener noreferrer">
            Learn more about observability
          </a>
          .
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

function DetailRow({ label, children, mono = true }: { label: string; children: React.ReactNode; mono?: boolean }) {
  return (
    <div>
      <span className="text-muted-foreground text-[10px] uppercase tracking-wider font-semibold">{label}</span>
      <div className={cn('flex items-center gap-1 mt-0.5', mono && 'font-mono')}>{children}</div>
    </div>
  )
}

function TraceExpandedRow({ sandboxId, trace }: { sandboxId: string; trace: TraceSummary }) {
  const { data: spans, isLoading, isError, refetch } = useSandboxTraceSpans(sandboxId, trace.traceId)
  const [selectedSpanId, setSelectedSpanId] = useState<string | null>(null)

  const spanTree = useMemo(() => (spans ? buildSpanTree(spans) : []), [spans])

  const { traceStart, traceDuration } = useMemo(() => {
    if (!spans?.length) return { traceStart: 0, traceDuration: 0 }
    const starts = spans.map((s) => new Date(s.timestamp).getTime())
    const ends = spans.map((s, i) => starts[i] + s.durationNs / 1_000_000)
    const start = Math.min(...starts)
    return { traceStart: start, traceDuration: Math.max(...ends) - start }
  }, [spans])

  const selectedSpan = useMemo(
    () => spanTree.find((s) => s.spanId === selectedSpanId) ?? null,
    [spanTree, selectedSpanId],
  )

  if (isLoading) {
    return (
      <div className="flex h-[340px] border-t border-border">
        <div className="w-1/2 border-r border-border flex flex-col overflow-hidden">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-muted/30 shrink-0">
            <Skeleton className="h-3 w-10" />
          </div>
          <div className="py-1 flex flex-col gap-0.5">
            {Array.from({ length: 8 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center gap-2 px-3 py-1"
                style={{ paddingLeft: `${12 + (i % 3) * 14}px` }}
              >
                <Skeleton className="shrink-0 h-3 w-[120px]" />
                <Skeleton className="flex-1 h-4 rounded min-w-[60px]" />
                <Skeleton className="shrink-0 h-3 w-[56px]" />
              </div>
            ))}
          </div>
        </div>
        <div className="w-1/2 flex flex-col overflow-hidden">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-muted/30 shrink-0">
            <span className="text-xs font-medium text-muted-foreground">Select a span</span>
          </div>
          <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
            Click a span to see details
          </div>
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 h-[340px] text-muted-foreground">
        <p className="text-sm">Failed to load spans.</p>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="size-4" />
          Retry
        </Button>
      </div>
    )
  }

  if (!spanTree.length) {
    return (
      <div className="flex items-center justify-center h-48 text-muted-foreground text-sm">
        No spans found for this trace.
      </div>
    )
  }

  return (
    <div className="flex h-[340px] border-t border-border">
      <div className="w-1/2 border-r border-border flex flex-col overflow-hidden">
        <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-muted/30 shrink-0">
          <span className="text-xs font-medium text-muted-foreground">Spans</span>
          <span className="text-xs text-muted-foreground">({spanTree.length})</span>
        </div>
        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="py-1">
            {spanTree.map((span) => {
              const spanStartMs = new Date(span.timestamp).getTime()
              const spanDurMs = span.durationNs / 1_000_000
              const offsetPct = traceDuration > 0 ? ((spanStartMs - traceStart) / traceDuration) * 100 : 0
              const widthPct = traceDuration > 0 ? (spanDurMs / traceDuration) * 100 : 100
              const isSelected = selectedSpanId === span.spanId

              return (
                <button
                  key={span.spanId}
                  type="button"
                  onClick={() => setSelectedSpanId(span.spanId)}
                  className={cn(
                    'w-full text-left flex items-center gap-2 px-3 py-1 hover:bg-muted/50 transition-colors',
                    isSelected && 'bg-muted',
                  )}
                  style={{ paddingLeft: `${12 + span.depth * 14}px` }}
                >
                  <span className="shrink-0 w-[120px] truncate text-xs">{span.spanName}</span>
                  <div className="flex-1 h-4 bg-muted/60 rounded relative overflow-hidden min-w-[60px]">
                    <div
                      className={cn('absolute h-full rounded', statusColor(span.statusCode))}
                      style={{
                        left: `${Math.min(offsetPct, 99)}%`,
                        width: `${Math.max(widthPct, 1)}%`,
                      }}
                    />
                  </div>
                  <span className="shrink-0 text-[10px] font-mono text-muted-foreground w-[56px] text-right">
                    {formatNsDuration(span.durationNs)}
                  </span>
                </button>
              )
            })}
          </div>
        </ScrollArea>
      </div>

      <div className="w-1/2 flex flex-col overflow-hidden">
        <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-muted/30 shrink-0">
          <span className="text-xs font-medium text-muted-foreground">
            {selectedSpan ? 'Span Detail' : 'Select a span'}
          </span>
        </div>
        {selectedSpan ? (
          <ScrollArea fade="mask" className="flex-1 min-h-0">
            <div className="p-3 space-y-3 text-xs">
              <DetailRow label="Name" mono={false}>
                {selectedSpan.spanName}
              </DetailRow>
              <DetailRow label="Span ID">
                <span className="font-mono">{selectedSpan.spanId}</span>
                <CopyButton value={selectedSpan.spanId} tooltipText="Copy" size="icon-xs" />
              </DetailRow>
              {selectedSpan.parentSpanId && (
                <DetailRow label="Parent Span ID">
                  <span className="font-mono">{selectedSpan.parentSpanId}</span>
                  <CopyButton value={selectedSpan.parentSpanId} tooltipText="Copy" size="icon-xs" />
                </DetailRow>
              )}
              <DetailRow label="Start">
                <span className="font-mono">{formatTimestamp(selectedSpan.timestamp)}</span>
              </DetailRow>
              <DetailRow label="Duration">
                <span className="font-mono">{formatNsDuration(selectedSpan.durationNs)}</span>
              </DetailRow>
              {selectedSpan.statusCode && (
                <DetailRow label="Status">
                  <span
                    className={cn(
                      'inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium',
                      selectedSpan.statusCode.toUpperCase().includes('ERROR')
                        ? 'bg-destructive/10 text-destructive'
                        : selectedSpan.statusCode.toUpperCase().includes('OK')
                          ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                          : 'bg-muted text-muted-foreground',
                    )}
                  >
                    {selectedSpan.statusCode}
                  </span>
                  {selectedSpan.statusMessage && (
                    <span className="text-muted-foreground ml-2">{selectedSpan.statusMessage}</span>
                  )}
                </DetailRow>
              )}
              {Object.keys(selectedSpan.spanAttributes ?? {}).length > 0 && (
                <div>
                  <span className="text-muted-foreground text-[10px] uppercase tracking-wider font-semibold">
                    Attributes
                  </span>
                  <div className="mt-1.5 rounded border border-border overflow-hidden">
                    <table className="w-full text-xs">
                      <tbody>
                        {Object.entries(selectedSpan.spanAttributes).map(([key, val]) => (
                          <tr key={key} className="border-b border-border last:border-b-0">
                            <td className="px-2 py-1 text-muted-foreground font-mono bg-muted/30 w-[40%] align-middle break-all">
                              {key}
                            </td>
                            <td className="px-2 py-1 font-mono align-middle break-all">{val}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
            Click a span to see details
          </div>
        )}
      </div>
    </div>
  )
}

export function SandboxTracesTab({ sandboxId }: { sandboxId: string }) {
  const [params, setParams] = useQueryStates(tracesSearchParams)
  const [timeRange, setTimeRange] = useQueryStates(timeRangeSearchParams)
  const [expandedTraceId, setExpandedTraceId] = useState<string | null>(null)
  const limit = 50

  const resolvedFrom = useMemo(() => timeRange.from ?? subHours(new Date(), 1), [timeRange.from])
  const resolvedTo = useMemo(() => timeRange.to ?? new Date(), [timeRange.to])

  const queryParams: TracesQueryParams = useMemo(
    () => ({
      from: resolvedFrom,
      to: resolvedTo,
      page: params.tracesPage,
      limit,
    }),
    [resolvedFrom, resolvedTo, params.tracesPage],
  )

  const { data, isLoading, isError, refetch } = useSandboxTraces(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback(
    (from: Date, to: Date) => {
      setTimeRange({ from, to })
      setParams({ tracesPage: 1 })
    },
    [setTimeRange, setParams],
  )

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
        <div className="flex-1 min-h-0 border rounded-md">
          <TracesTableSkeleton />
        </div>
      ) : isError ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <TracesErrorState onRetry={() => refetch()} />
        </div>
      ) : !data?.items?.length ? (
        <div className="flex-1 min-h-0 border rounded-md flex">
          <TracesEmptyState />
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
                <TableHead>Trace ID</TableHead>
                <TableHead>Root Span</TableHead>
                <TableHead>Start Time</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead className="text-center">Spans</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((trace: TraceSummary) => {
                const isExpanded = expandedTraceId === trace.traceId
                return (
                  <React.Fragment key={trace.traceId}>
                    <TableRow
                      className="cursor-pointer hover:bg-muted/50 group/trace-row"
                      onClick={() => setExpandedTraceId(isExpanded ? null : trace.traceId)}
                    >
                      <TableCell className="w-10 px-2">
                        <ChevronDown
                          className={cn(
                            'size-4 text-muted-foreground transition-transform duration-200',
                            isExpanded && 'rotate-180',
                          )}
                        />
                      </TableCell>
                      <TableCell className="font-mono text-xs whitespace-nowrap">
                        <div className="flex items-center gap-1">
                          <span>{truncateId(trace.traceId)}</span>
                          <CopyButton
                            value={trace.traceId}
                            tooltipText="Copy Trace ID"
                            size="icon-xs"
                            className="[@media(hover:hover)]:opacity-0 [@media(hover:hover)]:group-hover/trace-row:opacity-100 transition-opacity"
                            onClick={(e) => e.stopPropagation()}
                          />
                        </div>
                      </TableCell>
                      <TableCell className="max-w-xs truncate">{trace.rootSpanName}</TableCell>
                      <TableCell className="font-mono text-xs">{formatTimestamp(trace.startTime)}</TableCell>
                      <TableCell className="font-mono text-xs">{formatMsDuration(trace.durationMs)}</TableCell>
                      <TableCell className="text-center">{trace.spanCount}</TableCell>
                    </TableRow>
                    {isExpanded && (
                      <TableRow>
                        <TableCell colSpan={6} className="p-0">
                          <TraceExpandedRow sandboxId={sandboxId} trace={trace} />
                        </TableCell>
                      </TableRow>
                    )}
                  </React.Fragment>
                )
              })}
            </TableBody>
          </Table>
        </ScrollArea>
      )}

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between shrink-0">
          <span className="text-sm text-muted-foreground">
            Page {params.tracesPage} of {data.totalPages} ({data.total} total)
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={params.tracesPage <= 1}
              onClick={() => setParams({ tracesPage: params.tracesPage - 1 })}
            >
              <ChevronLeft className="size-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={params.tracesPage >= data.totalPages}
              onClick={() => setParams({ tracesPage: params.tracesPage + 1 })}
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
