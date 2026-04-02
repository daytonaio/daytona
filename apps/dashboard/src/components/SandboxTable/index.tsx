/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { useCommandPaletteAnalytics } from '@/hooks/useCommandPaletteAnalytics'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import {
  filterArchivable,
  filterDeletable,
  filterStartable,
  filterStoppable,
  getBulkActionCounts,
} from '@/lib/utils/sandbox'
import {
  getColumnPinningBorderClasses,
  getColumnPinningClasses,
  getColumnPinningStyles,
  getExplicitColumnSize,
} from '@/lib/utils/table'
import { OrganizationRolePermissionsEnum, Sandbox, SandboxState } from '@daytonaio/api-client'
import { flexRender } from '@tanstack/react-table'
import { Container } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCommandPaletteActions } from '../CommandPalette'
import { PageFooterPortal } from '../PageLayout'
import { Pagination } from '../Pagination'
import { SelectionToast } from '../SelectionToast'
import { Skeleton } from '../ui/skeleton'
import { TableEmptyState } from '../TableEmptyState'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../ui/table'
import { BulkAction, BulkActionAlertDialog } from './BulkActionAlertDialog'
import { SandboxTableHeader } from './SandboxTableHeader'
import { SandboxTableProps } from './types'
import { useSandboxCommands } from './useSandboxCommands'
import { useSandboxTable } from './useSandboxTable'

const FIXED_COLUMN_IDS = ['select', 'actions']

export function SandboxTable({
  data,
  sandboxIsLoading,
  sandboxStateIsTransitioning,
  loading,
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
  getWebTerminalUrl,
  handleCreateSshAccess,
  handleRevokeSshAccess,
  handleScreenRecordings,
  handleRefresh,
  isRefreshing,
  onRowClick,
  pagination,
  pageCount,
  totalItems,
  onPaginationChange,
  sorting,
  onSortingChange,
  filters,
  onFiltersChange,
  handleRecover,
}: SandboxTableProps) {
  const navigate = useNavigate()
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)

  const { table, regionOptions } = useSandboxTable({
    data,
    sandboxIsLoading,
    writePermitted,
    deletePermitted,
    handleStart,
    handleStop,
    handleDelete,
    handleArchive,
    handleVnc,
    getWebTerminalUrl,
    handleCreateSshAccess,
    handleRevokeSshAccess,
    handleScreenRecordings,
    pagination,
    pageCount,
    onPaginationChange,
    sorting,
    onSortingChange,
    filters,
    onFiltersChange,
    regionsData,
    handleRecover,
    getRegionName,
  })

  const leftPinnedCount = table.getLeftLeafColumns().length

  const [pendingBulkAction, setPendingBulkAction] = useState<BulkAction | null>(null)

  const selectedRows = table.getRowModel().rows.filter((row) => row.getIsSelected())
  const hasSelection = selectedRows.length > 0
  const selectedCount = selectedRows.length
  const totalCount = table.getRowModel().rows.length
  const selectedSandboxes: Sandbox[] = selectedRows.map((row) => row.original)
  const isEmpty = !loading && table.getRowModel().rows.length === 0

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
          const selectDisabled = sandboxIsLoading[row.original.id] || row.original.state === SandboxState.DESTROYED
          if (!selectDisabled) {
            row.toggleSelected(true)
          }
        }
      } else {
        table.toggleAllRowsSelected(selected)
      }
    },
    [sandboxIsLoading, table],
  )

  const selectableCount = useMemo(() => {
    return data.filter((sandbox) => !sandboxIsLoading[sandbox.id] && sandbox.state !== SandboxState.DESTROYED).length
  }, [sandboxIsLoading, data])

  useSandboxCommands({
    writePermitted,
    deletePermitted,
    selectedCount,
    totalCount,
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
    <>
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
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={table.getAllColumns().length}
              message="No Sandboxes yet."
              icon={<Container className="w-8 h-8" />}
              description={
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
                    {table.getVisibleLeafColumns().map((column, colIndex) =>
                      column.id === 'select' || column.id === 'actions' ? (
                        <TableCell
                          key={column.id}
                          className={cn(
                            getColumnPinningBorderClasses(column, leftPinnedCount, colIndex),
                            getColumnPinningClasses(column),
                          )}
                          style={getColumnPinningStyles(column, FIXED_COLUMN_IDS)}
                        />
                      ) : (
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
                      ),
                    )}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={cn('group/table-row transition-all', {
                    'opacity-80 pointer-events-none':
                      sandboxIsLoading[row.original.id] || row.original.state === SandboxState.DESTROYED,
                    'bg-muted animate-pulse': sandboxStateIsTransitioning[row.original.id],
                    'cursor-pointer': onRowClick,
                  })}
                  onClick={() => onRowClick?.(row.original)}
                >
                  {row.getVisibleCells().map((cell, cellIndex) => (
                    <TableCell
                      key={cell.id}
                      onClick={(e) => {
                        if (cell.column.id === 'select' || cell.column.id === 'actions') {
                          e.stopPropagation()
                        }
                      }}
                      className={cn(
                        getColumnPinningBorderClasses(cell.column, leftPinnedCount, cellIndex),
                        getColumnPinningClasses(cell.column),
                        { 'group-hover/table-row:underline': cell.column.id === 'name' },
                      )}
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
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Sandboxes" totalItems={totalItems} />
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
    </>
  )
}
