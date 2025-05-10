/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ImageDto, ImageState, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
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
import { AlertTriangle, CheckCircle, MoreHorizontal, Timer, XCircle } from 'lucide-react'
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

interface DataTableProps {
  data: ImageDto[]
  loading: boolean
  loadingImages: Record<string, boolean>
  onDelete: (image: ImageDto) => void
  onToggleEnabled: (image: ImageDto, enabled: boolean) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
}

export function ImageTable({
  data,
  loading,
  loadingImages,
  onDelete,
  onToggleEnabled,
  pagination,
  pageCount,
  onPaginationChange,
}: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_IMAGES),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_IMAGES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const columns = useMemo(
    () => getColumns({ onDelete, onToggleEnabled, loadingImages, writePermitted, deletePermitted }),
    [onDelete, onToggleEnabled, loadingImages, writePermitted, deletePermitted],
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
                  className={`${loadingImages[row.original.id] || row.original.state === ImageState.REMOVING ? 'opacity-50 pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              !loading && (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    No results.
                  </TableCell>
                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" />
    </div>
  )
}

const getColumns = ({
  onDelete,
  onToggleEnabled,
  loadingImages,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (image: ImageDto) => void
  onToggleEnabled: (image: ImageDto, enabled: boolean) => void
  loadingImages: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<ImageDto>[] => {
  const columns: ColumnDef<ImageDto>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
    },
    {
      header: 'Status',
      cell: ({ row }) => {
        const image = row.original
        const color = image.enabled ? 'text-green-500' : 'text-red-500'

        return (
          <div className={`flex items-center gap-2 ${color}`}>
            {image.enabled ? <CheckCircle className="w-4 h-4" /> : <XCircle className="w-4 h-4" />}
            {image.enabled ? 'Enabled' : 'Disabled'}
          </div>
        )
      },
    },
    {
      accessorKey: 'state',
      header: 'State',
      cell: ({ row }) => {
        const image = row.original
        const color = getStateColor(image.state)

        if (image.state === ImageState.ERROR && !!image.errorReason) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className={`flex items-center gap-2 ${color}`}>
                    {getStateIcon(image.state)}
                    {getStateLabel(image.state)}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{image.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return (
          <div className={`flex items-center gap-2 ${color}`}>
            {getStateIcon(image.state)}
            {getStateLabel(image.state)}
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
      cell: ({ row }) => new Date(row.original.createdAt).toLocaleDateString(),
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => (row.original.lastUsedAt ? new Date(row.original.lastUsedAt).toLocaleDateString() : '-'),
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if (!writePermitted && !deletePermitted) {
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
                  disabled={loadingImages[row.original.id]}
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
                    disabled={loadingImages[row.original.id]}
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

const getStateIcon = (state: ImageState) => {
  switch (state) {
    case ImageState.ACTIVE:
      return <CheckCircle className="w-4 h-4 flex-shrink-0" />
    case ImageState.ERROR:
      return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
    default:
      return <Timer className="w-4 h-4 flex-shrink-0" />
  }
}

const getStateColor = (state: ImageState) => {
  switch (state) {
    case ImageState.ACTIVE:
      return 'text-green-500'
    case ImageState.ERROR:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state: ImageState) => {
  // TODO: remove when removing is migrated to deleted
  if (state === ImageState.REMOVING) {
    return 'Deleting'
  }
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}
