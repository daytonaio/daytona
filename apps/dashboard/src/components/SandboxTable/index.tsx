/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSidebar } from '@/components/ui/sidebar'
import { RoutePath } from '@/enums/RoutePath'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn, pluralize } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, SandboxState } from '@daytonaio/api-client'
import { flexRender, Table as TableType } from '@tanstack/react-table'
import { CheckIcon, CommandIcon, Container, SquareIcon, Trash2Icon } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCommandPaletteActions, useRegisterCommands, type CommandConfig } from '../CommandPalette'
import { Pagination } from '../Pagination'
import { TableEmptyState } from '../TableEmptyState'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog'
import { Button } from '../ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table'
import { SandboxTableHeader } from './SandboxTableHeader'
import { SandboxTableProps } from './types'
import { useSandboxTable } from './useSandboxTable'

function useSandboxCommands({
  table,
  writePermitted,
  deletePermitted,
  onDelete,
}: {
  table: TableType<any>
  writePermitted: boolean
  deletePermitted: boolean
  onDelete: () => void
}) {
  const selectedCount = table.getRowModel().rows.filter((row) => row.getIsSelected()).length
  const totalCount = table.getRowModel().rows.length

  const rootCommands: CommandConfig[] = useMemo(() => {
    const commands: CommandConfig[] = []

    if (totalCount !== selectedCount) {
      commands.push({
        id: 'select-all-sandboxes',
        label: 'Select All Sandboxes',
        icon: <CheckIcon className="w-4 h-4" />,
        onSelect: () => table.toggleAllRowsSelected(true),
        chainable: true,
      })
    }

    if (selectedCount > 0) {
      commands.push({
        id: 'deselect-all-sandboxes',
        label: 'Deselect All Sandboxes',
        icon: <SquareIcon className="w-4 h-4" />,
        onSelect: () => table.toggleAllRowsSelected(false),
        chainable: true,
      })
    }

    if (deletePermitted && selectedCount > 0) {
      commands.push({
        id: 'delete-sandboxes',
        label: `Delete ${pluralize(selectedCount, 'Sandbox', 'Sandboxes')}`,
        icon: <Trash2Icon className="w-4 h-4" />,
        onSelect: onDelete,
      })
    }

    return commands
  }, [table, selectedCount, deletePermitted, onDelete, totalCount])

  useRegisterCommands(rootCommands, { groupId: 'sandbox-actions', groupLabel: 'Sandbox actions', groupOrder: 0 })
}

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
  handleStart,
  handleStop,
  handleDelete,
  handleBulkDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  handleCreateSshAccess,
  handleRevokeSshAccess,
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
  const { state: sidebarState } = useSidebar()

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
    pagination,
    pageCount,
    onPaginationChange,
    sorting,
    onSortingChange,
    filters,
    onFiltersChange,
    regionsData,
    handleRecover,
  })

  const [bulkDeleteDialogOpen, setBulkDeleteDialogOpen] = useState(false)

  const hasSelection = table.getRowModel().rows.some((row) => row.getIsSelected())
  const selectedCount = table.getRowModel().rows.filter((row) => row.getIsSelected()).length

  const handleBulkDeleteConfirm = () => {
    const selectedIds = table
      .getRowModel()
      .rows.filter((row) => row.getIsSelected())
      .map((row) => row.original.id)

    handleBulkDelete(selectedIds)
    setBulkDeleteDialogOpen(false)

    table.toggleAllRowsSelected(false)
  }

  useSandboxCommands({ table, writePermitted, deletePermitted, onDelete: () => setBulkDeleteDialogOpen(true) })
  const { setIsOpen } = useCommandPaletteActions()
  const handleOpenCommandPalette = () => {
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

      <Table className="border-separate border-spacing-0" style={{ tableLayout: 'fixed', width: '100%' }}>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    key={header.id}
                    data-state={header.column.getCanSort() && 'sortable'}
                    onClick={() =>
                      header.column.getCanSort() && header.column.toggleSorting(header.column.getIsSorted() === 'asc')
                    }
                    className={cn(
                      'sticky top-0 z-[3] border-b border-border',
                      header.column.getCanSort() ? 'hover:bg-muted cursor-pointer' : '',
                    )}
                    style={{
                      width: `${header.column.getSize()}px`,
                    }}
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
            <TableRow>
              <TableCell colSpan={table.getAllColumns().length} className="h-10 text-center">
                Loading...
              </TableCell>
            </TableRow>
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
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    onClick={(e) => {
                      if (cell.column.id === 'select' || cell.column.id === 'actions') {
                        e.stopPropagation()
                      }
                    }}
                    className={cn('border-b border-border', {
                      'group-hover/table-row:underline': cell.column.id === 'name',
                    })}
                    style={{
                      width: `${cell.column.getSize()}px`,
                    }}
                    sticky={cell.column.id === 'actions' ? 'right' : undefined}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableEmptyState
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
          )}
        </TableBody>
      </Table>

      <div className="flex items-center justify-end relative">
        <Pagination
          className="pb-2 pt-6"
          table={table}
          selectionEnabled={deletePermitted}
          entityName="Sandboxes"
          totalItems={totalItems}
        />

        <AnimatePresence>
          {hasSelection && (
            <motion.div
              initial={{ scale: 0.9, opacity: 0, y: 20, x: '-50%' }}
              animate={{ scale: 1, opacity: 1, y: 0, x: '-50%' }}
              exit={{ scale: 0.9, opacity: 0, y: 20, x: '-50%' }}
              className="bg-popover absolute bottom-5 left-1/2 -translate-x-1/2 z-50 w-full max-w-xs"
            >
              <div className="bg-background text-foreground border border-border rounded-lg shadow-lg pl-3 pr-1 py-1 flex items-center justify-between gap-4">
                <div className="text-sm">
                  {selectedCount} {selectedCount === 1 ? 'item' : 'items'} selected
                </div>

                <Button variant="ghost" size="sm" className="h-8" onClick={handleOpenCommandPalette}>
                  <CommandIcon className="w-4 h-4" />
                  <span className="text-sm">Actions</span>
                </Button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <AlertDialog open={bulkDeleteDialogOpen} onOpenChange={setBulkDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Sandboxes</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete {selectedCount === 1 ? 'this item' : `these ${selectedCount} items`}? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleBulkDeleteConfirm}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
