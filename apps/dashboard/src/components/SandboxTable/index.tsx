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

export function SandboxTable({
  data,
  loadingSandboxes,
  loading,
  snapshots,
  loadingSnapshots,
  handleStart,
  handleStop,
  handleDelete,
  handleBulkDelete,
  handleArchive,
  onRowClick,
}: SandboxTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)

  const { table, labelOptions } = useSandboxTable({
    data,
    loadingSandboxes,
    writePermitted,
    deletePermitted,
    handleStart,
    handleStop,
    handleDelete,
    handleArchive,
  })

  const [bulkDeleteDialogOpen, setBulkDeleteDialogOpen] = useState(false)

  const hasSelection = table.getRowModel().rows.some((row) => row.getIsSelected())

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
            <TableEmptyState colSpan={table.getAllColumns().length} message="No Sandboxes found." />
          )}
        </TableBody>
      </Table>

      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          {hasSelection && (
            <AlertDialog open={bulkDeleteDialogOpen} onOpenChange={setBulkDeleteDialogOpen}>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" size="sm" className="h-8">
                  Bulk Delete
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Sandboxes</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete these sandboxes? This action cannot be undone.
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
          )}
        </div>
        <Pagination className="pb-4 pt-6" table={table} selectionEnabled entityName="Sandboxes" />
      </div>
    </>
  )
}
