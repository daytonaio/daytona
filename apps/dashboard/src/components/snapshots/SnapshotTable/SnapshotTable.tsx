/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCommandPaletteActions } from '@/components/CommandPalette'
import { SelectionToast } from '@/components/SelectionToast'
import { Skeleton } from '@/components/ui/skeleton'
import { SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, SnapshotDto, SnapshotState } from '@daytonaio/api-client'
import { flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { Box } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import { useCallback, useMemo, useState } from 'react'
import { Pagination } from '../../Pagination'
import { TableEmptyState } from '../../TableEmptyState'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../ui/table'
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
}

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
    getCoreRowModel: getCoreRowModel(),
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
  const handleOpenCommandPalette = () => {
    setIsOpen(true)
  }

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      className={cn('px-2', header.column.getCanSort() && 'hover:bg-muted cursor-pointer')}
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
                {Array.from(new Array(10)).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, i, arr) =>
                      i === arr.length - 1 ? null : (
                        <TableCell key={column.id}>
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
                  data-state={row.getIsSelected() ? 'selected' : undefined}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'pointer-events-none' : ''}`}
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
                message="No Snapshots yet."
                icon={<Box className="w-8 h-8" />}
                description={
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
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination
        table={table}
        selectionEnabled={deletePermitted}
        entityName="Snapshots"
        totalItems={totalItems}
        className="mt-4"
      />
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
