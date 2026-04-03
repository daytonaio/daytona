/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { SelectionToast } from '@/components/SelectionToast'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableFacetedFilter, FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useCommandPaletteActions } from '@/components/CommandPalette'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getRelativeTimeString } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, VolumeDto, VolumeState } from '@daytona/api-client'
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
import { AlertTriangle, CheckCircle, HardDrive, Loader2, MoreHorizontal, Timer } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { TableEmptyState } from './TableEmptyState'
import { VolumeBulkAction, VolumeBulkActionAlertDialog } from './VolumeTable/BulkActionAlertDialog'
import { getVolumeBulkActionCounts, isVolumeDeletable, useVolumeCommands } from './VolumeTable/useVolumeCommands'

interface VolumeTableProps {
  data: VolumeDto[]
  loading: boolean
  processingVolumeAction: Record<string, boolean>
  onDelete: (volume: VolumeDto) => void
  onBulkDelete: (volumes: VolumeDto[]) => void
  onCreateVolume?: () => void
}

export function VolumeTable({
  data,
  loading,
  processingVolumeAction,
  onDelete,
  onBulkDelete,
  onCreateVolume,
}: VolumeTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const { setIsOpen } = useCommandPaletteActions()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_VOLUMES),
    [authenticatedUserHasPermission],
  )
  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_VOLUMES),
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
    enableRowSelection: deletePermitted,
    getRowId: (row) => row.id,
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })
  const selectedRows = table.getSelectedRowModel().rows
  const hasSelection = selectedRows.length > 0
  const selectedVolumes = selectedRows.map((row) => row.original)
  const selectableCount = table.getRowModel().rows.filter((row) => {
    const volume = row.original
    return (
      isVolumeDeletable(volume) &&
      !processingVolumeAction[volume.id] &&
      volume.state !== VolumeState.PENDING_DELETE &&
      volume.state !== VolumeState.DELETING
    )
  }).length
  const bulkActionCounts = useMemo(() => getVolumeBulkActionCounts(selectedVolumes), [selectedVolumes])
  const [pendingBulkAction, setPendingBulkAction] = useState<VolumeBulkAction | null>(null)

  const toggleAllRowsSelected = useCallback(
    (selected: boolean) => {
      if (selected) {
        for (const row of table.getRowModel().rows) {
          const isProcessing = processingVolumeAction[row.original.id]
          const isDeleting =
            row.original.state === VolumeState.PENDING_DELETE || row.original.state === VolumeState.DELETING

          if (!isProcessing && !isDeleting && isVolumeDeletable(row.original)) {
            row.toggleSelected(true)
          }
        }
      } else {
        table.toggleAllRowsSelected(false)
      }
    },
    [table, processingVolumeAction],
  )

  useVolumeCommands({
    writePermitted,
    deletePermitted,
    selectedCount: selectedRows.length,
    selectableCount,
    toggleAllRowsSelected,
    bulkActionCounts,
    onDelete: () => setPendingBulkAction(VolumeBulkAction.Delete),
    onCreateVolume,
  })

  const handleBulkActionConfirm = () => {
    if (pendingBulkAction === VolumeBulkAction.Delete) {
      onBulkDelete(selectedVolumes.filter(isVolumeDeletable))
    }

    setPendingBulkAction(null)
    table.toggleAllRowsSelected(false)
  }

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
      <Pagination table={table} selectionEnabled={deletePermitted} entityName="Volumes" className="mt-4" />
      <AnimatePresence>
        {hasSelection && (
          <SelectionToast
            className="absolute bottom-5 left-1/2 -translate-x-1/2 z-50"
            selectedCount={selectedRows.length}
            onClearSelection={() => table.resetRowSelection()}
            onActionClick={() => setIsOpen(true)}
          />
        )}
      </AnimatePresence>
      <VolumeBulkActionAlertDialog
        action={pendingBulkAction}
        count={pendingBulkAction === VolumeBulkAction.Delete ? bulkActionCounts.deletable : 0}
        onConfirm={handleBulkActionConfirm}
        onCancel={() => setPendingBulkAction(null)}
      />
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
          checked={
            table.getIsAllPageRowsSelected() ? true : table.getIsSomePageRowsSelected() ? 'indeterminate' : false
          }
          onCheckedChange={(value) => {
            for (const row of table.getRowModel().rows) {
              const isProcessing = processingVolumeAction[row.original.id]
              const isDeleting =
                row.original.state === VolumeState.PENDING_DELETE || row.original.state === VolumeState.DELETING
              const isDeletable = isVolumeDeletable(row.original)

              if (isProcessing || isDeleting || !isDeletable) {
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
