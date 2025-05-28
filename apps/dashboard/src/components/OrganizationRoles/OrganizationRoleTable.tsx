/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { MoreHorizontal } from 'lucide-react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { OrganizationRole, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { DeleteOrganizationRoleDialog } from '@/components/OrganizationRoles/DeleteOrganizationRoleDialog'
import { UpdateOrganizationRoleDialog } from '@/components/OrganizationRoles/UpdateOrganizationRoleDialog'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from '../TableEmptyState'

interface DataTableProps {
  data: OrganizationRole[]
  loadingData: boolean
  onUpdateRole: (
    roleId: string,
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ) => Promise<boolean>
  onDeleteRole: (roleId: string) => Promise<boolean>
  loadingRoleAction: Record<string, boolean>
}

export function OrganizationRoleTable({
  data,
  loadingData,
  onUpdateRole,
  onDeleteRole,
  loadingRoleAction,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [roleToDelete, setRoleToDelete] = useState<string | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [roleToUpdate, setRoleToUpdate] = useState<OrganizationRole | null>(null)
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = useState(false)

  const columns = getColumns({
    onUpdate: (role) => {
      setRoleToUpdate(role)
      setIsUpdateDialogOpen(true)
    },
    onDelete: (userId: string) => {
      setRoleToDelete(userId)
      setIsDeleteDialogOpen(true)
    },
  })

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

  const handleUpdateRole = async (
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ) => {
    if (roleToUpdate) {
      const success = await onUpdateRole(roleToUpdate.id, name, description, permissions)
      if (success) {
        setRoleToUpdate(null)
        setIsUpdateDialogOpen(false)
      }
      return success
    }
    return false
  }

  const handleConfirmDeleteRole = async () => {
    if (roleToDelete) {
      const success = await onDeleteRole(roleToDelete)
      if (success) {
        setRoleToDelete(null)
        setIsDeleteDialogOpen(false)
      }
      return success
    }
    return false
  }

  return (
    <>
      <div>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead key={header.id}>
                        {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    )
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {loadingData ? (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    Loading...
                  </TableCell>
                </TableRow>
              ) : table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className={loadingRoleAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableEmptyState colSpan={columns.length} message="No Roles found." />
              )}
            </TableBody>
          </Table>
        </div>
        <Pagination table={table} className="mt-4" entityName="Roles" />
      </div>

      {roleToUpdate && (
        <UpdateOrganizationRoleDialog
          open={isUpdateDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateDialogOpen(open)
            if (!open) {
              setRoleToUpdate(null)
            }
          }}
          initialData={roleToUpdate}
          onUpdateRole={handleUpdateRole}
        />
      )}

      {roleToDelete && (
        <DeleteOrganizationRoleDialog
          open={isDeleteDialogOpen}
          onOpenChange={(open) => {
            setIsDeleteDialogOpen(open)
            if (!open) {
              setRoleToDelete(null)
            }
          }}
          onDeleteRole={handleConfirmDeleteRole}
          loading={loadingRoleAction[roleToDelete]}
        />
      )}
    </>
  )
}

const getColumns = ({
  onUpdate,
  onDelete,
}: {
  onUpdate: (role: OrganizationRole) => void
  onDelete: (roleId: string) => void
}): ColumnDef<OrganizationRole>[] => {
  const columns: ColumnDef<OrganizationRole>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => {
        return <div className="min-w-48">{row.original.name}</div>
      },
    },
    {
      accessorKey: 'description',
      header: 'Description',
      cell: ({ row }) => {
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className="truncate max-w-md cursor-text">{row.original.description}</div>
              </TooltipTrigger>
              <TooltipContent>
                <p className="max-w-[300px]">{row.original.description}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      accessorKey: 'permissions',
      header: () => {
        return <div className="max-w-md px-3">Permissions</div>
      },
      cell: ({ row }) => {
        const permissions = row.original.permissions.join(', ')
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className="truncate max-w-md px-3 cursor-text">{permissions || '-'}</div>
              </TooltipTrigger>
              {permissions && (
                <TooltipContent>
                  <p className="max-w-[300px]">{permissions}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if (row.original.isGlobal) {
          return null
        }
        return (
          <div className="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <span className="sr-only">Open menu</span>
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end">
                <DropdownMenuItem className="cursor-pointer" onClick={() => onUpdate(row.original)}>
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  className="cursor-pointer text-red-600 dark:text-red-400"
                  onClick={() => onDelete(row.original.id)}
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        )
      },
    },
  ]

  return columns
}
