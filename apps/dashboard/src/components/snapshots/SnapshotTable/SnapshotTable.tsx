/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCommandPaletteActions } from '@/components/CommandPalette'
import { PageFooterPortal } from '@/components/PageLayout'
import { SelectionToast } from '@/components/SelectionToast'
import { Button } from '@/components/ui/button'
import { FacetFilter } from '@/components/ui/facet-filter'
import { Skeleton } from '@/components/ui/skeleton'
import { SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { useCommandPaletteAnalytics } from '@/hooks/useCommandPaletteAnalytics'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import {
  getColumnPinningBorderClasses,
  getColumnPinningClasses,
  getColumnPinningStyles,
  getExplicitColumnSize,
} from '@/lib/utils/table'
import { OrganizationRolePermissionsEnum, SnapshotDto, SnapshotState } from '@daytonaio/api-client'
import { flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { Box } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { Pagination } from '../../Pagination'
import { TableEmptyState } from '../../TableEmptyState'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../ui/table'
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
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
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

const FIXED_COLUMN_IDS = ['select', 'actions']

export function SnapshotTable({
  data,
  loading,
  loadingSnapshots,
  getRegionName,
  onDelete,
  onActivate,
  onDeactivate,
  onCreateSnapshot,
  pagination,
  pageCount,
  totalItems,
  onBulkDelete,
  onBulkActivate,
  onBulkDeactivate,
  onPaginationChange,
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

  const leftPinnedCount = table.getLeftLeafColumns().length

  const selectedRows = table.getSelectedRowModel().rows
  const hasSelection = selectedRows.length > 0
  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters = stateFilter.size > 0

  const [pendingBulkAction, setPendingBulkAction] = useState<SnapshotBulkAction | null>(null)
  const selectedSnapshots = selectedRows.map((row) => row.original)

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
              icon={<Box className="w-8 h-8" />}
              description={
                hasFilters ? undefined : (
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
                  <Button variant="outline" onClick={() => onStateFilterChange(new Set())}>
                    Clear filters
                  </Button>
                ) : undefined
              }
            />
          ) : undefined
        }
      >
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header, headerIndex) => {
                  return (
                    <TableHead
                      key={header.id}
                      className={cn(
                        'px-2',
                        header.column.getCanSort() && 'hover:bg-muted cursor-pointer',
                        !isEmpty && getColumnPinningBorderClasses(header.column, leftPinnedCount, headerIndex),
                        !isEmpty && getColumnPinningClasses(header.column, true),
                      )}
                      style={
                        isEmpty
                          ? undefined
                          : {
                              ...getExplicitColumnSize(header),
                              ...getColumnPinningStyles(header.column, FIXED_COLUMN_IDS),
                            }
                      }
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
                {Array.from({ length: 25 }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, colIndex) => (
                      <TableCell
                        key={column.id}
                        className={cn(
                          getColumnPinningBorderClasses(column, leftPinnedCount, colIndex),
                          getColumnPinningClasses(column),
                        )}
                        style={getColumnPinningStyles(column, FIXED_COLUMN_IDS)}
                      >
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
                  data-state={row.getIsSelected() ? 'selected' : undefined}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell, cellIndex) => (
                    <TableCell
                      className={cn(
                        'px-2',
                        getColumnPinningBorderClasses(cell.column, leftPinnedCount, cellIndex),
                        getColumnPinningClasses(cell.column),
                      )}
                      key={cell.id}
                      style={getColumnPinningStyles(cell.column, FIXED_COLUMN_IDS)}
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
            className="absolute bottom-5 left-1/2 -translate-x-1/2 z-50"
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
