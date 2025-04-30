/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  SortingState,
} from '@tanstack/react-table'
import { ImageDto, ImageState, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { useMemo, useState } from 'react'
import { Button } from './ui/button'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Checkbox } from './ui/checkbox'
import { Pagination } from './Pagination'
import { TooltipProvider, Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'
import { DebouncedInput } from './DebouncedInput'
import { AlertTriangle, ArrowDown, ArrowUp, ArrowUpDown, CheckCircle, MoreHorizontal, Timer } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

interface ImageTableProps {
  data: ImageDto[]
  loading: boolean
  loadingImages: Record<string, boolean>
  onDelete: (image: ImageDto) => void
  onBulkDelete: (ids: string[]) => void
  onToggleEnabled: (image: ImageDto, enabled: boolean) => void
}

export function ImageTable({ data, loading, loadingImages, onDelete, onBulkDelete, onToggleEnabled }: ImageTableProps) {
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
  const [filterValue, setFilterValue] = useState('')
  const [bulkDeleteDialog, setBulkDeleteDialog] = useState(false)

  const columns = useMemo(
    () => getColumns({ onDelete, onToggleEnabled, loadingImages, writePermitted, deletePermitted }),
    [onDelete, onToggleEnabled, loadingImages, writePermitted, deletePermitted],
  )

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    enableRowSelection: true,
    state: { sorting },
    onSortingChange: setSorting,
    getRowId: (row) => row.id,
  })

  const selectedIds = table.getSelectedRowModel().rows.map((r) => r.original.id)

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={filterValue}
          onChange={setFilterValue}
          placeholder="Search images..."
          className="max-w-sm mr-4"
        />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
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
            ) : table.getRowModel().rows.length ? (
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
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No images found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between py-4">
        {selectedIds.length > 0 && (
          <Popover open={bulkDeleteDialog} onOpenChange={setBulkDeleteDialog}>
            <PopoverTrigger>
              <Button variant="destructive" size="sm">
                Bulk Delete ({selectedIds.length})
              </Button>
            </PopoverTrigger>
            <PopoverContent className="space-y-4">
              <p>Are you sure you want to delete {selectedIds.length} image(s)?</p>
              <div className="flex gap-2">
                <Button
                  variant="destructive"
                  onClick={() => {
                    onBulkDelete(selectedIds)
                    setBulkDeleteDialog(false)
                  }}
                >
                  Delete
                </Button>
                <Button variant="outline" onClick={() => setBulkDeleteDialog(false)}>
                  Cancel
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <Pagination table={table} />
      </div>
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
}): ColumnDef<ImageDto>[] => [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={table.getIsAllRowsSelected()}
        onCheckedChange={(value) => table.toggleAllRowsSelected(!!value)}
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
  },
  {
    accessorKey: 'name',
    header: ({ column }) => <SortableHeader column={column} title="Name" />,
  },
  {
    accessorKey: 'state',
    header: ({ column }) => <SortableHeader column={column} title="State" />,
    cell: ({ row }) => {
      const image = row.original
      const color = getStateColor(image.state)

      if (image.state === ImageState.ERROR && image.errorReason) {
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className={`flex items-center gap-2 ${color}`}>
                  {getStateIcon(image.state)} {getStateLabel(image.state)}
                </div>
              </TooltipTrigger>
              <TooltipContent>{image.errorReason}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      }

      return (
        <div className={`flex items-center gap-2 ${color}`}>
          {getStateIcon(image.state)} {getStateLabel(image.state)}
        </div>
      )
    },
  },
  {
    accessorKey: 'size',
    header: ({ column }) => <SortableHeader column={column} title="Size" />,
    cell: ({ row }) => {
      const size = row.original.size
      return size ? `${(size * 1024).toFixed(2)} MB` : '-'
    },
  },
  {
    accessorKey: 'createdAt',
    header: ({ column }) => <SortableHeader column={column} title="Created" />,
    cell: ({ row }) => new Date(row.original.createdAt).toLocaleDateString(),
  },
  {
    accessorKey: 'lastUsedAt',
    header: ({ column }) => <SortableHeader column={column} title="Last Used" />,
    cell: ({ row }) => (row.original.lastUsedAt ? new Date(row.original.lastUsedAt).toLocaleDateString() : '-'),
  },
  {
    id: 'actions',
    cell: ({ row }) => {
      const image = row.original
      if (!writePermitted && !deletePermitted) return null

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {writePermitted && (
              <DropdownMenuItem
                onClick={() => onToggleEnabled(image, !image.enabled)}
                disabled={loadingImages[image.id]}
              >
                {image.enabled ? 'Disable' : 'Enable'}
              </DropdownMenuItem>
            )}
            {deletePermitted && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => onDelete(image)}
                  className="text-red-600 dark:text-red-400"
                  disabled={loadingImages[image.id]}
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

const SortableHeader = ({ column, title }: { column: any; title: string }) => (
  <Button
    variant="ghost"
    onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
    className="px-2 hover:bg-muted/50"
  >
    {title}
    {{
      asc: <ArrowUp className="ml-2 h-4 w-4" />,
      desc: <ArrowDown className="ml-2 h-4 w-4" />,
    }[column.getIsSorted() as string] || <ArrowUpDown className="ml-2 h-4 w-4" />}
  </Button>
)

const getStateIcon = (state: ImageState) => {
  switch (state) {
    case ImageState.ACTIVE:
      return <CheckCircle className="w-4 h-4" />
    case ImageState.ERROR:
      return <AlertTriangle className="w-4 h-4" />
    default:
      return <Timer className="w-4 h-4" />
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
  if (state === ImageState.REMOVING) return 'Deleting'
  return state
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
    .join(' ')
}
