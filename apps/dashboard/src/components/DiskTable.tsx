/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Loader2, AlertTriangle, MoreHorizontal, CheckCircle, Timer, HardDrive, Disc } from 'lucide-react'
import { useMemo, useState } from 'react'
import { OrganizationRolePermissionsEnum, DiskDto, DiskState } from '@daytonaio/api-client'
import {
  ColumnDef,
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
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableFacetedFilter, FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getRelativeTimeString } from '@/lib/utils'
import { TableEmptyState } from './TableEmptyState'

interface DiskTableProps {
  data: DiskDto[]
  loading: boolean
  processingDiskAction: Record<string, boolean>
  onDelete: (disk: DiskDto) => void
  onBulkDelete: (disks: DiskDto[]) => void
}

export function DiskTable({ data, loading, processingDiskAction, onDelete, onBulkDelete }: DiskTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const columns = getColumns({
    onDelete,
    processingDiskAction,
    deletePermitted,
  })
  const table = useReactTable({
    data,
    columns,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
    },
    enableRowSelection: true,
    getRowId: (row) => row.id,
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('name')?.setFilterValue(value)}
          placeholder="Search..."
          className="max-w-sm mr-4"
        />
        {table.getColumn('state') && (
          <DataTableFacetedFilter column={table.getColumn('state')} title="State" options={statuses} />
        )}
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id} colSpan={header.colSpan}>
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState colSpan={columns.length} message={loading ? 'Loading disks...' : 'No disks found'} />
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 py-4">
        <div className="flex items-center space-x-2">
          <p className="text-sm font-medium">Rows per page</p>
          <select
            value={table.getState().pagination.pageSize}
            onChange={(e) => {
              table.setPageSize(Number(e.target.value))
            }}
            className="h-8 w-[70px] rounded border border-input bg-background px-3 py-1 text-sm ring-offset-background focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
          >
            {[10, 20, 30, 40, 50].map((pageSize) => (
              <option key={pageSize} value={pageSize}>
                {pageSize}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center space-x-6 lg:space-x-8">
          <div className="flex items-center space-x-2">
            <p className="text-sm font-medium">
              Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              className="hidden h-8 w-8 p-0 lg:flex"
              onClick={() => table.setPageIndex(0)}
              disabled={!table.getCanPreviousPage()}
            >
              <span className="sr-only">Go to first page</span>
              {'<<'}
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              <span className="sr-only">Go to previous page</span>
              {'<'}
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
            >
              <span className="sr-only">Go to next page</span>
              {'>'}
            </Button>
            <Button
              variant="outline"
              className="hidden h-8 w-8 p-0 lg:flex"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              disabled={!table.getCanNextPage()}
            >
              <span className="sr-only">Go to last page</span>
              {'>>'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

const statuses: FacetedFilterOption[] = [
  {
    value: DiskState.FRESH,
    label: 'Fresh',
    icon: ({ className }: { className?: string }) => <Disc className={className} />,
  },
  {
    value: DiskState.PULLING,
    label: 'Pulling',
    icon: ({ className }: { className?: string }) => <Loader2 className={className} />,
  },
  {
    value: DiskState.READY,
    label: 'Ready',
    icon: ({ className }: { className?: string }) => <CheckCircle className={className} />,
  },
  {
    value: DiskState.ATTACHED,
    label: 'Attached',
    icon: ({ className }: { className?: string }) => <HardDrive className={className} />,
  },
  {
    value: DiskState.PUSHING,
    label: 'Pushing',
    icon: ({ className }: { className?: string }) => <Timer className={className} />,
  },
  {
    value: DiskState.STORED,
    label: 'Stored',
    icon: ({ className }: { className?: string }) => <CheckCircle className={className} />,
  },
]

function getColumns({
  onDelete,
  processingDiskAction,
  deletePermitted,
}: {
  onDelete: (disk: DiskDto) => void
  processingDiskAction: Record<string, boolean>
  deletePermitted: boolean
}): ColumnDef<DiskDto>[] {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'id',
      header: 'ID',
      cell: ({ row }) => {
        const id = row.getValue('id') as string
        return (
          <div className="font-mono text-sm">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="cursor-pointer">{id.substring(0, 8)}...</span>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{id}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        )
      },
    },
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => {
        const name = row.getValue('name') as string
        return <div className="font-medium">{name}</div>
      },
    },
    {
      accessorKey: 'size',
      header: 'Size',
      cell: ({ row }) => {
        const size = row.getValue('size') as number
        return <div className="text-sm">{size} GB</div>
      },
    },
    {
      accessorKey: 'state',
      header: 'State',
      cell: ({ row }) => {
        const state = row.getValue('state') as DiskState
        return <StateBadge state={state} />
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id))
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created At',
      cell: ({ row }) => {
        const createdAt = row.getValue('createdAt') as string
        const { relativeTimeString } = getRelativeTimeString(new Date(createdAt))
        return <div className="text-sm">{relativeTimeString}</div>
      },
    },
    {
      accessorKey: 'updatedAt',
      header: 'Updated At',
      cell: ({ row }) => {
        const updatedAt = row.getValue('updatedAt') as string
        const { relativeTimeString } = getRelativeTimeString(new Date(updatedAt))
        return <div className="text-sm">{relativeTimeString}</div>
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const disk = row.original
        const isProcessing = processingDiskAction[disk.id]

        return (
          <div className="flex items-center">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0" disabled={isProcessing}>
                  <span className="sr-only">Open menu</span>
                  {isProcessing ? <Loader2 className="h-4 w-4 animate-spin" /> : <MoreHorizontal className="h-4 w-4" />}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {deletePermitted && (
                  <DropdownMenuItem onClick={() => onDelete(disk)} className="text-red-600" disabled={isProcessing}>
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
}

function StateBadge({ state }: { state: DiskState }) {
  const getStateConfig = (state: DiskState) => {
    switch (state) {
      case DiskState.FRESH:
        return {
          icon: <Disc className="mr-1 h-3 w-3" />,
          label: 'Fresh',
          className: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
        }
      case DiskState.PULLING:
        return {
          icon: <Loader2 className="mr-1 h-3 w-3 animate-spin" />,
          label: 'Pulling',
          className: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300',
        }
      case DiskState.READY:
        return {
          icon: <CheckCircle className="mr-1 h-3 w-3" />,
          label: 'Ready',
          className: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
        }
      case DiskState.ATTACHED:
        return {
          icon: <HardDrive className="mr-1 h-3 w-3" />,
          label: 'Attached',
          className: 'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300',
        }
      case DiskState.PUSHING:
        return {
          icon: <Timer className="mr-1 h-3 w-3" />,
          label: 'Pushing',
          className: 'bg-orange-100 text-orange-800 dark:bg-orange-900/50 dark:text-orange-300',
        }
      case DiskState.STORED:
        return {
          icon: <CheckCircle className="mr-1 h-3 w-3" />,
          label: 'Stored',
          className: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
        }
      default:
        return {
          icon: <AlertTriangle className="mr-1 h-3 w-3" />,
          label: 'Unknown',
          className: 'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300',
        }
    }
  }

  const config = getStateConfig(state)

  return (
    <div className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${config.className}`}>
      {config.icon}
      {config.label}
    </div>
  )
}
