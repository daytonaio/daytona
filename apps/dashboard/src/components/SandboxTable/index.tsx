/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
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
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationRolePermissionsEnum, Sandbox, SandboxState } from '@daytona/api-client'
import { flexRender } from '@tanstack/react-table'
import { Container } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useImperativeHandle, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCommandPaletteActions } from '../CommandPalette'
import { PageFooterPortal } from '../PageLayout'
import { Pagination } from '../Pagination'
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
import { SandboxTableHeader } from './SandboxTableHeader'
import type { SandboxTableProps, SandboxTableRef } from './types'
import { useSandboxCommands } from './useSandboxCommands'
import { useSandboxTable } from './useSandboxTable'

export function SandboxTable({
  ref,
  data,
  sandboxIsLoading,
  sandboxStateIsTransitioning,
  activeSandboxId,
  loading,
  snapshots,
  loadingSnapshots,
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
  handleViewForks,
  handleOpenTerminal,
}: SandboxTableProps) {
  const navigate = useNavigate()
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)

  const { table, labelOptions, regionOptions } = useSandboxTable({
    data,
    sandboxIsLoading,
    writePermitted,
    deletePermitted,
    regionsData,
    handleStart,
    handleStop,
    handleDelete,
    handleArchive,
    handleVnc,
    handleCreateSshAccess,
    handleRevokeSshAccess,
    handleScreenRecordings,
    handleRecover,
    getRegionName,
    handleCreateSnapshot,
    handleFork,
    handleViewForks,
    handleOpenTerminal,
  })

  useImperativeHandle(
    ref,
    (): SandboxTableRef => ({
      table,
    }),
    [table],
  )

  const [pendingBulkAction, setPendingBulkAction] = useState<BulkAction | null>(null)

  const selectedRows = table.getRowModel().rows.filter((row) => row.getIsSelected())
  const hasSelection = selectedRows.length > 0
  const selectedCount = selectedRows.length
  const totalCount = table.getRowModel().rows.length
  const selectedSandboxes: Sandbox[] = selectedRows.map((row) => row.original)
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
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <SandboxTableHeader
        table={table}
        labelOptions={labelOptions}
        regionOptions={regionOptions}
        regionsDataIsLoading={regionsDataIsLoading}
        snapshots={snapshots}
        loadingSnapshots={loadingSnapshots}
      />

      <TableContainer
        className={cn({
          'min-h-[26rem]': isEmpty,
        })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={table.getAllColumns().length}
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
                      table.setPageIndex(0)
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
        <Table style={{ minWidth: table.getTotalSize() }}>
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
                  data-selected={row.getIsSelected() || row.original.id === activeSandboxId ? true : undefined}
                  className={cn('group/table-row transition-all', {
                    'opacity-80 pointer-events-none':
                      sandboxIsLoading[row.original.id] || row.original.state === SandboxState.DESTROYED,
                    'bg-muted animate-pulse': sandboxStateIsTransitioning[row.original.id],
                    'cursor-pointer': onRowClick,
                  })}
                  onClick={() => onRowClick?.(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      onClick={(e) => {
                        if (cell.column.id === 'select' || cell.column.id === 'actions') {
                          e.stopPropagation()
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
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Sandboxes" totalItems={data.length} />
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
