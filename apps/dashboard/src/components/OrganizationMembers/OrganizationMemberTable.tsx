/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMemo, useState } from 'react'
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
import { OrganizationUser, OrganizationUserRoleEnum } from '@daytona/api-client'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { RemoveOrganizationMemberDialog } from '@/components/OrganizationMembers/RemoveOrganizationMemberDialog'
import { capitalize, cn } from '@/lib/utils'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from '../TableEmptyState'

interface DataTableProps {
  data: OrganizationUser[]
  loadingData: boolean
  onUpdateMemberAccess: (userId: string, role: OrganizationUserRoleEnum, assignedRoleIds: string[]) => Promise<boolean>
  onRemoveMember: (userId: string) => Promise<boolean>
  pendingMemberIds: Set<string>
  ownerMode: boolean
  currentUserId?: string
}

export function OrganizationMemberTable({
  data,
  loadingData,
  onUpdateMemberAccess,
  onRemoveMember,
  pendingMemberIds,
  ownerMode,
  currentUserId,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [memberToUpdate, setMemberToUpdate] = useState<OrganizationUser | null>(null)
  const [isUpdateMemberAccessDialogOpen, setIsUpdateMemberAccessDialogOpen] = useState(false)
  const [memberToRemove, setMemberToRemove] = useState<string | null>(null)
  const [isRemoveDialogOpen, setIsRemoveDialogOpen] = useState(false)

  const columns = getColumns({
    onUpdateMemberRole: (member) => {
      setMemberToUpdate(member)
      setIsUpdateMemberAccessDialogOpen(true)
    },
    onUpdateAssignedRoles: (member) => {
      setMemberToUpdate(member)
      setIsUpdateMemberAccessDialogOpen(true)
    },
    onRemove: (userId: string) => {
      setMemberToRemove(userId)
      setIsRemoveDialogOpen(true)
    },
    ownerMode,
    currentUserId,
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

  const handleUpdateMemberAccess = async (role: OrganizationUserRoleEnum, assignedRoleIds: string[]) => {
    if (memberToUpdate) {
      const success = await onUpdateMemberAccess(memberToUpdate.userId, role, assignedRoleIds)
      if (success) {
        setMemberToUpdate(null)
        setIsUpdateMemberAccessDialogOpen(false)
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

  const initialMemberAccess = useMemo(
    () =>
      memberToUpdate
        ? {
            email: memberToUpdate.email,
            role: memberToUpdate.role,
            assignedRoleIds: memberToUpdate.assignedRoles.map((assignment) => assignment.id),
          }
        : undefined,
    [memberToUpdate],
  )

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
                    className={cn('h-14', {
                      'opacity-50 pointer-events-none': pendingMemberIds.has(row.original.userId),
                    })}
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
        <UpsertOrganizationAccessSheet
          mode="edit"
          trigger={null}
          open={isUpdateMemberAccessDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateMemberAccessDialogOpen(open)
            if (!open) {
              setMemberToUpdate(null)
            }
          }}
          initialMember={initialMemberAccess}
          title="Update Access"
          description="Manage access to the organization with an appropriate role and assignments."
          onSubmit={({ role, assignedRoleIds }) => handleUpdateMemberAccess(role, assignedRoleIds)}
          reducedRoleWarning="Removing assignments will automatically revoke any API keys this member created using permissions granted from those assignments."
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
          loading={memberToRemove ? pendingMemberIds.has(memberToRemove) : false}
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
  currentUserId,
}: {
  onUpdateMemberRole: (member: OrganizationUser) => void
  onUpdateAssignedRoles: (member: OrganizationUser) => void
  onRemove: (userId: string) => void
  ownerMode: boolean
  currentUserId?: string
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
        const canUpdateAccess = row.original.userId !== currentUserId

        if (!ownerMode || !canUpdateAccess) {
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
          const canUpdateAccess = row.original.userId !== currentUserId

          if (row.original.role === OrganizationUserRoleEnum.OWNER) {
            return <div className="px-3 text-sm text-muted-foreground">Full Access</div>
          }

          const roleCount = row.original.assignedRoles?.length || 0
          const roleText = roleCount === 1 ? '1 role' : `${roleCount} roles`

          if (!canUpdateAccess) {
            return <div className="px-3 text-sm">{roleText}</div>
          }

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
          const canUpdateAccess = row.original.userId !== currentUserId

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
                  {canUpdateAccess && (
                    <DropdownMenuItem className="cursor-pointer" onClick={() => onUpdateMemberRole(row.original)}>
                      Change Role
                    </DropdownMenuItem>
                  )}
                  {canUpdateAccess && row.original.role !== OrganizationUserRoleEnum.OWNER && (
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
