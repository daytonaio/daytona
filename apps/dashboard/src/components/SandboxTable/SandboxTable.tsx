/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { RoutePath } from '@/enums/RoutePath'
import { useCommandPaletteAnalytics } from '@/hooks/useCommandPaletteAnalytics'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn, getRegionFullDisplayName } from '@/lib/utils'
import {
  filterArchivable,
  filterDeletable,
  filterStartable,
  filterStoppable,
  getBulkActionCounts,
  isTransitioning,
} from '@/lib/utils/sandbox'
import { DEFAULT_TABLE_COLUMN, getColumnSizeStyles, getTableSizeStyles } from '@/lib/utils/table'
import { OrganizationRolePermissionsEnum, SandboxListItem, SandboxState } from '@daytona/api-client'
import {
  type ColumnPinningState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  type OnChangeFn,
  useReactTable,
  type VisibilityState,
} from '@tanstack/react-table'
import { Container } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useImperativeHandle, useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import { useCommandPaletteActions } from '../CommandPalette'
import { SelectionToast } from '../SelectionToast'
import { Button } from '../ui/button'
import { Skeleton } from '../ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table'
import { BulkAction, BulkActionAlertDialog } from './BulkActionAlertDialog'
import { columns } from './columns'
import { SandboxTableHeader } from './SandboxTableHeader'
import type { FacetedFilterOption, SandboxTableProps, SandboxTableRef } from './types'
import {
  convertApiFiltersToTableFilters,
  convertApiSortingToTableSorting,
  convertTableFiltersToApiFilters,
  convertTableSortingToApiSorting,
} from './types'
import { useSandboxCommands } from './useSandboxCommands'

const DEFAULT_SANDBOX_TABLE_COLUMN_VISIBILITY: VisibilityState = {
  id: false,
  isPublic: false,
  isRecoverable: false,
  labels: false,
  sandboxClass: false,
}

const DEFAULT_SANDBOX_TABLE_COLUMN_PINNING: ColumnPinningState = {
  left: ['select', 'name'],
  right: ['actions'],
}

function getSandboxTableColumnVisibility(columnVisibility: VisibilityState): VisibilityState {
  return {
    ...columnVisibility,
    isPublic: false,
    isRecoverable: false,
  }
}

export function SandboxTable({
  ref,
  data,
  sandboxIsLoading,
  activeSandboxId,
  loading,
  isShowingPreviousData,
  snapshots,
  snapshotsDataIsLoading,
  snapshotsDataHasMore,
  onChangeSnapshotSearchValue,
  regionsData,
  regionsDataIsLoading,
  getRegionName,
  handleStart,
  handleStop,
  handleDelete,
  handleBulkDelete,
  handleBulkStart,
  handleBulkStop,
  handleBulkArchive,
  handleArchive,
  handleVnc,
  handleCreateSshAccess,
  handleRevokeSshAccess,
  handleScreenRecordings,
  onRowClick,
  handleRecover,
  handleCreateSnapshot,
  handleFork,
  handlePause,
  handleViewForks,
  handleRefresh,
  isRefreshing,
  sorting,
  onSortingChange,
  filters,
  onFiltersChange,
  handleOpenTerminal,
}: SandboxTableProps) {
  const navigate = useNavigate()
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)

  const [columnOrder, setColumnOrder] = useState<string[]>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(DEFAULT_SANDBOX_TABLE_COLUMN_VISIBILITY)
  const [columnPinning, setColumnPinning] = useState<ColumnPinningState>(DEFAULT_SANDBOX_TABLE_COLUMN_PINNING)
  const handleColumnVisibilityChange = useCallback<OnChangeFn<VisibilityState>>((updater) => {
    setColumnVisibility((currentColumnVisibility) => {
      const nextColumnVisibility = typeof updater === 'function' ? updater(currentColumnVisibility) : updater

      return getSandboxTableColumnVisibility(nextColumnVisibility)
    })
  }, [])

  const tableSorting = useMemo(() => convertApiSortingToTableSorting(sorting), [sorting])
  const tableFilters = useMemo(() => convertApiFiltersToTableFilters(filters), [filters])

  const regionOptions: FacetedFilterOption[] = useMemo(() => {
    return regionsData.map((region) => ({
      label: getRegionFullDisplayName(region),
      value: region.id,
    }))
  }, [regionsData])

  const selectableCount = useMemo(() => {
    return data.filter((sandbox) => !sandboxIsLoading[sandbox.id] && sandbox.state !== SandboxState.DESTROYED).length
  }, [sandboxIsLoading, data])
  const table = useReactTable({
    columnResizeMode: 'onEnd',
    data,
    columns,
    manualFiltering: true,
    onColumnFiltersChange: (updater) => {
      const newTableFilters = typeof updater === 'function' ? updater(table.getState().columnFilters) : updater
      onFiltersChange(convertTableFiltersToApiFilters(newTableFilters))
    },
    getCoreRowModel: getCoreRowModel(),
    manualSorting: true,
    onSortingChange: (updater) => {
      const newTableSorting = typeof updater === 'function' ? updater(table.getState().sorting) : updater
      onSortingChange(convertTableSortingToApiSorting(newTableSorting))
    },
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    onColumnOrderChange: setColumnOrder,
    onColumnPinningChange: setColumnPinning,
    state: {
      sorting: tableSorting,
      columnFilters: tableFilters,
      columnOrder,
      columnVisibility,
      columnPinning,
    },
    onColumnVisibilityChange: handleColumnVisibilityChange,
    defaultColumn: DEFAULT_TABLE_COLUMN,
    enableRowSelection: (row) =>
      (writePermitted || deletePermitted) &&
      !sandboxIsLoading[row.original.id] &&
      row.original.state !== SandboxState.DESTROYED,
    meta: {
      sandbox: {
        sandboxIsLoading,
        writePermitted,
        deletePermitted,
        selectableCount,
        handleStart,
        handleStop,
        handleDelete,
        handleArchive,
        handleVnc,
        handleCreateSshAccess,
        handleRevokeSshAccess,
        handleRecover,
        getRegionName,
        handleScreenRecordings,
        handleCreateSnapshot,
        handleFork,
        handlePause,
        handleViewForks,
        handleOpenTerminal,
      },
    },
    getRowId: (row) => row.id,
  })

  useImperativeHandle(
    ref,
    (): SandboxTableRef => ({
      table,
    }),
    [table],
  )

  const [pendingBulkAction, setPendingBulkAction] = useState<BulkAction | null>(null)

  const selectedRows = table.getSelectedRowModel().rows
  const hasSelection = selectedRows.length > 0
  const selectedCount = selectedRows.length
  const selectedSandboxes: SandboxListItem[] = selectedRows.map((row) => row.original)
  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters =
    table.getState().columnFilters.length > 0 || String(table.getState().globalFilter ?? '').trim().length > 0

  const bulkActionCounts = useMemo(() => getBulkActionCounts(selectedSandboxes), [selectedSandboxes])

  const handleBulkActionConfirm = () => {
    if (!pendingBulkAction) return

    const handlers: Record<BulkAction, () => void> = {
      [BulkAction.Delete]: () => handleBulkDelete(filterDeletable(selectedSandboxes).map((s) => s.id)),
      [BulkAction.Start]: () => handleBulkStart(filterStartable(selectedSandboxes).map((s) => s.id)),
      [BulkAction.Stop]: () => handleBulkStop(filterStoppable(selectedSandboxes).map((s) => s.id)),
      [BulkAction.Archive]: () => handleBulkArchive(filterArchivable(selectedSandboxes).map((s) => s.id)),
    }

    handlers[pendingBulkAction]()
    setPendingBulkAction(null)
    table.toggleAllRowsSelected(false)
  }

  const toggleAllRowsSelected = useCallback(
    (selected: boolean) => {
      if (selected) {
        for (const row of table.getRowModel().rows) {
          if (row.getCanSelect()) {
            row.toggleSelected(true)
          }
        }
      } else {
        table.toggleAllRowsSelected(false)
      }
    },
    [table],
  )

  useSandboxCommands({
    writePermitted,
    deletePermitted,
    selectedCount,
    selectableCount,
    toggleAllRowsSelected,
    bulkActionCounts,
    onDelete: () => setPendingBulkAction(BulkAction.Delete),
    onStart: () => setPendingBulkAction(BulkAction.Start),
    onStop: () => setPendingBulkAction(BulkAction.Stop),
    onArchive: () => setPendingBulkAction(BulkAction.Archive),
  })

  const { setIsOpen } = useCommandPaletteActions()
  const { trackOpened } = useCommandPaletteAnalytics()
  const handleOpenCommandPalette = () => {
    trackOpened('sandbox_selection_toast')
    setIsOpen(true)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <SandboxTableHeader
        table={table}
        regionOptions={regionOptions}
        regionsDataIsLoading={regionsDataIsLoading}
        snapshots={snapshots}
        snapshotsDataIsLoading={snapshotsDataIsLoading}
        snapshotsDataHasMore={snapshotsDataHasMore}
        onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
        onRefresh={handleRefresh}
        isRefreshing={isRefreshing}
      />

      <TableContainer
        className={cn({
          'min-h-[26rem]': isEmpty,
        })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={table.getVisibleLeafColumns().length}
              message={hasFilters ? 'No matching sandboxes found.' : 'No Sandboxes yet.'}
              icon={<Container />}
              description={
                hasFilters ? null : (
                  <div className="space-y-2">
                    <p>Spin up a Sandbox to run code in an isolated environment.</p>
                    <p>Use the Daytona SDK or CLI to create one.</p>
                    <p>
                      <button
                        onClick={() => navigate(RoutePath.ONBOARDING)}
                        className="text-primary hover:underline font-medium"
                      >
                        Check out the Onboarding guide
                      </button>{' '}
                      to learn more.
                    </p>
                  </div>
                )
              }
              action={
                hasFilters ? (
                  <Button
                    variant="outline"
                    onClick={() => {
                      table.resetGlobalFilter()
                      table.resetColumnFilters()
                    }}
                  >
                    Clear filters
                  </Button>
                ) : null
              }
            />
          ) : null
        }
      >
        <Table className="table-fixed" style={getTableSizeStyles(table)}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    header={header}
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
                  data-selected={row.getIsSelected() || row.original.id === activeSandboxId ? true : undefined}
                  className={cn('group/table-row transition-all', {
                    'opacity-50': isShowingPreviousData,
                    'opacity-80 pointer-events-none':
                      sandboxIsLoading[row.original.id] || row.original.state === SandboxState.DESTROYED,
                    'bg-muted animate-pulse': isTransitioning(row.original),
                    'cursor-pointer': onRowClick,
                  })}
                  onClick={() => onRowClick?.(row.original)}
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

      <BulkActionAlertDialog
        action={pendingBulkAction}
        count={
          pendingBulkAction
            ? {
                [BulkAction.Delete]: bulkActionCounts.deletable,
                [BulkAction.Start]: bulkActionCounts.startable,
                [BulkAction.Stop]: bulkActionCounts.stoppable,
                [BulkAction.Archive]: bulkActionCounts.archivable,
              }[pendingBulkAction]
            : 0
        }
        onConfirm={handleBulkActionConfirm}
        onCancel={() => setPendingBulkAction(null)}
      />
    </div>
  )
}
