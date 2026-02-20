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
import { EndpointMessageOut } from 'svix'
import { columns, eventTypeOptions, statusOptions } from './columns'
import { EventDetailsSheet } from './EventDetailsSheet'

interface EndpointEventsTableProps {
  data: EndpointMessageOut[]
  loading: boolean
  onReplay: (msgId: string) => void
}

export function EndpointEventsTable({ data, loading, onReplay }: EndpointEventsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [selectedEventIndex, setSelectedEventIndex] = useState<number | null>(null)
  const [sheetOpen, setSheetOpen] = useState(false)

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
      const event = row.original
      const searchValue = filterValue.toLowerCase()
      return (
        (event.id?.toLowerCase().includes(searchValue) ?? false) ||
        (event.eventType?.toLowerCase().includes(searchValue) ?? false) ||
        (event.statusText?.toLowerCase().includes(searchValue) ?? false)
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
    meta: {
      endpointEvents: {
        onReplay,
      },
    },
  })

  const handleRowClick = useCallback((index: number) => {
    setSelectedEventIndex(index)
    setSheetOpen(true)
  }, [])

  const rowCount = table.getRowModel().rows.length

  const handleNavigate = useCallback(
    (direction: 'prev' | 'next') => {
      setSelectedEventIndex((prev) => {
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
          placeholder="Search by Event Type, Message ID, or Status"
          className="max-w-sm mr-4"
        />
        {table.getColumn('eventType') && (
          <DataTableFacetedFilter
            column={table.getColumn('eventType')}
            title="Event Type"
            options={eventTypeOptions}
            className="mr-2"
          />
        )}
        {table.getColumn('status') && (
          <DataTableFacetedFilter column={table.getColumn('status')} title="Status" options={statusOptions} />
        )}
      </div>
      <div className="rounded-md">
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
              table.getRowModel().rows.map((row, rowIndex) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`cursor-pointer hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none ${sheetOpen && selectedEventIndex === rowIndex ? 'bg-muted/50' : ''}`}
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
                icon={<Mail className="size-8" />}
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
      <EventDetailsSheet
        event={selectedEventIndex !== null ? (table.getRowModel().rows[selectedEventIndex]?.original ?? null) : null}
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        onNavigate={handleNavigate}
        hasPrev={selectedEventIndex !== null && selectedEventIndex > 0}
        hasNext={selectedEventIndex !== null && selectedEventIndex < table.getRowModel().rows.length - 1}
        onReplay={onReplay}
      />
    </div>
  )
}
