/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Loader2, AlertTriangle, MoreHorizontal, CheckCircle, Timer, HardDrive } from 'lucide-react'
import { useMemo, useState } from 'react'
import { OrganizationRolePermissionsEnum, VolumeDto, VolumeState } from '@daytonaio/api-client'
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

interface VolumeTableProps {
  data: VolumeDto[]
  loading: boolean
  processingVolumeAction: Record<string, boolean>
  onDelete: (volume: VolumeDto) => void
  onBulkDelete: (volumes: VolumeDto[]) => void
}

export function VolumeTable({ data, loading, processingVolumeAction, onDelete, onBulkDelete }: VolumeTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const columns = getColumns({
    onDelete,
    processingVolumeAction,
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
                    <TableHead className="px-2" key={header.id}>
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`${processingVolumeAction[row.original.id] || row.original.state === VolumeState.PENDING_DELETE || row.original.state === VolumeState.DELETING ? 'opacity-50 pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell className="px-2" key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No Volumes yet."
                icon={<HardDrive className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>
                      Volumes are shared, persistent directories backed by S3-compatible storage, perfect for reusing
                      datasets, caching dependencies, or passing files across sandboxes.
                    </p>
                    <p>
                      Create one via the SDK or CLI. <br />
                      <a
                        href="https://www.daytona.io/docs/volumes"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        Read the Volumes guide
                      </a>{' '}
                      to learn more.
                    </p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 py-4">
        {table.getRowModel().rows.some((row) => row.getIsSelected()) && (
          <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
            <PopoverTrigger>
              <Button variant="destructive" size="sm" className="h-8">
                Bulk Delete
              </Button>
            </PopoverTrigger>
            <PopoverContent side="top">
              <div className="flex flex-col gap-4">
                <p>Are you sure you want to delete these Volumes?</p>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="destructive"
                    onClick={() => {
                      onBulkDelete(
                        table
                          .getRowModel()
                          .rows.filter((row) => row.getIsSelected())
                          .map((row) => row.original),
                      )
                      setBulkDeleteConfirmationOpen(false)
                    }}
                  >
                    Delete
                  </Button>
                  <Button variant="outline" onClick={() => setBulkDeleteConfirmationOpen(false)}>
                    Cancel
                  </Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <Pagination table={table} selectionEnabled entityName="Volumes" />
      </div>
    </div>
  )
}

const getStateIcon = (state: VolumeState) => {
  switch (state) {
    case VolumeState.READY:
      return <CheckCircle className="w-4 h-4 flex-shrink-0" />
    case VolumeState.ERROR:
      return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
    default:
      return <Timer className="w-4 h-4 flex-shrink-0" />
  }
}

const getStateColor = (state: VolumeState) => {
  switch (state) {
    case VolumeState.READY:
      return 'text-green-500'
    case VolumeState.ERROR:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state: VolumeState) => {
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

const statuses: FacetedFilterOption[] = [
  { label: getStateLabel(VolumeState.CREATING), value: VolumeState.CREATING, icon: Timer },
  { label: getStateLabel(VolumeState.READY), value: VolumeState.READY, icon: CheckCircle },
  { label: getStateLabel(VolumeState.PENDING_CREATE), value: VolumeState.PENDING_CREATE, icon: Timer },
  { label: getStateLabel(VolumeState.PENDING_DELETE), value: VolumeState.PENDING_DELETE, icon: Timer },
  { label: getStateLabel(VolumeState.DELETING), value: VolumeState.DELETING, icon: Timer },
  { label: getStateLabel(VolumeState.DELETED), value: VolumeState.DELETED, icon: Timer },
  { label: getStateLabel(VolumeState.ERROR), value: VolumeState.ERROR, icon: AlertTriangle },
]

const getColumns = ({
  onDelete,
  processingVolumeAction,
  deletePermitted,
}: {
  onDelete: (volume: VolumeDto) => void
  processingVolumeAction: Record<string, boolean>
  deletePermitted: boolean
}): ColumnDef<VolumeDto>[] => {
  const columns: ColumnDef<VolumeDto>[] = [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            for (const row of table.getRowModel().rows) {
              if (processingVolumeAction[row.original.id]) {
                row.toggleSelected(false)
              } else {
                row.toggleSelected(!!value)
              }
            }
          }}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => {
        if (processingVolumeAction[row.original.id]) {
          return <Loader2 className="w-4 h-4 animate-spin" />
        }
        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            className="translate-y-[2px]"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => {
        return <div className="w-40">{row.original.name}</div>
      },
    },
    {
      id: 'state',
      header: 'State',
      cell: ({ row }) => {
        const volume = row.original
        const state = row.original.state
        const color = getStateColor(state)

        if (state === VolumeState.ERROR && !!volume.errorReason) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className={`flex items-center gap-2 ${color}`}>
                    {getStateIcon(state)}
                    {getStateLabel(state)}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{volume.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return (
          <div className={`flex items-center gap-2 w-40 ${color}`}>
            {getStateIcon(state)}
            <span>{getStateLabel(state)}</span>
          </div>
        )
      },
      accessorKey: 'state',
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id))
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => {
        return getRelativeTimeString(row.original.createdAt).relativeTimeString
      },
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => {
        return getRelativeTimeString(row.original.lastUsedAt).relativeTimeString
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        if (!deletePermitted) {
          return null
        }

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                className={`cursor-pointer text-red-600 dark:text-red-400 ${
                  processingVolumeAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''
                }`}
                disabled={processingVolumeAction[row.original.id]}
                onClick={() => onDelete(row.original)}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}
