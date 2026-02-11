/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { WEBHOOK_EVENTS } from '@/constants/webhook-events'
import { getRelativeTimeString } from '@/lib/utils'
import { ColumnDef, RowData, Table } from '@tanstack/react-table'
import { CheckCircle, Clock, MoreHorizontal, XCircle } from 'lucide-react'
import { EndpointMessageOut } from 'svix'
import { CopyButton } from '../../CopyButton'

type EndpointEventsTableMeta = {
  onReplay: (msgId: string) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    endpointEvents?: TData extends EndpointMessageOut ? EndpointEventsTableMeta : never
  }
}

const getMeta = (table: Table<EndpointMessageOut>) => {
  return table.options.meta?.endpointEvents as EndpointEventsTableMeta
}

const columns: ColumnDef<EndpointMessageOut>[] = [
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
    id: 'status',
    accessorFn: (row) => row.statusText || 'unknown',
    header: 'Status',
    size: 150,
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
    cell: ({ row }) => {
      const status = row.original.status
      const variant = status === 0 ? 'success' : status === 1 ? 'secondary' : 'destructive'
      return <Badge variant={variant}>{status === 0 ? 'Success' : status === 1 ? 'Pending' : 'Failed'}</Badge>
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
    header: 'Sent',
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
  {
    id: 'actions',
    size: 50,
    enableHiding: false,
    cell: ({ row, table }) => {
      const { onReplay } = getMeta(table)
      const msgId = row.original.id
      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon-xs">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem className="cursor-pointer" onClick={() => onReplay(msgId)}>
              Replay
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )
    },
  },
]

const eventTypeOptions: FacetedFilterOption[] = WEBHOOK_EVENTS.map((event) => ({
  label: event.label,
  value: event.value,
}))

const statusOptions: FacetedFilterOption[] = [
  { label: 'Success', value: 'success', icon: CheckCircle },
  { label: 'Pending', value: 'pending', icon: Clock },
  { label: 'Failed', value: 'fail', icon: XCircle },
]

export { columns, eventTypeOptions, statusOptions }
export type { EndpointEventsTableMeta }
