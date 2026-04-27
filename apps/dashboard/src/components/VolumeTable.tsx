/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCommandPaletteActions } from '@/components/CommandPalette'
import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { SelectionToast } from '@/components/SelectionToast'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge, BadgeProps } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableFacetedFilter, FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
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
  Table as ReactTable,
  RowData,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { AlertTriangle, CheckCircle, HardDrive, Loader2, MoreHorizontal, Timer } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { VolumeBulkAction, VolumeBulkActionAlertDialog } from './VolumeTable/BulkActionAlertDialog'
import { getVolumeBulkActionCounts, isVolumeDeletable, useVolumeCommands } from './VolumeTable/useVolumeCommands'

type VolumeTableMeta = {
  onDelete: (volume: VolumeDto) => void
  processingVolumeAction: Record<string, boolean>
  deletePermitted: boolean
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    volume?: TData extends VolumeDto ? VolumeTableMeta : never
  }
}

const getMeta = (table: ReactTable<VolumeDto>) => {
  return table.options.meta?.volume as VolumeTableMeta
}

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

  const table = useReactTable({
    data,
    columns,
    meta: {
      volume: { onDelete, processingVolumeAction, deletePermitted },
    },
    defaultColumn: {
      minSize: 0,
    },
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
      columnPinning: {
        left: ['select'],
        right: ['actions'],
      },
    },
  })
  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters = table.getState().columnFilters.length > 0
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
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-2">
        <SearchInput
          debounced
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onValueChange={(value) => table.getColumn('name')?.setFilterValue(value)}
          placeholder="Search by Name, ID, or State"
          containerClassName="max-w-sm"
        />
        {table.getColumn('state') && (
          <DataTableFacetedFilter column={table.getColumn('state')} title="State" options={statuses} />
        )}
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasFilters ? 'No matching volumes found.' : 'No Volumes yet.'}
              icon={<HardDrive />}
              description={
                hasFilters ? null : (
                  <div className="space-y-2">
                    <p>
                      Volumes are shared, persistent directories backed by S3-compatible storage, perfect for reusing
                      datasets, caching dependencies, or passing files across sandboxes.
                    </p>
                    <p>
                      Create one via the SDK or CLI.{' '}
                      <a
                        href="https://www.daytona.io/docs/volumes"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-foreground hover:underline"
                      >
                        Read the Volumes guide
                      </a>{' '}
                      to learn more.
                    </p>
                  </div>
                )
              }
              action={
                hasFilters ? (
                  <Button variant="outline" onClick={() => table.resetColumnFilters()}>
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
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    sticky={header.column.getIsPinned()}
                    style={getColumnSizeStyles(header.column)}
                  >
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <>
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell key={column.id} sticky={column.getIsPinned()} style={getColumnSizeStyles(column)}>
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
                  className={cn({
                    'opacity-50 pointer-events-none':
                      processingVolumeAction[row.original.id] ||
                      row.original.state === VolumeState.PENDING_DELETE ||
                      row.original.state === VolumeState.DELETING,
                  })}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      sticky={cell.column.getIsPinned()}
                      style={getColumnSizeStyles(cell.column)}
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
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Volumes" />
      </PageFooterPortal>
      <AnimatePresence>
        {hasSelection && (
          <SelectionToast
            className="absolute bottom-[120px] sm:bottom-20 left-1/2 -translate-x-1/2 z-50"
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

const getStateBadgeVariant = (state: VolumeState): BadgeProps['variant'] => {
  switch (state) {
    case VolumeState.READY:
      return 'success'
    case VolumeState.ERROR:
      return 'destructive'
    case VolumeState.CREATING:
    case VolumeState.PENDING_CREATE:
    case VolumeState.PENDING_DELETE:
    case VolumeState.DELETING:
      return 'warning'
    case VolumeState.DELETED:
    default:
      return 'secondary'
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

const columns: ColumnDef<VolumeDto>[] = [
  {
    id: 'select',
    size: 44,
    minSize: 44,
    maxSize: 44,
    header: ({ table }) => {
      const { processingVolumeAction } = getMeta(table)
      return (
        <div className="flex justify-center">
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
          />
        </div>
      )
    },
    cell: ({ row, table }) => {
      const { processingVolumeAction } = getMeta(table)
      if (processingVolumeAction[row.original.id]) {
        return (
          <div className="flex justify-center">
            <Loader2 className="w-4 h-4 animate-spin" />
          </div>
        )
      }
      return (
        <div className="flex justify-center">
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
          />
        </div>
      )
    },
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'name',
    size: 200,
    header: 'Name',
    filterFn: (row, _id, filterValue) => {
      const volume = row.original
      const searchValue = String(filterValue).toLowerCase()

      return (
        volume.name.toLowerCase().includes(searchValue) ||
        volume.id.toLowerCase().includes(searchValue) ||
        volume.state.toLowerCase().includes(searchValue) ||
        (volume.errorReason?.toLowerCase().includes(searchValue) ?? false)
      )
    },
    cell: ({ row }) => {
      return <div className="w-40">{row.original.name}</div>
    },
  },
  {
    id: 'state',
    size: 120,
    header: 'State',
    cell: ({ row }) => {
      const volume = row.original
      const state = row.original.state
      const variant = getStateBadgeVariant(state)
      const badge = <Badge variant={variant}>{getStateLabel(state)}</Badge>

      if (state === VolumeState.ERROR && !!volume.errorReason) {
        return (
          <Tooltip>
            <TooltipTrigger asChild>{badge}</TooltipTrigger>
            <TooltipContent>
              <p className="max-w-[300px]">{volume.errorReason}</p>
            </TooltipContent>
          </Tooltip>
        )
      }

      return badge
    },
    accessorKey: 'state',
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
  },
  {
    accessorKey: 'createdAt',
    size: 120,
    header: 'Created',
    cell: ({ row }) => {
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
    accessorKey: 'lastUsedAt',
    size: 120,
    header: 'Last Used',
    cell: ({ row }) => {
      return getRelativeTimeString(row.original.lastUsedAt).relativeTimeString
    },
  },
  {
    id: 'actions',
    header: () => null,
    enableHiding: false,
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { deletePermitted, processingVolumeAction, onDelete } = getMeta(table)

      if (!deletePermitted) {
        return null
      }

      return (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                variant="destructive"
                disabled={processingVolumeAction[row.original.id]}
                onClick={() => onDelete(row.original)}
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
