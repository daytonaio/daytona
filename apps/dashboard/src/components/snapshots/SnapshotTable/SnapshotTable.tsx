/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCommandPaletteActions } from '@/components/CommandPalette'
import { PageFooterPortal } from '@/components/PageLayout'
import { SearchInput } from '@/components/SearchInput'
import { SelectionToast } from '@/components/SelectionToast'
import { Button } from '@/components/ui/button'
import { FacetFilter } from '@/components/ui/facet-filter'
import { Skeleton } from '@/components/ui/skeleton'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { useCommandPaletteAnalytics } from '@/hooks/useCommandPaletteAnalytics'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationRolePermissionsEnum, SnapshotDto, SnapshotState } from '@daytona/api-client'
import { flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { Box } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { Pagination } from '../../Pagination'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '../../ui/table'
import { SnapshotBulkAction, SnapshotBulkActionAlertDialog } from './BulkActionAlertDialog'
import { columns } from './columns'
import {
  getSnapshotBulkActionCounts,
  isSnapshotActivatable,
  isSnapshotDeactivatable,
  isSnapshotDeletable,
  useSnapshotsCommands,
} from './useSnapshotsCommands'
import { convertApiSortingToTableSorting, convertTableSortingToApiSorting } from './utils'

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  onDelete: (snapshot: SnapshotDto) => void
  onBulkDelete?: (snapshots: SnapshotDto[]) => void
  onBulkDeactivate?: (snapshots: SnapshotDto[]) => void
  onBulkActivate?: (snapshots: SnapshotDto[]) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
  onCreateSnapshot?: () => void
  onRowClick?: (snapshot: SnapshotDto, orderedSnapshots: SnapshotDto[]) => void
  activeSnapshotId?: string
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  searchValue: string
  onSearchChange: (value: string) => void
  sorting: SnapshotSorting
  onSortingChange: (sorting: SnapshotSorting) => void
  stateFilter: Set<string>
  onStateFilterChange: (values: Set<string>) => void
}

const SNAPSHOT_STATE_OPTIONS = [
  { label: 'Active', value: SnapshotState.ACTIVE },
  { label: 'Inactive', value: SnapshotState.INACTIVE },
  { label: 'Building', value: SnapshotState.BUILDING },
  { label: 'Pending', value: SnapshotState.PENDING },
  { label: 'Pulling', value: SnapshotState.PULLING },
  { label: 'Error', value: SnapshotState.ERROR },
  { label: 'Build Failed', value: SnapshotState.BUILD_FAILED },
]

export function SnapshotTable({
  data,
  loading,
  loadingSnapshots,
  getRegionName,
  onDelete,
  onActivate,
  onDeactivate,
  onCreateSnapshot,
  onRowClick,
  activeSnapshotId,
  pagination,
  pageCount,
  totalItems,
  onBulkDelete,
  onBulkActivate,
  onBulkDeactivate,
  onPaginationChange,
  searchValue,
  onSearchChange,
  sorting,
  onSortingChange,
  stateFilter,
  onStateFilterChange,
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

  const tableSorting = useMemo(() => convertApiSortingToTableSorting(sorting), [sorting])

  const selectableCount = useMemo(() => {
    return data.filter(
      (snapshot) => !snapshot.general && !loadingSnapshots[snapshot.id] && snapshot.state !== SnapshotState.REMOVING,
    ).length
  }, [data, loadingSnapshots])

  const table = useReactTable({
    data,
    columns,
    defaultColumn: {
      minSize: 0,
    },
    getCoreRowModel: getCoreRowModel(),
    initialState: {
      columnPinning: {
        left: ['select'],
        right: ['actions'],
      },
    },
    manualSorting: true,
    onSortingChange: (updater) => {
      const newTableSorting = typeof updater === 'function' ? updater(table.getState().sorting) : updater
      const newApiSorting = convertTableSortingToApiSorting(newTableSorting)
      onSortingChange(newApiSorting)
    },
    manualPagination: true,
    pageCount: pageCount || 1,
    onPaginationChange: pagination
      ? (updater) => {
          const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
          onPaginationChange(newPagination)
        }
      : undefined,
    state: {
      sorting: tableSorting,
      pagination: {
        pageIndex: pagination?.pageIndex || 0,
        pageSize: pagination?.pageSize || 10,
      },
    },
    meta: {
      snapshot: {
        writePermitted,
        deletePermitted,
        loadingSnapshots,
        getRegionName,
        selectableCount,
        onDelete,
        loading,
        onActivate,
        onDeactivate,
      },
    },
    getRowId: (row) => row.id,
    enableRowSelection: deletePermitted,
  })

  const selectedRows = table.getSelectedRowModel().rows
  const hasSelection = selectedRows.length > 0
  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters = stateFilter.size > 0 || searchValue.trim().length > 0

  const [pendingBulkAction, setPendingBulkAction] = useState<SnapshotBulkAction | null>(null)
  const selectedSnapshots = selectedRows.map((row) => row.original)

  const handleClearFilters = () => {
    onSearchChange('')
    onStateFilterChange(new Set())
  }

  const bulkActionCounts = useMemo(() => getSnapshotBulkActionCounts(selectedSnapshots), [selectedSnapshots])

  const handleBulkActionConfirm = () => {
    if (!pendingBulkAction) return

    const handlers: Record<SnapshotBulkAction, () => void> = {
      [SnapshotBulkAction.Delete]: () => {
        if (onBulkDelete) {
          onBulkDelete(selectedSnapshots.filter(isSnapshotDeletable))
        }
      },
      [SnapshotBulkAction.Deactivate]: () => {
        if (onBulkDeactivate) {
          onBulkDeactivate(selectedSnapshots.filter(isSnapshotDeactivatable))
        }
      },
    }

    handlers[pendingBulkAction]()
    setPendingBulkAction(null)
    table.toggleAllRowsSelected(false)
  }

  const toggleAllRowsSelected = useCallback(
    (selected: boolean) => {
      if (selected) {
        for (const row of table.getRowModel().rows) {
          const isGeneral = row.original.general
          const isLoading = loadingSnapshots[row.original.id]
          const isRemoving = row.original.state === SnapshotState.REMOVING
          if (!isGeneral && !isLoading && !isRemoving) {
            row.toggleSelected(true)
          }
        }
      } else {
        table.toggleAllRowsSelected(false)
      }
    },
    [table, loadingSnapshots],
  )

  useSnapshotsCommands({
    writePermitted,
    deletePermitted,
    selectedCount: selectedRows.length,
    totalCount: data.length,
    selectableCount,
    toggleAllRowsSelected,
    bulkActionCounts,
    onDelete: () => setPendingBulkAction(SnapshotBulkAction.Delete),
    onDeactivate: () => setPendingBulkAction(SnapshotBulkAction.Deactivate),
    onActivate: () => onBulkActivate?.(selectedSnapshots.filter(isSnapshotActivatable)),
    onCreateSnapshot: onCreateSnapshot,
  })

  const { setIsOpen } = useCommandPaletteActions()
  const { trackOpened } = useCommandPaletteAnalytics()
  const handleOpenCommandPalette = () => {
    trackOpened('snapshot_selection_toast')
    setIsOpen(true)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-2">
        <SearchInput
          debounced
          value={searchValue}
          onValueChange={onSearchChange}
          placeholder="Search by Name"
          containerClassName="max-w-sm"
        />
        <FacetFilter
          title="State"
          className="h-8"
          options={SNAPSHOT_STATE_OPTIONS}
          selectedValues={stateFilter}
          setSelectedValues={onStateFilterChange}
        />
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasFilters ? 'No matching snapshots found.' : 'No Snapshots yet.'}
              icon={<Box />}
              description={
                hasFilters ? null : (
                  <div className="space-y-2">
                    <p>
                      Snapshots are reproducible, pre-configured environments based on any Docker-compatible image. Use
                      them to define language runtimes, dependencies, and tools for your sandboxes.
                    </p>
                    <p>
                      Create one from the Dashboard, CLI, or SDK to get started. <br />
                      <a
                        href="https://www.daytona.io/docs/snapshots"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        Read the Snapshots guide
                      </a>{' '}
                      to learn more.
                    </p>
                  </div>
                )
              }
              action={
                hasFilters ? (
                  <Button variant="outline" onClick={handleClearFilters}>
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
                  data-selected={row.getIsSelected() || row.original.id === activeSnapshotId ? true : undefined}
                  className={cn('group/table-row transition-all', {
                    'opacity-50 pointer-events-none':
                      loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING,
                    'cursor-pointer': onRowClick,
                  })}
                  onClick={() =>
                    onRowClick?.(
                      row.original,
                      table.getRowModel().rows.map((row) => row.original),
                    )
                  }
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      onClick={(event) => {
                        if (cell.column.id === 'select' || cell.column.id === 'actions') {
                          event.stopPropagation()
                        }
                      }}
                      className={cn({ 'group-hover/table-row:underline': cell.column.id === 'name' })}
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
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Snapshots" totalItems={totalItems} />
      </PageFooterPortal>
      <AnimatePresence>
        {hasSelection && (
          <SelectionToast
            className="absolute bottom-[120px] sm:bottom-20 left-1/2 -translate-x-1/2 z-50"
            selectedCount={selectedRows.length}
            onClearSelection={() => table.resetRowSelection()}
            onActionClick={handleOpenCommandPalette}
          />
        )}
      </AnimatePresence>

      <SnapshotBulkActionAlertDialog
        action={pendingBulkAction}
        count={
          pendingBulkAction
            ? {
                [SnapshotBulkAction.Delete]: bulkActionCounts.deletable,
                [SnapshotBulkAction.Deactivate]: bulkActionCounts.deactivatable,
              }[pendingBulkAction]
            : 0
        }
        onConfirm={handleBulkActionConfirm}
        onCancel={() => setPendingBulkAction(null)}
      />
    </div>
  )
}
