/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { getRelativeTimeString } from '@/lib/utils'
import { ChevronDown, ChevronUp, RefreshCw, X } from 'lucide-react'
import { EndpointMessageOut } from 'svix'

interface EventDetailsSheetProps {
  event: EndpointMessageOut | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onNavigate: (direction: 'prev' | 'next') => void
  hasPrev: boolean
  hasNext: boolean
  onReplay: (msgId: string) => void
}

export function EventDetailsSheet({
  event,
  open,
  onOpenChange,
  onNavigate,
  hasPrev,
  hasNext,
  onReplay,
}: EventDetailsSheetProps) {
  if (!event) return null

  const hasPayload = event.payload && Object.keys(event.payload).length > 0
  const payload = hasPayload
    ? typeof event.payload === 'string'
      ? event.payload
      : JSON.stringify(event.payload, null, 2)
    : ''
  const { relativeTimeString } = getRelativeTimeString(event.timestamp)

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[480px] p-0 flex flex-col gap-0 [&>button]:hidden" side="right">
        <SheetHeader className="flex flex-row items-center justify-between p-4 px-5 space-y-0">
          <SheetTitle className="text-lg font-medium">Event Details</SheetTitle>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev} onClick={() => onNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous event</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext} onClick={() => onNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next event</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => onOpenChange(false)}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <Separator />

        <div className="flex flex-col px-5 py-4 gap-3">
          <span className="text-lg font-medium">Overview</span>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Message ID</span>
            <div className="flex items-center gap-1 group/copy-button">
              <span className="text-sm font-mono">{event.id}</span>
              <CopyButton value={event.id} size="icon-xs" tooltipText="Copy Message ID" />
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Status</span>
            <Badge variant={event.status === 0 ? 'success' : event.status === 1 ? 'secondary' : 'destructive'}>
              {event.status === 0 ? 'Success' : event.status === 1 ? 'Pending' : 'Failed'}
            </Badge>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Event Type</span>
            <Badge variant="secondary">{event.eventType}</Badge>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Sent</span>
            <TimestampTooltip
              timestamp={event.timestamp instanceof Date ? event.timestamp.toISOString() : String(event.timestamp)}
            >
              <span className="text-sm cursor-default">{relativeTimeString}</span>
            </TimestampTooltip>
          </div>
          <Button variant="outline" size="sm" className="w-full mt-1" onClick={() => onReplay(event.id)}>
            <RefreshCw className="size-3.5 mr-1.5" />
            Replay
          </Button>
        </div>

        <Separator />

        <div className="flex-1 flex flex-col min-h-0">
          <div className="flex items-center justify-between px-5 py-3">
            <span className="text-lg font-medium">Payload</span>
            {hasPayload && <CopyButton value={payload} size="icon-xs" tooltipText="Copy Payload" />}
          </div>
          <div className="flex-1 min-h-0 overflow-auto px-5 pb-5">
            {hasPayload ? (
              <pre className="text-sm font-mono bg-muted/80 p-3 rounded-md overflow-auto whitespace-pre-wrap break-all">
                {payload}
              </pre>
            ) : (
              <div className="text-sm bg-muted/80 p-3 rounded-md">
                <span className="italic text-muted-foreground">This event has no payload</span>
              </div>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
