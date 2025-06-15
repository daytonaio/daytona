/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotDto, SnapshotState, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useMemo, useState } from 'react'
import { AlertTriangle, CheckCircle, MoreHorizontal, Timer, Trash2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Pagination } from './Pagination'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Checkbox } from './ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { getRelativeTimeString } from '@/lib/utils'
import { TableEmptyState } from './TableEmptyState'
import { Loader2 } from 'lucide-react'
import { Badge } from './ui/badge'

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  onDelete: (snapshot: SnapshotDto) => void
  onBulkDelete?: (snapshots: SnapshotDto[]) => void
  onToggleEnabled: (snapshot: SnapshotDto, enabled: boolean) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
}

export function SnapshotTable({
  data,
  loading,
  loadingSnapshots,
  onDelete,
  onToggleEnabled,
  pagination,
  pageCount,
  onBulkDelete,
  onPaginationChange,
}: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])

  const columns = useMemo(
    () =>
      getColumns({
        onDelete,
        onToggleEnabled,
        loadingSnapshots,
        writePermitted,
        deletePermitted,
      }),
    [onDelete, onToggleEnabled, loadingSnapshots, writePermitted, deletePermitted],
  )

  const columnsWithSelection = useMemo(() => {
    const selectionColumn: ColumnDef<SnapshotDto> = {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            table.getRowModel().rows.forEach((row) => {
              if (!row.original.general) {
                row.toggleSelected()
              }
            })
          }}
          aria-label="Select all"
          disabled={!deletePermitted || loading}
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => {
        if (loadingSnapshots[row.original.id]) {
          return <Loader2 className="w-4 h-4 animate-spin" />
        }

        if (row.original.general) {
          return null
        }

        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            disabled={!deletePermitted || loadingSnapshots[row.original.id] || loading}
            className="translate-y-[2px]"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    }

    return deletePermitted ? [selectionColumn, ...columns] : columns
  }, [deletePermitted, columns, loading, loadingSnapshots])

  const table = useReactTable({
    data,
    columns: columnsWithSelection,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    pageCount: pageCount || 1,
    onPaginationChange: pagination
      ? (updater) => {
          const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
          onPaginationChange(newPagination)
        }
      : undefined,
    state: {
      sorting,
      pagination: {
        pageIndex: pagination?.pageIndex || 0,
        pageSize: pagination?.pageSize || 10,
      },
    },
    getRowId: (row) => row.id,
    enableRowSelection: deletePermitted,
  })

  const selectedRows = table.getSelectedRowModel().rows
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)
  const selectedImages = selectedRows.map((row) => row.original)

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedImages.length > 0) {
      onBulkDelete(selectedImages)
    }
  }

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
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
                <TableCell colSpan={columnsWithSelection.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() ? 'selected' : undefined}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState colSpan={columns.length} message="No Images found." />
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between gap-2 mt-4">
        <div className="flex items-center gap-4">
          {deletePermitted && selectedRows.length > 0 && (
            <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
              <PopoverTrigger>
                <Button variant="destructive" size="sm">
                  <Trash2 className="h-4 w-4 mr-2" />
                  Bulk Delete
                </Button>
              </PopoverTrigger>
              <PopoverContent side="top">
                <div className="flex flex-col gap-4">
                  <p>Are you sure you want to delete these images?</p>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="destructive"
                      onClick={() => {
                        handleBulkDelete()
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
          {deletePermitted && (
            <span className="text-sm text-muted-foreground whitespace-nowrap">
              {selectedRows.length} of {data.length} row(s) selected
            </span>
          )}
        </div>
        <Pagination table={table} entityName="Images" />
      </div>
    </div>
  )
}

const getColumns = ({
  onDelete,
  onToggleEnabled,
  loadingSnapshots,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (snapshot: SnapshotDto) => void
  onToggleEnabled: (snapshot: SnapshotDto, enabled: boolean) => void
  loadingSnapshots: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<SnapshotDto>[] => {
  const columns: ColumnDef<SnapshotDto>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => {
        const snapshot = row.original
        return (
          <div className="flex items-center gap-2">
            {snapshot.name}
            {snapshot.general && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-green-100 text-blue-800 dark:bg-green-900 dark:text-green-300">
                System
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'imageName',
      header: 'Image',
      cell: ({ row }) => {
        const snapshot = row.original
        if (!snapshot.imageName && snapshot.buildInfo) {
          return (
            <Badge variant="secondary" className="rounded-sm px-1 font-medium">
              DECLARATIVE BUILD
            </Badge>
          )
        }
        return snapshot.imageName
      },
    },
    {
      id: 'resources',
      header: 'Resources',
      cell: ({ row }) => {
        const snapshot = row.original
        return `${snapshot.cpu}vCPU / ${snapshot.mem}GiB / ${snapshot.disk}GiB`
      },
    },
    {
      accessorKey: 'state',
      header: 'State',
      cell: ({ row }) => {
        const snapshot = row.original
        const color = getStateColor(snapshot.state)

        if (
          (snapshot.state === SnapshotState.ERROR || snapshot.state === SnapshotState.BUILD_FAILED) &&
          !!snapshot.errorReason
        ) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className={`flex items-center gap-2 ${color}`}>
                    {getStateIcon(snapshot.state)}
                    {getStateLabel(snapshot.state)}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{snapshot.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return (
          <div className={`flex items-center gap-2 ${color}`}>
            {getStateIcon(snapshot.state)}
            {getStateLabel(snapshot.state)}
          </div>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.createdAt).relativeTimeString
      },
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.lastUsedAt).relativeTimeString
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if ((!writePermitted && !deletePermitted) || row.original.general) {
          return null
        }

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {writePermitted && (
                <DropdownMenuItem
                  onClick={() => onToggleEnabled(row.original, !row.original.enabled)}
                  className="cursor-pointer"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  {row.original.enabled ? 'Disable' : 'Enable'}
                </DropdownMenuItem>
              )}
              {deletePermitted && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => onDelete(row.original)}
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    disabled={loadingSnapshots[row.original.id]}
                  >
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}

const getStateIcon = (state: SnapshotState) => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return <CheckCircle className="w-4 h-4 flex-shrink-0" />
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
    default:
      return <Timer className="w-4 h-4 flex-shrink-0" />
  }
}

const getStateColor = (state: SnapshotState) => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return 'text-green-500'
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state: SnapshotState) => {
  // TODO: remove when removing is migrated to deleted
  if (state === SnapshotState.REMOVING) {
    return 'Deleting'
  }
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}
