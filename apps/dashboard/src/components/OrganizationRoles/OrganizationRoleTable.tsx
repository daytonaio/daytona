/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DeleteOrganizationRoleDialog } from '@/components/OrganizationRoles/DeleteOrganizationRoleDialog'
import { UpdateOrganizationRoleDialog } from '@/components/OrganizationRoles/UpdateOrganizationRoleDialog'
import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { OrganizationRole, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { MoreHorizontal, Shield } from 'lucide-react'
import { useState } from 'react'
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
      columnPinning: {
        right: ['actions'],
      },
    },
  })

  const isEmpty = !loadingData && table.getRowModel().rows.length === 0

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
      <div className="flex min-h-0 flex-1 flex-col pt-2">
        <TableContainer
          className={isEmpty ? 'min-h-[26rem]' : undefined}
          empty={
            isEmpty ? (
              <TableEmptyState
                overlay
                colSpan={columns.length}
                message="No Roles found."
                icon={<Shield className="h-5 w-5" />}
                description="Create custom roles to manage permissions in your organization."
              />
            ) : undefined
          }
        >
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead
                        key={header.id}
                        className={cn(header.column.id === 'actions' && 'sticky right-0 z-[2]')}
                        sticky={header.column.id === 'actions' ? 'right' : undefined}
                      >
                        {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    )
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {loadingData ? (
                <>
                  {Array.from({ length: 25 }).map((_, i) => (
                    <TableRow key={i} className="h-14">
                      {columns.map((column, colIndex) => (
                        <TableCell
                          key={colIndex}
                          className={cn(column.id === 'actions' && 'sticky right-0 z-[1]')}
                          sticky={column.id === 'actions' ? 'right' : undefined}
                        >
                          <Skeleton className="h-4 w-3/4" />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </>
              ) : table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className={loadingRoleAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        className={cn(cell.column.id === 'actions' && 'sticky right-0 z-[1]')}
                        sticky={cell.column.id === 'actions' ? 'right' : undefined}
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
          <Pagination table={table} entityName="Roles" />
        </PageFooterPortal>
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
          <Tooltip>
            <TooltipTrigger>
              <div className="truncate max-w-md cursor-text">{row.original.description}</div>
            </TooltipTrigger>
            <TooltipContent>
              <p className="max-w-[300px]">{row.original.description}</p>
            </TooltipContent>
          </Tooltip>
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
        )
      },
    },
    {
      id: 'actions',
      size: 48,
      minSize: 48,
      maxSize: 48,
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
