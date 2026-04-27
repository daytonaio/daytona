/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { Region, Runner, RunnerState } from '@daytona/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  Table as ReactTable,
  RowData,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { MoreHorizontal, Server } from 'lucide-react'
import { useState } from 'react'
import { CopyButton } from './CopyButton'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { RefreshIntervalValue, RefreshSegmentedButton } from './RefreshSegmentedButton'
import { SearchInput } from './SearchInput'
import { Badge, BadgeProps } from './ui/badge'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Skeleton } from './ui/skeleton'
import { Switch } from './ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from './ui/table'

type RunnerTableMeta = {
  deletePermitted: boolean
  getRegionName: (regionId: string) => string | undefined
  isLoadingRunner: (runner: Runner) => boolean
  onDelete: (runner: Runner) => void
  onToggleEnabled: (runner: Runner) => void
  writePermitted: boolean
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    runner?: TData extends Runner ? RunnerTableMeta : never
  }
}

const getMeta = (table: ReactTable<Runner>) => {
  return table.options.meta?.runner as RunnerTableMeta
}

interface RunnerTableProps {
  data: Runner[]
  regions: Region[]
  loading: boolean
  isLoadingRunner: (runner: Runner) => boolean
  writePermitted: boolean
  deletePermitted: boolean
  onToggleEnabled: (runner: Runner) => void
  onDelete: (runner: Runner) => void
  getRegionName: (regionId: string) => string | undefined
  onRowClick?: (runner: Runner) => void
  refreshInterval?: RefreshIntervalValue
  onRefreshIntervalChange?: (value: RefreshIntervalValue) => void
  onRefresh?: () => void
  isRefreshing?: boolean
  lastUpdatedAt?: number
}

export function RunnerTable({
  data,
  regions,
  loading,
  isLoadingRunner,
  writePermitted,
  deletePermitted,
  onToggleEnabled,
  onDelete,
  getRegionName,
  onRowClick,
  refreshInterval = false,
  onRefreshIntervalChange,
  onRefresh,
  isRefreshing = false,
  lastUpdatedAt,
}: RunnerTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  const table = useReactTable({
    data,
    columns: runnerColumns,
    meta: {
      runner: {
        deletePermitted,
        getRegionName,
        isLoadingRunner,
        onDelete,
        onToggleEnabled,
        writePermitted,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, columnId, filterValue) => {
      const runner = row.original as Runner
      const searchValue = filterValue.toLowerCase()
      const regionName = getRegionName(runner.region) ?? runner.region
      return (
        runner.id.toLowerCase().includes(searchValue) ||
        runner.name.toLowerCase().includes(searchValue) ||
        regionName.toLowerCase().includes(searchValue)
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
      columnPinning: {
        right: ['actions'],
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasSearch = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-2">
        <SearchInput
          debounced
          value={globalFilter ?? ''}
          onValueChange={handleChangeFilter}
          placeholder="Search by ID, Name, or Region"
          containerClassName="max-w-sm"
        />
        {onRefreshIntervalChange && onRefresh && (
          <RefreshSegmentedButton
            className="ml-auto"
            value={refreshInterval}
            onChange={onRefreshIntervalChange}
            onRefresh={onRefresh}
            isRefreshing={isRefreshing}
            lastUpdatedAt={lastUpdatedAt}
            options={[
              { label: 'Off', value: false },
              { label: 'Every 5s', value: 5000 },
              { label: 'Every 10s', value: 10000 },
              { label: 'Every 30s', value: 30000 },
              { label: 'Every 1m', value: 60000 },
            ]}
          />
        )}
      </div>

      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={runnerColumns.length}
              message={hasSearch ? 'No matching runners found.' : 'No runners found.'}
              icon={<Server />}
              description={
                hasSearch ? null : (
                  <div className="space-y-2">
                    <p>Runners are the machines that run your sandboxes.</p>
                    {regions.length === 0 && (
                      <p>There must be at least one region in your organization before runners can be created.</p>
                    )}
                  </div>
                )
              }
              action={
                hasSearch ? (
                  <Button variant="outline" onClick={() => handleChangeFilter('')}>
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
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell
                        className="px-2"
                        key={`${rowIndex}-${column.id}`}
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
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={cn(
                    'group/table-row',
                    isLoadingRunner(row.original) ? 'opacity-50 pointer-events-none' : '',
                    onRowClick && 'cursor-pointer hover:bg-muted/50',
                  )}
                  onClick={() => onRowClick?.(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      className={cn('px-2', {
                        'group-hover/table-row:underline': onRowClick && cell.column.id === 'name',
                      })}
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
        <Pagination table={table} entityName="Runners" />
      </PageFooterPortal>
    </div>
  )
}

const getStateBadgeVariant = (state: RunnerState): BadgeProps['variant'] => {
  switch (state) {
    case RunnerState.READY:
      return 'success'
    case RunnerState.UNRESPONSIVE:
      return 'destructive'
    case RunnerState.INITIALIZING:
      return 'warning'
    case RunnerState.DISABLED:
    case RunnerState.DECOMMISSIONED:
    default:
      return 'secondary'
  }
}

const getStateLabel = (state: RunnerState) => {
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

const isRunnerSchedulable = (runner: Runner) => {
  return !runner.unschedulable
}

const runnerColumns: ColumnDef<Runner>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    size: 240,
    cell: ({ row }) => {
      return (
        <div className="w-full truncate flex items-center gap-1 group/copy-button">
          <span className="truncate block text-sm">{row.original.name}</span>
          <CopyButton value={row.original.name} size="icon-xs" autoHide tooltipText="Copy Name" />
        </div>
      )
    },
  },
  {
    accessorKey: 'id',
    header: 'ID',
    size: 240,
    cell: ({ row }) => {
      return (
        <div className="w-full truncate flex items-center gap-1 group/copy-button">
          <span className="truncate block text-sm">{row.original.id}</span>
          <CopyButton value={row.original.id} size="icon-xs" autoHide tooltipText="Copy ID" />
        </div>
      )
    },
  },
  {
    accessorKey: 'regionId',
    header: 'Region',
    size: 180,
    cell: ({ row, table }) => {
      const { getRegionName } = getMeta(table)
      const regionName = getRegionName(row.original.region) ?? row.original.region

      return (
        <div className="w-full truncate flex items-center gap-1 group/copy-button">
          <span className="truncate block text-sm">{regionName}</span>
          <CopyButton value={regionName} size="icon-xs" autoHide tooltipText="Copy Region" />
        </div>
      )
    },
  },
  {
    accessorKey: 'state',
    header: 'State',
    size: 200,
    cell: ({ row }) => (
      <Badge variant={getStateBadgeVariant(row.original.state)}>{getStateLabel(row.original.state)}</Badge>
    ),
  },
  {
    accessorKey: 'unschedulable',
    header: 'Schedulable',
    size: 60,
    cell: ({ row, table }) => {
      const { isLoadingRunner, onToggleEnabled, writePermitted } = getMeta(table)
      const isLoading = isLoadingRunner(row.original)

      return (
        <Switch
          checked={isRunnerSchedulable(row.original)}
          onCheckedChange={() => writePermitted && !isLoading && onToggleEnabled(row.original)}
          disabled={!writePermitted || isLoading}
          onClick={(e) => e.stopPropagation()}
        />
      )
    },
  },
  {
    id: 'actions',
    size: 48,
    minSize: 48,
    maxSize: 48,
    header: () => {
      return null
    },
    cell: ({ row, table }) => {
      const { deletePermitted, isLoadingRunner, onDelete } = getMeta(table)

      if (!deletePermitted) {
        return null
      }

      const isLoading = isLoadingRunner(row.original)

      return (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu" disabled={isLoading}>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete(row.original)
                }}
                variant="destructive"
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
