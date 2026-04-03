/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { TableEmptyState } from '@/components/TableEmptyState'
import { PageFooterPortal } from '@/components/PageLayout'
import { DataTableFacetedFilter } from '@/components/ui/data-table-faceted-filter'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '@/components/ui/table'
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
import { cn } from '@/lib/utils'
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
      columnPinning: {
        right: ['actions'],
      },
    },
    meta: {
      endpointEvents: {
        onReplay,
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0

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
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center">
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
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message="No events found."
              icon={<Mail className="size-8" />}
              description={
                <div className="space-y-2">
                  <p>Events will appear here when webhooks are triggered.</p>
                </div>
              }
            />
          ) : undefined
        }
        style={isEmpty ? undefined : { tableLayout: 'fixed', width: '100%' }}
      >
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      className={cn('px-2', header.column.id === 'actions' && 'sticky right-0 z-[2]')}
                      key={header.id}
                      style={isEmpty ? undefined : { width: `${header.column.getSize()}px` }}
                      sticky={header.column.id === 'actions' ? 'right' : undefined}
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
                {Array.from({ length: 25 }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell
                        key={column.id}
                        className={cn('px-2', column.id === 'actions' && 'sticky right-0 z-[1]')}
                        sticky={column.id === 'actions' ? 'right' : undefined}
                      >
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
                      className={cn('px-2', cell.column.id === 'actions' && 'sticky right-0 z-[1]')}
                      key={cell.id}
                      style={{
                        width: `${cell.column.getSize()}px`,
                      }}
                      sticky={cell.column.id === 'actions' ? 'right' : undefined}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : null}
          </TableBody>
        </Table>
      </TableContainer>
      <PageFooterPortal>
        <Pagination table={table} entityName="Events" />
      </PageFooterPortal>
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
