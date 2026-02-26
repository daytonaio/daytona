/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import { WEBHOOK_EVENTS } from '@/constants/webhook-events'
import { getRelativeTimeString } from '@/lib/utils'
import { ColumnDef } from '@tanstack/react-table'
import { MessageOut } from 'svix'
import { CopyButton } from '../../CopyButton'

const columns: ColumnDef<MessageOut>[] = [
  {
    accessorKey: 'id',
    header: 'Message ID',
    size: 300,
    cell: ({ row }) => {
      const msgId = row.original.id
      return (
        <div className="w-full truncate flex items-center gap-2 group/copy-button">
          <span className="truncate block font-mono text-sm hover:underline focus:underline cursor-pointer">
            {msgId ?? '-'}
          </span>
          {msgId && (
            <span onClick={(e) => e.stopPropagation()}>
              <CopyButton value={msgId} size="icon-xs" autoHide tooltipText="Copy Message ID" />
            </span>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'eventType',
    header: 'Event Type',
    size: 200,
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
    cell: ({ row }) => {
      const eventType = row.original.eventType
      return (
        <Badge variant="secondary" className="font-normal text-xs">
          {eventType}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'timestamp',
    header: 'Timestamp',
    size: 200,
    cell: ({ row }) => {
      const timestamp = row.original.timestamp
      if (!timestamp) {
        return <span className="text-muted-foreground">-</span>
      }
      const relativeTime = getRelativeTimeString(timestamp)

      return (
        <TimestampTooltip timestamp={timestamp.toString()}>
          <span className="cursor-default">{relativeTime.relativeTimeString}</span>
        </TimestampTooltip>
      )
    },
  },
]

const eventTypeOptions: FacetedFilterOption[] = WEBHOOK_EVENTS.map((event) => ({
  label: event.label,
  value: event.value,
}))

export { columns, eventTypeOptions }
