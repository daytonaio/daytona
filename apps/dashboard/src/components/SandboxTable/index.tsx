/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { flexRender } from '@tanstack/react-table'
import { useState } from 'react'
import { Button } from '../ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '../ui/alert-dialog'
import { Pagination } from '../Pagination'
import { TableEmptyState } from '../TableEmptyState'
import { SandboxTableProps } from './types'
import { useSandboxTable } from './useSandboxTable'
import { SandboxTableHeader } from './SandboxTableHeader'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { cn } from '@/lib/utils'
import { Container, Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { RoutePath } from '@/enums/RoutePath'
import { AnimatePresence, motion } from 'motion/react'
import { useSidebar } from '@/components/ui/sidebar'

export function SandboxTable({
  data,
  loadingSandboxes,
  transitioningSandboxes,
  loading,
  snapshots,
  loadingSnapshots,
  handleStart,
  handleStop,
  handleDelete,
  handleBulkDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  onRowClick,
}: SandboxTableProps) {
  const navigate = useNavigate()
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)
  const { state: sidebarState } = useSidebar()

  const { table, labelOptions, regionOptions } = useSandboxTable({
    data,
    loadingSandboxes,
    writePermitted,
    deletePermitted,
    handleStart,
    handleStop,
    handleDelete,
    handleArchive,
    handleVnc,
    getWebTerminalUrl,
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

  return (
    <>
      <SandboxTableHeader
        table={table}
        labelOptions={labelOptions}
        regionOptions={regionOptions}
        snapshots={snapshots}
        loadingSnapshots={loadingSnapshots}
      />

      <Table className="border-separate border-spacing-0">
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
                className={`${
                  loadingSandboxes[row.original.id]
                    ? 'opacity-80 pointer-events-none'
                    : '[&:hover>*:nth-child(2)]:underline'
                } ${
                  transitioningSandboxes[row.original.id]
                    ? 'bg-muted transition-colors duration-300 animate-pulse'
                    : 'transition-colors duration-300'
                } ${onRowClick ? 'cursor-pointer' : ''}`}
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
                    className="border-b border-border"
                    style={{
                      width: cell.column.id === 'id' ? '35%' : cell.column.id === 'select' ? '30px' : 'auto',
                      maxWidth: cell.column.id === 'select' ? '30px' : cell.column.getSize() + 80,
                      minWidth: cell.column.id === 'select' ? '30px' : cell.column.getSize(),
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

      <div className="flex items-center justify-end">
        <Pagination className="pb-2 pt-6" table={table} selectionEnabled entityName="Sandboxes" />
      </div>

      {/* Floating Action Bar */}
      <AnimatePresence>
        {hasSelection && (
          <motion.div
            initial={{ scale: 0.9, opacity: 0, y: 56, x: '-50%' }}
            animate={{ scale: 1, opacity: 1, y: 0, x: '-50%' }}
            exit={{ scale: 0.9, opacity: 0, y: 56, x: '-50%' }}
            className="dark fixed bottom-5 z-50 w-full max-w-md"
            style={{
              left:
                sidebarState === 'collapsed'
                  ? 'calc(50% + var(--sidebar-width-icon, 65px) / 2)'
                  : 'calc(50% + var(--sidebar-width, 16rem) / 2)',
            }}
          >
            <div className="bg-background text-foreground border border-border rounded-lg shadow-lg pl-3 pr-2 py-1 flex items-center justify-between gap-4">
              <div className="text-sm text-muted-foreground">
                {selectedCount} {selectedCount === 1 ? 'sandbox' : 'sandboxes'} selected
              </div>
              <AlertDialog open={bulkDeleteDialogOpen} onOpenChange={setBulkDeleteDialogOpen}>
                <AlertDialogTrigger asChild>
                  <Button variant="ghost" size="sm" className="h-8">
                    <Trash2 className="w-4 h-4" />
                    Delete {selectedCount > 1 ? 'All' : ''}
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete Sandboxes</AlertDialogTitle>
                    <AlertDialogDescription>
                      Are you sure you want to delete{' '}
                      {selectedCount === 1 ? 'this sandbox' : `these ${selectedCount} sandboxes`}? This action cannot be
                      undone.
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
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
