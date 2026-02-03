/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { TableEmptyState } from '@/components/TableEmptyState'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { getRelativeTimeString } from '@/lib/utils'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Mail } from 'lucide-react'
import { useState } from 'react'
import { EndpointMessageOut } from 'svix'
import { CopyButton } from '../CopyButton'

interface EndpointEventsTableProps {
  data: EndpointMessageOut[]
  loading: boolean
}

export function EndpointEventsTable({ data, loading }: EndpointEventsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns = getColumns()

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, columnId, filterValue) => {
      const attempt = row.original
      const searchValue = filterValue.toLowerCase()
      return (
        (attempt.id?.toLowerCase().includes(searchValue) ?? false) ||
        (attempt.statusText?.toLowerCase().includes(searchValue) ?? false)
      )
    },
    state: {
      sorting,
      globalFilter,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={globalFilter ?? ''}
          onChange={(value) => setGlobalFilter(String(value))}
          placeholder="Search by Attempt ID or Status"
          className="max-w-sm"
        />
      </div>
      <div className="rounded-md border">
        <Table style={{ tableLayout: 'fixed', width: '100%' }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      className="px-2"
                      key={header.id}
                      style={{
                        width: `${header.column.getSize()}px`,
                      }}
                    >
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <>
                {Array.from(new Array(5)).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell key={column.id} className="px-2">
                        <Skeleton className="h-4 w-10/12" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      className="px-2"
                      key={cell.id}
                      style={{
                        width: `${cell.column.getSize()}px`,
                      }}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No events found."
                icon={<Mail className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>Events will appear here when webhooks are triggered.</p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Events" />
    </div>
  )
}

const getColumns = (): ColumnDef<EndpointMessageOut>[] => {
  const columns: ColumnDef<EndpointMessageOut>[] = [
    {
      accessorKey: 'id',
      header: 'Attempt ID',
      size: 300,
      cell: ({ row }) => {
        const attemptId = row.original.id
        return (
          <div
            className="w-full truncate flex items-center gap-2 group/copy-button"
            onClick={(e) => {
              e.stopPropagation()
            }}
          >
            <span className="truncate block font-mono text-sm">{attemptId ?? '-'}</span>
            {attemptId && <CopyButton value={attemptId} size="icon-xs" autoHide />}
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      size: 150,
      cell: ({ row }) => {
        const status = row.original.status
        const statusText = row.original.statusText || 'unknown'
        const variant = status === 0 ? 'success' : status === 1 ? 'secondary' : 'destructive'
        return (
          <Badge variant={variant} className="capitalize">
            {statusText}
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
        const relativeTime = getRelativeTimeString(timestamp).relativeTimeString

        return (
          <TimestampTooltip timestamp={timestamp}>
            <span className="cursor-default">{relativeTime}</span>
          </TimestampTooltip>
        )
      },
    },
    {
      accessorKey: 'nextAttempt',
      header: 'Next Attempt',
      size: 200,
      cell: ({ row }) => {
        const nextAttempt = row.original.nextAttempt
        if (!nextAttempt) {
          return <span className="text-muted-foreground">-</span>
        }
        const relativeTime = getRelativeTimeString(nextAttempt).relativeTimeString

        return (
          <TimestampTooltip timestamp={nextAttempt}>
            <span className="cursor-default">{relativeTime}</span>
          </TimestampTooltip>
        )
      },
    },
  ]

  return columns
}
