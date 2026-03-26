/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { format, formatDistanceToNow } from 'date-fns'
import { ReactNode } from 'react'
import { Separator } from './ui/separator'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

interface TimestampTooltipProps {
  timestamp?: string
  children: ReactNode
  time?: boolean
}

export const TimestampTooltip = ({ children, timestamp, time = true }: TimestampTooltipProps) => {
  if (!timestamp) {
    return children
  }

  const date = new Date(timestamp)
  const relativeTimeString = formatDistanceToNow(date, { addSuffix: true })

  const dateFormat = 'MMM d, yyyy'
  const timeFormat = 'HH:mm:ss'

  const utcDate = new Date(
    date.getUTCFullYear(),
    date.getUTCMonth(),
    date.getUTCDate(),
    date.getUTCHours(),
    date.getUTCMinutes(),
    date.getUTCSeconds(),
  )
  const utcDateFormatted = format(utcDate, dateFormat)
  const utcTimeFormatted = format(utcDate, timeFormat)

  const localDateFormatted = format(date, dateFormat)
  const localTimeFormatted = format(date, timeFormat)

  const timezoneFormatter = new Intl.DateTimeFormat('en-US', {
    timeZoneName: 'short',
  })
  const timezoneParts = timezoneFormatter.formatToParts(date)
  const localTimezone = timezoneParts.find((part) => part.type === 'timeZoneName')?.value || 'Local'

  return (
    <Tooltip>
      <TooltipTrigger>{children}</TooltipTrigger>
      <TooltipContent className="flex flex-col gap-1.5 text-xs">
        <div className="font-medium first-letter:capitalize">{relativeTimeString}</div>
        <Separator className="-mx-3 w-[calc(100%+1.5rem)]" />
        <table className="border-collapse border-0">
          <tbody>
            <tr>
              <td className="text-muted-foreground pr-2 border-0">[UTC]</td>
              <td className="border-0 pr-2">{utcDateFormatted}</td>
              {time && <td className="border-0 text-muted-foreground">{utcTimeFormatted}</td>}
            </tr>
            <tr>
              <td className="text-muted-foreground pr-2 border-0">[{localTimezone}]</td>
              <td className="border-0 pr-2">{localDateFormatted}</td>
              {time && <td className="border-0 text-muted-foreground">{localTimeFormatted}</td>}
            </tr>
          </tbody>
        </table>
      </TooltipContent>
    </Tooltip>
  )
}
