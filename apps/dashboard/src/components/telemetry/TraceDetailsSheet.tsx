/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo } from 'react'
import { useSandboxTraceSpans } from '@/hooks/useSandboxTraceSpans'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { CopyButton } from '@/components/CopyButton'
import { TraceSpan } from '@daytonaio/api-client'

interface TraceDetailsSheetProps {
  sandboxId: string
  traceId: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface SpanWithDepth extends TraceSpan {
  depth: number
  children: SpanWithDepth[]
}

export const TraceDetailsSheet: React.FC<TraceDetailsSheetProps> = ({ sandboxId, traceId, open, onOpenChange }) => {
  const { data: spans, isLoading } = useSandboxTraceSpans(sandboxId, traceId ?? undefined, {
    enabled: !!traceId && open,
  })

  // Build tree structure from flat span list
  const spanTree = useMemo(() => {
    if (!spans || spans.length === 0) return []

    const spanMap = new Map<string, SpanWithDepth>()
    const roots: SpanWithDepth[] = []

    // First pass: create all spans with depth 0
    spans.forEach((span) => {
      spanMap.set(span.spanId, { ...span, depth: 0, children: [] })
    })

    // Second pass: build tree and calculate depths
    spans.forEach((span) => {
      const spanWithDepth = spanMap.get(span.spanId)!
      if (span.parentSpanId && spanMap.has(span.parentSpanId)) {
        const parent = spanMap.get(span.parentSpanId)!
        parent.children.push(spanWithDepth)
      } else {
        roots.push(spanWithDepth)
      }
    })

    // Third pass: calculate depths
    const setDepths = (spans: SpanWithDepth[], depth: number) => {
      spans.forEach((span) => {
        span.depth = depth
        setDepths(span.children, depth + 1)
      })
    }
    setDepths(roots, 0)

    // Flatten tree in order
    const result: SpanWithDepth[] = []
    const flatten = (spans: SpanWithDepth[]) => {
      // Sort by timestamp
      spans.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
      spans.forEach((span) => {
        result.push(span)
        flatten(span.children)
      })
    }
    flatten(roots)

    return result
  }, [spans])

  // Calculate trace start time and total duration for waterfall
  const { traceStart, traceDuration } = useMemo(() => {
    if (!spans || spans.length === 0) return { traceStart: 0, traceDuration: 0 }

    const times = spans.map((s) => new Date(s.timestamp).getTime())
    const durations = spans.map((s, i) => times[i] + s.durationNs / 1_000_000)

    const start = Math.min(...times)
    const end = Math.max(...durations)

    return { traceStart: start, traceDuration: end - start }
  }, [spans])

  const formatDuration = (durationNs: number) => {
    const ms = durationNs / 1_000_000
    if (ms < 1) {
      return `${(ms * 1000).toFixed(0)}us`
    }
    if (ms < 1000) {
      return `${ms.toFixed(2)}ms`
    }
    return `${(ms / 1000).toFixed(2)}s`
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[700px] p-0 flex flex-col gap-0">
        <SheetHeader className="p-4 border-b">
          <SheetTitle className="flex items-center gap-2">
            Trace Details
            {traceId && (
              <>
                <code className="text-sm font-mono text-muted-foreground font-normal">{traceId}</code>
                <CopyButton value={traceId} tooltipText="Copy Trace ID" size="icon-xs" />
              </>
            )}
          </SheetTitle>
        </SheetHeader>

        <ScrollArea className="flex-1">
          {isLoading ? (
            <div className="flex items-center justify-center h-40">
              <Spinner className="w-6 h-6" />
            </div>
          ) : !spanTree.length ? (
            <div className="flex items-center justify-center h-40 text-muted-foreground">
              <span className="text-sm">No spans found</span>
            </div>
          ) : (
            <div className="p-4 space-y-2">
              {spanTree.map((span) => {
                const spanStart = new Date(span.timestamp).getTime()
                const spanDuration = span.durationNs / 1_000_000
                const offsetPercent = traceDuration > 0 ? ((spanStart - traceStart) / traceDuration) * 100 : 0
                const widthPercent = traceDuration > 0 ? (spanDuration / traceDuration) * 100 : 100

                return (
                  <div key={span.spanId} className="group">
                    <div className="flex items-center gap-2" style={{ paddingLeft: `${span.depth * 16}px` }}>
                      <div className="flex-shrink-0 w-48 truncate text-sm">{span.spanName}</div>
                      <div className="flex-1 h-6 bg-muted rounded relative overflow-hidden">
                        <div
                          className="absolute h-full bg-primary/70 rounded transition-all group-hover:bg-primary"
                          style={{
                            left: `${Math.min(offsetPercent, 99)}%`,
                            width: `${Math.max(widthPercent, 1)}%`,
                          }}
                        />
                      </div>
                      <div className="flex-shrink-0 w-20 text-right text-xs font-mono text-muted-foreground">
                        {formatDuration(span.durationNs)}
                      </div>
                    </div>

                    <div
                      className="hidden group-hover:block mt-2 p-3 bg-muted/50 rounded text-xs"
                      style={{ marginLeft: `${span.depth * 16}px` }}
                    >
                      <div className="grid grid-cols-2 gap-2">
                        <div>
                          <span className="text-muted-foreground">Span ID:</span>
                          <code className="ml-1 font-mono">{span.spanId.slice(0, 16)}</code>
                        </div>
                        {span.parentSpanId && (
                          <div>
                            <span className="text-muted-foreground">Parent:</span>
                            <code className="ml-1 font-mono">{span.parentSpanId.slice(0, 16)}</code>
                          </div>
                        )}
                        {span.statusCode && (
                          <div>
                            <span className="text-muted-foreground">Status:</span>
                            <span className="ml-1">{span.statusCode}</span>
                          </div>
                        )}
                      </div>
                      {Object.keys(span.spanAttributes || {}).length > 0 && (
                        <div className="mt-2">
                          <span className="text-muted-foreground">Attributes:</span>
                          <pre className="mt-1 p-2 bg-background rounded overflow-x-auto">
                            {JSON.stringify(span.spanAttributes, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
