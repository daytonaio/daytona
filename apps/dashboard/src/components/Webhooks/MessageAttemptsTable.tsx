/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { getRelativeTimeString } from '@/lib/utils'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronDown, ChevronRight, LoaderCircle } from 'lucide-react'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'
import { MessageAttemptOut } from 'svix'
import { useMessageAttempts } from 'svix-react'

const RELOAD_DELAY = 5

function AttemptStatusBadge({ status }: { status: number }) {
  const variant = status === 0 ? 'success' : status === 1 ? 'secondary' : 'destructive'
  const label = status === 0 ? 'Success' : status === 1 ? 'Pending' : status === 3 ? 'Sending' : 'Failed'
  return <Badge variant={variant}>{label}</Badge>
}

function TriggerTypeBadge({ triggerType }: { triggerType: number }) {
  return (
    <Badge variant="outline" className="font-normal text-xs">
      {triggerType === 1 ? 'Manual' : 'Scheduled'}
    </Badge>
  )
}

function AttemptExpandedRow({ attempt }: { attempt: MessageAttemptOut }) {
  let responseBody: string
  try {
    const parsed = JSON.parse(attempt.response)
    responseBody = JSON.stringify(parsed, null, 2)
  } catch {
    responseBody = attempt.response || '(empty)'
  }

  return (
    <div className="flex flex-col gap-3 px-3 py-3">
      <div className="flex flex-col gap-2 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Status Code</span>
          <Badge
            variant={attempt.responseStatusCode >= 200 && attempt.responseStatusCode < 300 ? 'success' : 'destructive'}
          >
            {attempt.responseStatusCode}
          </Badge>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Duration</span>
          <span className="font-mono">{attempt.responseDurationMs}ms</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Trigger</span>
          <TriggerTypeBadge triggerType={attempt.triggerType} />
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Endpoint ID</span>
          <div className="flex items-center gap-1 group/copy-button">
            <span className="font-mono truncate max-w-[120px]">{attempt.endpointId}</span>
            <CopyButton value={attempt.endpointId} size="icon-xs" tooltipText="Copy Endpoint ID" />
          </div>
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-sm text-muted-foreground">Response Body</span>
          <CopyButton value={attempt.response || ''} size="icon-xs" tooltipText="Copy Response" />
        </div>
        <pre className="text-xs font-mono bg-muted/80 p-2.5 rounded-md overflow-auto whitespace-pre-wrap break-all max-h-[200px]">
          {responseBody}
        </pre>
      </div>
    </div>
  )
}

export function MessageAttemptsTable({ messageId, reloadKey }: { messageId: string; reloadKey?: number }) {
  const attempts = useMessageAttempts(messageId)
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set())
  const [countdown, setCountdown] = useState<number | null>(null)
  const prevReloadKey = useRef(reloadKey)
  const timerRef = useRef<ReturnType<typeof setInterval>>(null)

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
  }, [])

  useEffect(() => {
    if (reloadKey !== undefined && reloadKey !== prevReloadKey.current) {
      prevReloadKey.current = reloadKey
      clearTimer()
      setCountdown(RELOAD_DELAY)

      timerRef.current = setInterval(() => {
        setCountdown((prev) => {
          if (prev === null || prev <= 1) {
            clearTimer()
            attempts.reload()
            return null
          }
          return prev - 1
        })
      }, 1000)
    }
  }, [reloadKey, attempts, clearTimer])

  useEffect(() => {
    return clearTimer
  }, [clearTimer])

  const toggleRow = (attemptId: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev)
      if (next.has(attemptId)) {
        next.delete(attemptId)
      } else {
        next.add(attemptId)
      }
      return next
    })
  }

  const header = (
    <div className="flex items-center gap-2 mb-3">
      <span className="text-base font-medium">Delivery Attempts</span>
      <AnimatePresence>
        {countdown !== null && (
          <motion.span
            key="countdown"
            initial={{ y: 10, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: -10, opacity: 0 }}
            transition={{ duration: 0.15, ease: 'easeOut' }}
            className="flex items-center gap-1.5 text-xs text-muted-foreground"
          >
            <LoaderCircle className="size-3 animate-spin" />
            Reloading in {countdown}...
          </motion.span>
        )}
      </AnimatePresence>
    </div>
  )

  if (attempts.loading) {
    return (
      <div>
        {header}
        <div className="space-y-2">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
        </div>
      </div>
    )
  }

  if (attempts.error) {
    return (
      <div>
        {header}
        <div className="text-sm text-muted-foreground">Failed to load message attempts.</div>
      </div>
    )
  }

  const data = attempts.data ?? []

  if (data.length === 0) {
    return (
      <div>
        {header}
        <div className="text-sm text-muted-foreground">No delivery attempts yet.</div>
      </div>
    )
  }

  return (
    <div>
      {header}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="px-3 w-[28px]" />
              <TableHead className="px-3">Status</TableHead>
              <TableHead className="px-3">URL</TableHead>
              <TableHead className="px-3">Timestamp</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((attempt: MessageAttemptOut) => {
              const { relativeTimeString } = getRelativeTimeString(attempt.timestamp)
              const isExpanded = expandedRows.has(attempt.id)
              return (
                <Fragment key={attempt.id}>
                  <TableRow className="cursor-pointer hover:bg-muted/50" onClick={() => toggleRow(attempt.id)}>
                    <TableCell className="px-3 w-[28px]">
                      {isExpanded ? (
                        <ChevronDown className="size-3.5 text-muted-foreground" />
                      ) : (
                        <ChevronRight className="size-3.5 text-muted-foreground" />
                      )}
                    </TableCell>
                    <TableCell className="px-3">
                      <AttemptStatusBadge status={attempt.status} />
                    </TableCell>
                    <TableCell className="px-3">
                      <span className="text-sm font-mono truncate block max-w-[200px]">{attempt.url}</span>
                    </TableCell>
                    <TableCell className="px-3">
                      <TimestampTooltip
                        timestamp={
                          attempt.timestamp instanceof Date
                            ? attempt.timestamp.toISOString()
                            : String(attempt.timestamp)
                        }
                      >
                        <span className="text-sm cursor-default">{relativeTimeString}</span>
                      </TimestampTooltip>
                    </TableCell>
                  </TableRow>
                  {isExpanded && (
                    <TableRow className="hover:bg-transparent">
                      <TableCell colSpan={4} className="p-0 border-b">
                        <AttemptExpandedRow attempt={attempt} />
                      </TableCell>
                    </TableRow>
                  )}
                </Fragment>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
