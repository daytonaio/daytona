/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, SnapshotDto, SnapshotState } from '@daytonaio/api-client'
import { flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { Box } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Pagination } from '../../Pagination'
import { TableEmptyState } from '../../TableEmptyState'
import { Button } from '../../ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '../../ui/popover'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../ui/table'
import { columns } from './columns'
import { convertApiSortingToTableSorting, convertTableSortingToApiSorting } from './utils'

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  onDelete: (snapshot: SnapshotDto) => void
  onBulkDelete?: (snapshots: SnapshotDto[]) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
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
  pagination,
  pageCount,
  totalItems,
  onBulkDelete,
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
        deletePermitted: false,
        loadingSnapshots,
        getRegionName,
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
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)
  const selectedImages = selectedRows.map((row) => row.original)

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedImages.length > 0) {
      onBulkDelete(selectedImages)
    }
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
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
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
      <div className="flex items-center justify-between space-x-2 py-4">
        {deletePermitted && selectedRows.length > 0 && (
          <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
            <PopoverTrigger>
              <Button variant="destructive" size="sm" className="h-8">
                Bulk Delete
              </Button>
            </PopoverTrigger>
            <PopoverContent side="top">
              <div className="flex flex-col gap-4">
                <p>Are you sure you want to delete these Snapshots?</p>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="destructive"
                    onClick={() => {
                      handleBulkDelete()
                      setBulkDeleteConfirmationOpen(false)
                    }}
                  >
                    Delete
                  </Button>
                  <Button variant="outline" onClick={() => setBulkDeleteConfirmationOpen(false)}>
                    Cancel
                  </Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Snapshots" totalItems={totalItems} />
      </div>
    </div>
  )
}
