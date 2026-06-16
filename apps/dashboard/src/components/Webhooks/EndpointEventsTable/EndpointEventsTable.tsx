/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandCheckboxItem,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandList,
} from '@/components/ui/command'
import { DataTableFacetedFilter, type FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import {
  type Column,
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
import { CircleDot, ListFilter, Mail, Tag } from 'lucide-react'
import { type ReactNode, useCallback, useState } from 'react'
import { EndpointMessageOut } from 'svix'
import { columns, eventTypeOptions, statusOptions } from './columns'
import { EventDetailsSheet } from './EventDetailsSheet'

interface EndpointEventsTableProps {
  data: EndpointMessageOut[]
  loading: boolean
  onReplay: (msgId: string) => void
}

interface WebhookFilterSubmenuProps {
  column?: Column<EndpointMessageOut, unknown>
  icon: ReactNode
  options: readonly FacetedFilterOption[]
  title: string
}

function WebhookFilterSubmenu({ column, icon, options, title }: WebhookFilterSubmenuProps) {
  if (!column) {
    return null
  }

  const values = (column.getFilterValue() as string[] | undefined) ?? []

  const handleFilterChange = (nextValues: string[]) => {
    column.setFilterValue(nextValues.length > 0 ? nextValues : undefined)
  }

  return (
    <DropdownMenuSub>
      <DropdownMenuSubTrigger>
        {icon}
        {title}
      </DropdownMenuSubTrigger>
      <DropdownMenuPortal>
        <DropdownMenuSubContent className="p-0 w-72">
          <Command>
            <CommandInput placeholder={title}>
              <CommandInputButton
                className="text-sm text-muted-foreground hover:text-primary px-2"
                onClick={() => column.setFilterValue(undefined)}
              >
                Clear
              </CommandInputButton>
            </CommandInput>
            <CommandList>
              <CommandGroup>
                {options.map((option) => (
                  <CommandCheckboxItem
                    key={option.value}
                    checked={values.includes(option.value)}
                    onSelect={() => {
                      const nextValues = values.includes(option.value)
                        ? values.filter((value) => value !== option.value)
                        : [...values, option.value]

                      handleFilterChange(nextValues)
                    }}
                  >
                    {option.label}
                  </CommandCheckboxItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </DropdownMenuSubContent>
      </DropdownMenuPortal>
    </DropdownMenuSub>
  )
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
  const hasFilters = globalFilter.trim().length > 0 || columnFilters.length > 0
  const eventTypeColumn = table.getColumn('eventType')
  const statusColumn = table.getColumn('status')
  const hasEventTypeFilter = ((eventTypeColumn?.getFilterValue() as string[]) || []).length > 0
  const hasStatusFilter = ((statusColumn?.getFilterValue() as string[]) || []).length > 0
  const hasColumnFilters = hasEventTypeFilter || hasStatusFilter

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

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  const handleClearFilters = () => {
    handleChangeFilter('')
    table.resetColumnFilters()
  }

  const handleClearColumnFilters = () => {
    table.resetColumnFilters()
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <SearchInput
              debounced
              value={globalFilter ?? ''}
              onValueChange={handleChangeFilter}
              placeholder="Search by Event Type, Message ID, or Status"
              containerClassName="min-w-0 flex-1 sm:max-w-sm"
            />
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  className="shrink-0 bg-transparent hover:bg-accent dark:bg-input/30 dark:hover:bg-accent"
                  aria-label="Filter"
                >
                  <ListFilter className="size-4" />
                  <span className="max-[420px]:hidden">Filter</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-48" align="start">
                <WebhookFilterSubmenu
                  column={eventTypeColumn}
                  icon={<Tag className="size-4" />}
                  title="Event Type"
                  options={eventTypeOptions}
                />
                <WebhookFilterSubmenu
                  column={statusColumn}
                  icon={<CircleDot className="size-4" />}
                  title="Status"
                  options={statusOptions}
                />
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        {hasColumnFilters ? (
          <div className="flex items-start gap-2">
            <div className="flex min-w-0 flex-1 flex-wrap items-center gap-1">
              {hasEventTypeFilter && eventTypeColumn && (
                <DataTableFacetedFilter column={eventTypeColumn} title="Event Type" options={eventTypeOptions} />
              )}
              {hasStatusFilter && statusColumn && (
                <DataTableFacetedFilter column={statusColumn} title="Status" options={statusOptions} />
              )}
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 shrink-0 px-3 text-muted-foreground hover:text-foreground"
              onClick={handleClearColumnFilters}
            >
              Clear
            </Button>
          </div>
        ) : null}
      </div>
      <TableContainer
        className={cn({ 'min-h-[26rem]': isEmpty })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasFilters ? 'No matching events found.' : 'No events found.'}
              icon={<Mail />}
              description={hasFilters ? null : <p>Events will appear here when webhooks are triggered.</p>}
              action={
                hasFilters ? (
                  <Button variant="outline" onClick={handleClearFilters}>
                    Clear filters
                  </Button>
                ) : null
              }
            />
          ) : null
        }
      >
        <Table className="table-fixed" style={{ minWidth: table.getTotalSize() }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      className="px-2"
                      key={header.id}
                      style={getColumnSizeStyles(header.column)}
                      sticky={header.column.getIsPinned()}
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
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell
                        key={column.id}
                        className="px-2"
                        style={getColumnSizeStyles(column)}
                        sticky={column.getIsPinned()}
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
                  className={cn(
                    'cursor-pointer hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-hidden',
                    {
                      'bg-muted/50': sheetOpen && selectedEventIndex === rowIndex,
                    },
                  )}
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
                      style={getColumnSizeStyles(cell.column)}
                      sticky={cell.column.getIsPinned()}
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
