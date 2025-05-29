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
import { AlertTriangle, CheckCircle, MoreHorizontal, Timer } from 'lucide-react'
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
import { getRelativeTimeString } from '@/lib/utils'
import { TableEmptyState } from './TableEmptyState'

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  onDelete: (snapshot: SnapshotDto) => void
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
    () => getColumns({ onDelete, onToggleEnabled, loadingSnapshots, writePermitted, deletePermitted }),
    [onDelete, onToggleEnabled, loadingSnapshots, writePermitted, deletePermitted],
  )
  const table = useReactTable({
    data,
    columns,
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
  })

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
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'opacity-60 pointer-events-none' : ''}`}
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
      <Pagination table={table} className="mt-4" entityName="Images" />
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

        if (snapshot.state === SnapshotState.ERROR && !!snapshot.errorReason) {
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
      accessorKey: 'size',
      header: 'Size',
      cell: ({ row }) => {
        const size = row.original.size
        return size ? `${(size * 1024).toFixed(2)} MB` : '-'
      },
    },
    {
      accessorKey: 'entrypoint',
      header: 'Entrypoint',
      cell: ({ row }) => (row.original.entrypoint ? row.original.entrypoint.join(' ') : '-'),
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => getRelativeTimeString(row.original.createdAt).relativeTimeString,
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => getRelativeTimeString(row.original.lastUsedAt).relativeTimeString,
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
