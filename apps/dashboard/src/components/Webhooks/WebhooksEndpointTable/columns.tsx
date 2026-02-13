/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { getRelativeTimeString } from '@/lib/utils'
import { ColumnDef, RowData, Table } from '@tanstack/react-table'
import { MoreHorizontal } from 'lucide-react'
import { EndpointOut } from 'svix'
import { CopyButton } from '../../CopyButton'

type WebhooksEndpointTableMeta = {
  onDisable: (endpoint: EndpointOut) => void
  onDelete: (endpoint: EndpointOut) => void
  isLoadingEndpoint: (endpoint: EndpointOut) => boolean
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    webhookEndpoints?: TData extends EndpointOut ? WebhooksEndpointTableMeta : never
  }
}

const getMeta = (table: Table<EndpointOut>) => {
  return table.options.meta?.webhookEndpoints as WebhooksEndpointTableMeta
}

const columns: ColumnDef<EndpointOut>[] = [
  {
    accessorKey: 'description',
    header: 'Name',
    size: 200,
    cell: ({ row }) => (
      <div className="w-full truncate flex items-center gap-2">
        <span className="truncate block hover:underline focus:underline cursor-pointer">
          {row.original.description || 'Unnamed Endpoint'}
        </span>
      </div>
    ),
  },
  {
    accessorKey: 'url',
    header: 'URL',
    size: 300,
    cell: ({ row }) => (
      <div className="w-full truncate flex items-center gap-2 group/copy-button">
        <span className="truncate block">{row.original.url}</span>
        <CopyButton value={row.original.url} size="icon-xs" autoHide tooltipText="Copy URL" />
      </div>
    ),
  },
  {
    accessorKey: 'disabled',
    header: 'Status',
    size: 100,
    cell: ({ row }) => (
      <Badge variant={row.original.disabled ? 'secondary' : 'success'}>
        {row.original.disabled ? 'Disabled' : 'Active'}
      </Badge>
    ),
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    size: 150,
    cell: ({ row }) => {
      const createdAt = row.original.createdAt
      const relativeTime = getRelativeTimeString(createdAt).relativeTimeString

      return (
        <TimestampTooltip timestamp={typeof createdAt === 'string' ? createdAt : createdAt.toISOString()}>
          <span className="cursor-default">{relativeTime}</span>
        </TimestampTooltip>
      )
    },
  },
  {
    id: 'actions',
    header: () => null,
    size: 50,
    cell: ({ row, table }) => {
      const { onDisable, onDelete, isLoadingEndpoint } = getMeta(table)
      const isLoading = isLoadingEndpoint(row.original)

      return (
        <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-xs" disabled={isLoading}>
                <MoreHorizontal className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onDisable(row.original)} className="cursor-pointer" disabled={isLoading}>
                {row.original.disabled ? 'Enable' : 'Disable'}
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => onDelete(row.original)}
                className="cursor-pointer"
                disabled={isLoading}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]

export { columns }
export type { WebhooksEndpointTableMeta }
