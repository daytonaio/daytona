/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { Region, RegionType } from '@daytona/api-client'
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
import { MapPinned, MoreHorizontal } from 'lucide-react'
import { useState } from 'react'
import { CopyButton } from './CopyButton'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { SearchInput } from './SearchInput'
import { TimestampTooltip } from './TimestampTooltip'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Skeleton } from './ui/skeleton'
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

type RegionTableMeta = {
  deletePermitted: boolean
  isLoadingRegion: (region: Region) => boolean
  onDelete: (region: Region) => void
  onOpenDetails: (region: Region) => void
  writePermitted: boolean
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    region?: TData extends Region ? RegionTableMeta : never
  }
}

const getMeta = (table: ReactTable<Region>) => {
  return table.options.meta?.region as RegionTableMeta
}

interface DataTableProps {
  data: Region[]
  loading: boolean
  isLoadingRegion: (region: Region) => boolean
  deletePermitted: boolean
  writePermitted: boolean
  onDelete: (region: Region) => void
  onOpenDetails: (region: Region) => void
}

export function RegionTable({
  data,
  loading,
  isLoadingRegion,
  deletePermitted,
  writePermitted,
  onDelete,
  onOpenDetails,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  const table = useReactTable({
    data,
    columns: regionColumns,
    meta: {
      region: {
        deletePermitted,
        isLoadingRegion,
        onDelete,
        onOpenDetails,
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
      const region = row.original as Region
      const searchValue = filterValue.toLowerCase()
      return region.name.toLowerCase().includes(searchValue) || region.id.toLowerCase().includes(searchValue)
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
          placeholder="Search by Name or ID"
          containerClassName="max-w-sm"
        />
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={regionColumns.length}
              message={hasSearch ? 'No matching regions found.' : 'No custom regions found.'}
              icon={<MapPinned />}
              description={hasSearch ? null : <p>Create regions for grouping runners and sandboxes.</p>}
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
                        key={`${rowIndex}-${column.id}`}
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
              table.getRowModel().rows.map((row) => {
                const isCustom = row.original.regionType === RegionType.CUSTOM
                const isLoading = isLoadingRegion(row.original)
                return (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className={cn('group/table-row', {
                      'opacity-50 pointer-events-none': isLoading,
                      'cursor-pointer hover:bg-muted/50': isCustom && !isLoading,
                    })}
                    onClick={() => {
                      if (isCustom && !isLoading) {
                        onOpenDetails(row.original)
                      }
                    }}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        className={cn('px-2', {
                          'group-hover/table-row:underline': isCustom && !isLoading && cell.column.id === 'name',
                        })}
                        key={cell.id}
                        style={getColumnSizeStyles(cell.column)}
                        sticky={cell.column.getIsPinned()}
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                )
              })
            ) : null}
          </TableBody>
        </Table>
      </TableContainer>
      <PageFooterPortal>
        <Pagination table={table} entityName="Regions" />
      </PageFooterPortal>
    </div>
  )
}

const regionColumns: ColumnDef<Region>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    size: 300,
    cell: ({ row }) => {
      return (
        <div className="w-full truncate flex items-center gap-1 group/copy-button">
          <span className="truncate block">{row.original.name}</span>
          <CopyButton value={row.original.name} size="icon-xs" autoHide tooltipText="Copy Name" />
        </div>
      )
    },
  },
  {
    accessorKey: 'id',
    header: 'ID',
    size: 300,
    cell: ({ row }) => {
      return (
        <div className="w-full truncate flex items-center gap-1 group/copy-button">
          <span className="truncate block">{row.original.id}</span>
          <CopyButton value={row.original.id} size="icon-xs" autoHide tooltipText="Copy ID" />
        </div>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: ({ row }) => {
      if (row.original.regionType !== RegionType.CUSTOM) {
        return null
      }

      const createdAt = row.original.createdAt
      const relativeTime = getRelativeTimeString(createdAt).relativeTimeString

      return (
        <TimestampTooltip timestamp={createdAt?.toString()}>
          <span className="cursor-default">{relativeTime}</span>
        </TimestampTooltip>
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
      const { deletePermitted, isLoadingRegion, onDelete, onOpenDetails, writePermitted } = getMeta(table)

      if (row.original.regionType !== RegionType.CUSTOM || (!deletePermitted && !writePermitted)) {
        return <div className="flex justify-end h-8 w-8" />
      }

      const isLoading = isLoadingRegion(row.original)

      return (
        <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu" disabled={isLoading}>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onOpenDetails(row.original)} disabled={isLoading}>
                Details
              </DropdownMenuItem>
              {deletePermitted && (
                <DropdownMenuItem onClick={() => onDelete(row.original)} variant="destructive" disabled={isLoading}>
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]
