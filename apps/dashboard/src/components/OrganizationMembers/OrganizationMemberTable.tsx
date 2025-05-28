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
import { OrganizationRole, OrganizationUser, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { RemoveOrganizationMemberDialog } from '@/components/OrganizationMembers/RemoveOrganizationMemberDialog'
import { UpdateOrganizationMemberRoleDialog } from '@/components/OrganizationMembers/UpdateOrganizationMemberRoleDialog'
import { UpdateAssignedOrganizationRolesDialog } from '@/components/OrganizationMembers/UpdateAssignedOrganizationRolesDialog.tsx'
import { capitalize } from '@/lib/utils'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from '../TableEmptyState'

interface DataTableProps {
  data: OrganizationUser[]
  loadingData: boolean
  availableOrganizationRoles: OrganizationRole[]
  loadingAvailableOrganizationRoles: boolean
  onUpdateMemberRole: (userId: string, role: OrganizationUserRoleEnum) => Promise<boolean>
  onUpdateAssignedOrganizationRoles: (userId: string, roleIds: string[]) => Promise<boolean>
  onRemoveMember: (userId: string) => Promise<boolean>
  loadingMemberAction: Record<string, boolean>
  ownerMode: boolean
}

export function OrganizationMemberTable({
  data,
  loadingData,
  availableOrganizationRoles,
  loadingAvailableOrganizationRoles,
  onUpdateMemberRole,
  onUpdateAssignedOrganizationRoles,
  onRemoveMember,
  loadingMemberAction,
  ownerMode,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [memberToUpdate, setMemberToUpdate] = useState<OrganizationUser | null>(null)
  const [isUpdateMemberRoleDialogOpen, setIsUpdateMemberRoleDialogOpen] = useState(false)
  const [isUpdateAssignedRolesDialogOpen, setIsUpdateAssignedRolesDialogOpen] = useState(false)
  const [memberToRemove, setMemberToRemove] = useState<string | null>(null)
  const [isRemoveDialogOpen, setIsRemoveDialogOpen] = useState(false)

  const columns = getColumns({
    onUpdateMemberRole: (member) => {
      setMemberToUpdate(member)
      setIsUpdateMemberRoleDialogOpen(true)
    },
    onUpdateAssignedRoles: (member) => {
      setMemberToUpdate(member)
      setIsUpdateAssignedRolesDialogOpen(true)
    },
    onRemove: (userId: string) => {
      setMemberToRemove(userId)
      setIsRemoveDialogOpen(true)
    },
    ownerMode,
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

  const handleUpdateMemberRole = async (role: OrganizationUserRoleEnum) => {
    if (memberToUpdate) {
      const success = await onUpdateMemberRole(memberToUpdate.userId, role)
      if (success) {
        setMemberToUpdate(null)
        setIsUpdateMemberRoleDialogOpen(false)
      }
      return success
    }
    return false
  }

  const handleUpdateAssignedRoles = async (roleIds: string[]) => {
    if (memberToUpdate) {
      const success = await onUpdateAssignedOrganizationRoles(memberToUpdate.userId, roleIds)
      if (success) {
        setMemberToUpdate(null)
        setIsUpdateAssignedRolesDialogOpen(false)
      }
      return success
    }
    return false
  }

  const handleConfirmRemove = async () => {
    if (memberToRemove) {
      const success = await onRemoveMember(memberToRemove)
      if (success) {
        setMemberToRemove(null)
        setIsRemoveDialogOpen(false)
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
                    className={`h-14 ${loadingMemberAction[row.original.userId] ? 'opacity-50 pointer-events-none' : ''}`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableEmptyState colSpan={columns.length} message="No Members found." />
              )}
            </TableBody>
          </Table>
        </div>
        <Pagination table={table} className="mt-4" entityName="Members" />
      </div>

      {memberToUpdate && (
        <UpdateOrganizationMemberRoleDialog
          open={isUpdateMemberRoleDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateMemberRoleDialogOpen(open)
            if (!open) {
              setMemberToUpdate(null)
            }
          }}
          initialRole={memberToUpdate.role}
          onUpdateMemberRole={handleUpdateMemberRole}
          loading={loadingMemberAction[memberToUpdate.userId]}
        />
      )}

      {memberToUpdate && (
        <UpdateAssignedOrganizationRolesDialog
          open={isUpdateAssignedRolesDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateAssignedRolesDialogOpen(open)
            if (!open) {
              setMemberToUpdate(null)
            }
          }}
          initialData={memberToUpdate.assignedRoles}
          availableRoles={availableOrganizationRoles}
          loadingAvailableRoles={loadingAvailableOrganizationRoles}
          onUpdateAssignedRoles={handleUpdateAssignedRoles}
          loading={loadingMemberAction[memberToUpdate.userId]}
        />
      )}

      {memberToRemove && (
        <RemoveOrganizationMemberDialog
          open={isRemoveDialogOpen}
          onOpenChange={(open) => {
            setIsRemoveDialogOpen(open)
            if (!open) {
              setMemberToRemove(null)
            }
          }}
          onRemoveMember={handleConfirmRemove}
          loading={loadingMemberAction[memberToRemove]}
        />
      )}
    </>
  )
}

const getColumns = ({
  onUpdateMemberRole,
  onUpdateAssignedRoles,
  onRemove,
  ownerMode,
}: {
  onUpdateMemberRole: (member: OrganizationUser) => void
  onUpdateAssignedRoles: (member: OrganizationUser) => void
  onRemove: (userId: string) => void
  ownerMode: boolean
}): ColumnDef<OrganizationUser>[] => {
  const columns: ColumnDef<OrganizationUser>[] = [
    {
      accessorKey: 'email',
      header: 'Email',
    },
    {
      accessorKey: 'role',
      header: () => {
        return <div className="px-3 w-24">Role</div>
      },
      cell: ({ row }) => {
        const role = capitalize(row.original.role)

        if (!ownerMode) {
          return <div className="px-3 text-sm">{role}</div>
        }

        return (
          <Button variant="ghost" className="w-auto px-3" onClick={() => onUpdateMemberRole(row.original)}>
            {role}
          </Button>
        )
      },
    },
  ]

  if (ownerMode) {
    const extraColumns: ColumnDef<OrganizationUser>[] = [
      {
        accessorKey: 'assignedRoles',
        header: () => {
          return <div className="px-3 w-32">Assignments</div>
        },
        cell: ({ row }) => {
          if (row.original.role === OrganizationUserRoleEnum.OWNER) {
            return <div className="px-3 text-sm text-muted-foreground">Full Access</div>
          }

          const roleCount = row.original.assignedRoles?.length || 0
          const roleText = roleCount === 1 ? '1 role' : `${roleCount} roles`

          return (
            <Button variant="ghost" className="w-auto px-3" onClick={() => onUpdateAssignedRoles(row.original)}>
              {roleText}
            </Button>
          )
        },
      },
      {
        id: 'actions',
        cell: ({ row }) => {
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
                  <DropdownMenuItem className="cursor-pointer" onClick={() => onUpdateMemberRole(row.original)}>
                    Change Role
                  </DropdownMenuItem>
                  {row.original.role !== OrganizationUserRoleEnum.OWNER && (
                    <DropdownMenuItem className="cursor-pointer" onClick={() => onUpdateAssignedRoles(row.original)}>
                      Manage Assignments
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuItem
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    onClick={() => onRemove(row.original.userId)}
                  >
                    Remove
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )
        },
      },
    ]
    columns.push(...extraColumns)
  }

  return columns
}
