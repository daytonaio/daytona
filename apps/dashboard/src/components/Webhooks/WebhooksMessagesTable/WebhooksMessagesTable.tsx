/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { TableEmptyState } from '@/components/TableEmptyState'
import { DataTableFacetedFilter } from '@/components/ui/data-table-faceted-filter'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import {
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Mail } from 'lucide-react'
import { useCallback, useState } from 'react'
import { useMessages } from 'svix-react'
import { columns, eventTypeOptions } from './columns'
import { MessageDetailsSheet } from './MessageDetailsSheet'

export function WebhooksMessagesTable() {
  const messages = useMessages()
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [selectedMessageIndex, setSelectedMessageIndex] = useState<number | null>(null)
  const [sheetOpen, setSheetOpen] = useState(false)

  const data = messages.data ?? []

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onColumnFiltersChange: setColumnFilters,
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const message = row.original
      const searchValue = filterValue.toLowerCase()
      return (
        (message.id?.toLowerCase().includes(searchValue) ?? false) ||
        (message.eventType?.toLowerCase().includes(searchValue) ?? false) ||
        (message.eventId?.toLowerCase().includes(searchValue) ?? false)
      )
    },
    state: {
      sorting,
      columnFilters,
      globalFilter,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  const handleRowClick = useCallback((index: number) => {
    setSelectedMessageIndex(index)
    setSheetOpen(true)
  }, [])

  const rowCount = table.getRowModel().rows.length

  const handleNavigate = useCallback(
    (direction: 'prev' | 'next') => {
      setSelectedMessageIndex((prev) => {
        if (prev === null) return null
        if (direction === 'prev' && prev > 0) return prev - 1
        if (direction === 'next' && prev < rowCount - 1) return prev + 1
        return prev
      })
    },
    [rowCount],
  )

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={globalFilter ?? ''}
          onChange={(value) => setGlobalFilter(String(value))}
          placeholder="Search by Message ID, Event Type, or Event ID"
          className="max-w-sm mr-4"
        />
        {table.getColumn('eventType') && (
          <DataTableFacetedFilter column={table.getColumn('eventType')} title="Event Type" options={eventTypeOptions} />
        )}
      </div>
      <div className="rounded-md border">
        <Table style={{ tableLayout: 'fixed', width: '100%' }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead className="px-2" key={header.id} style={{ width: `${header.column.getSize()}px` }}>
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {messages.loading ? (
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
              table.getRowModel().rows.map((row, rowIndex) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`cursor-pointer hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none ${sheetOpen && selectedMessageIndex === rowIndex ? 'bg-muted/50' : ''}`}
                  tabIndex={0}
                  role="button"
                  onClick={() => handleRowClick(rowIndex)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault()
                      handleRowClick(rowIndex)
                    }
                  }}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell className="px-2" key={cell.id} style={{ width: `${cell.column.getSize()}px` }}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No messages found."
                icon={<Mail className="size-8" />}
                description={
                  <div className="space-y-2">
                    <p>Messages will appear here when webhook events are triggered.</p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Messages" />
      <MessageDetailsSheet
        message={
          selectedMessageIndex !== null ? (table.getRowModel().rows[selectedMessageIndex]?.original ?? null) : null
        }
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        onNavigate={handleNavigate}
        hasPrev={selectedMessageIndex !== null && selectedMessageIndex > 0}
        hasNext={selectedMessageIndex !== null && selectedMessageIndex < table.getRowModel().rows.length - 1}
      />
    </div>
  )
}
