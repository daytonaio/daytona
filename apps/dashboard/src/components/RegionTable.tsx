/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { Region, RegionType } from '@daytonaio/api-client'
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
import { Copy, MapPinned, MoreHorizontal } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { DebouncedInput } from './DebouncedInput'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { TableEmptyState } from './TableEmptyState'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Skeleton } from './ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from './ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

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

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  const columns = getColumns({
    onDelete,
    isLoadingRegion,
    deletePermitted,
    writePermitted,
    copyToClipboard,
    onOpenDetails,
  })
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

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-4">
        <DebouncedInput
          value={globalFilter ?? ''}
          onChange={(value) => setGlobalFilter(String(value))}
          placeholder="Search by Name or ID"
          className="max-w-sm"
        />
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message="No custom regions found."
              icon={<MapPinned className="w-8 h-8" />}
              description={
                <div className="space-y-2">
                  <p>Create regions for grouping runners and sandboxes.</p>
                </div>
              }
            />
          ) : undefined
        }
      >
        <Table style={isEmpty ? undefined : { tableLayout: 'fixed', width: '100%' }}>
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
                {Array.from({ length: 25 }).map((_, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {columns.map((column, columnIndex) => (
                      <TableCell
                        key={`${rowIndex}-${column.id ?? columnIndex}`}
                        className={cn('px-2', column.id === 'actions' && 'sticky right-0 z-[1]')}
                        style={{
                          width: `${column.size ?? 160}px`,
                        }}
                        sticky={column.id === 'actions' ? 'right' : undefined}
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
                    className={`${isLoading ? 'opacity-50 pointer-events-none' : ''} ${isCustom && !isLoading ? 'cursor-pointer hover:bg-muted/50' : ''}`}
                    onClick={() => {
                      if (isCustom && !isLoading) {
                        onOpenDetails(row.original)
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

const getColumns = ({
  onDelete,
  isLoadingRegion,
  deletePermitted,
  writePermitted,
  copyToClipboard,
  onOpenDetails,
}: {
  onDelete: (region: Region) => void
  isLoadingRegion: (region: Region) => boolean
  deletePermitted: boolean
  writePermitted: boolean
  copyToClipboard: (text: string) => Promise<void>
  onOpenDetails: (region: Region) => void
}): ColumnDef<Region>[] => {
  const columns: ColumnDef<Region>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      size: 300,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block">{row.original.name}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              copyToClipboard(row.original.name)
            }}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Copy Name"
          >
            <Copy className="w-3 h-3" />
          </button>
        </div>
      ),
    },
    {
      accessorKey: 'id',
      header: 'ID',
      size: 300,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block">{row.original.id}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              copyToClipboard(row.original.id)
            }}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Copy ID"
          >
            <Copy className="w-3 h-3" />
          </button>
        </div>
      ),
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
        const fullDate = new Date(createdAt).toLocaleString()

        return (
          <Tooltip>
            <TooltipTrigger>
              <span className="cursor-default">{relativeTime}</span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{fullDate}</p>
            </TooltipContent>
          </Tooltip>
        )
      },
    },
  ]

  columns.push({
    id: 'actions',
    size: 48,
    minSize: 48,
    maxSize: 48,
    header: () => {
      return null
    },
    cell: ({ row }) => {
      if (row.original.regionType !== RegionType.CUSTOM || (!deletePermitted && !writePermitted)) {
        return <div className="flex justify-end h-8 w-8" />
      }

      const isLoading = isLoadingRegion(row.original)

      return (
        <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={isLoading}>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() => onOpenDetails(row.original)}
                className="cursor-pointer"
                disabled={isLoading}
              >
                Details
              </DropdownMenuItem>
              {deletePermitted && (
                <DropdownMenuItem
                  onClick={() => onDelete(row.original)}
                  className="cursor-pointer text-red-600 dark:text-red-400"
                  disabled={isLoading}
                >
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  })

  return columns
}
