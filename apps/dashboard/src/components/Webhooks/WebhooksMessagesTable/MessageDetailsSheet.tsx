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
import { ChevronDown, ChevronUp, X } from 'lucide-react'
import { MessageOut } from 'svix'
import { MessageAttemptsTable } from '../MessageAttemptsTable'

interface MessageDetailsSheetProps {
  message: MessageOut | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onNavigate: (direction: 'prev' | 'next') => void
  hasPrev: boolean
  hasNext: boolean
}

export function MessageDetailsSheet({
  message,
  open,
  onOpenChange,
  onNavigate,
  hasPrev,
  hasNext,
}: MessageDetailsSheetProps) {
  if (!message) return null

  const hasPayload = message.payload && Object.keys(message.payload).length > 0
  const payload = hasPayload
    ? typeof message.payload === 'string'
      ? message.payload
      : JSON.stringify(message.payload, null, 2)
    : ''
  const { relativeTimeString } = getRelativeTimeString(message.timestamp)

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[520px] p-0 flex flex-col gap-0 [&>button]:hidden" side="right">
        <SheetHeader className="flex flex-row items-center justify-between p-4 px-5 space-y-0">
          <SheetTitle className="text-lg font-medium">Message Details</SheetTitle>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev} onClick={() => onNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous message</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext} onClick={() => onNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next message</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => onOpenChange(false)}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <Separator />

        <div className="flex-1 min-h-0 overflow-auto">
          <div className="flex flex-col px-5 py-4 gap-3">
            <span className="text-base font-medium">Overview</span>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Message ID</span>
              <div className="flex items-center gap-1 group/copy-button">
                <span className="text-sm font-mono">{message.id}</span>
                <CopyButton value={message.id} size="icon-xs" tooltipText="Copy Message ID" />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Event Type</span>
              <Badge variant="secondary">{message.eventType}</Badge>
            </div>
            {message.eventId && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Event ID</span>
                <div className="flex items-center gap-1 group/copy-button">
                  <span className="text-sm font-mono">{message.eventId}</span>
                  <CopyButton value={message.eventId} size="icon-xs" tooltipText="Copy Event ID" />
                </div>
              </div>
            )}
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Timestamp</span>
              <TimestampTooltip
                timestamp={
                  message.timestamp instanceof Date ? message.timestamp.toISOString() : String(message.timestamp)
                }
              >
                <span className="text-sm cursor-default">{relativeTimeString}</span>
              </TimestampTooltip>
            </div>
            {message.channels && message.channels.length > 0 && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Channels</span>
                <div className="flex items-center gap-1 flex-wrap justify-end">
                  {message.channels.map((channel) => (
                    <Badge key={channel} variant="outline" className="font-normal text-xs">
                      {channel}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            {message.tags && message.tags.length > 0 && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Tags</span>
                <div className="flex items-center gap-1 flex-wrap justify-end">
                  {message.tags.map((tag) => (
                    <Badge key={tag} variant="outline" className="font-normal text-xs">
                      {tag}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </div>

          <Separator />

          <div className="flex flex-col px-5 py-4">
            <div className="flex items-center justify-between mb-3">
              <span className="text-base font-medium">Payload</span>
              {hasPayload && <CopyButton value={payload} size="icon-xs" tooltipText="Copy Payload" />}
            </div>
            {hasPayload ? (
              <pre className="text-sm font-mono bg-muted/80 p-3 rounded-md overflow-auto whitespace-pre-wrap break-all">
                {payload}
              </pre>
            ) : (
              <div className="text-sm bg-muted/80 p-3 rounded-md">
                <span className="italic text-muted-foreground">This message has no payload</span>
              </div>
            )}
          </div>

          <Separator />

          <div className="flex flex-col px-5 py-4">
            <MessageAttemptsTable messageId={message.id} />
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
