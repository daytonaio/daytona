/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DockerRegistry, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useMemo, useState } from 'react'
import { MoreHorizontal, Package } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { DialogTrigger } from './ui/dialog'
import { Pagination } from './Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from './TableEmptyState'

interface DataTableProps {
  data: DockerRegistry[]
  loading: boolean
  onDelete: (id: string) => void
  onEdit: (registry: DockerRegistry) => void
}

export function RegistryTable({ data, loading, onDelete, onEdit }: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const columns = getColumns({ onDelete, onEdit, loading, writePermitted, deletePermitted })
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id} className="px-2">
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
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
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
                message="No Container registries found."
                icon={<Package className="w-8 h-8" />}
                description="Connect to external container registries (e.g., Docker Hub, GCR, ECR) to pull images for your Sandboxes."
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Registries" />
    </div>
  )
}

const getColumns = ({
  onDelete,
  onEdit,
  loading,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (id: string) => void
  onEdit: (registry: DockerRegistry) => void
  loading: boolean
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<DockerRegistry>[] => {
  const columns: ColumnDef<DockerRegistry>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
    },
    {
      accessorKey: 'url',
      header: 'URL',
    },
    {
      accessorKey: 'project',
      header: 'Project',
    },
    {
      accessorKey: 'username',
      header: 'Username',
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if (!writePermitted && !deletePermitted) {
          return null
        }

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              {writePermitted && (
                <DialogTrigger asChild>
                  <DropdownMenuItem onClick={() => onEdit(row.original)} className="cursor-pointer" disabled={loading}>
                    Edit
                  </DropdownMenuItem>
                </DialogTrigger>
              )}
              {deletePermitted && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    disabled={loading}
                    onClick={() => onDelete(row.original.id)}
                  >
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}
