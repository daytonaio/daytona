/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RemoveOrganizationMemberDialog } from '@/components/OrganizationMembers/RemoveOrganizationMemberDialog'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { capitalize, cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationUser, OrganizationUserRoleEnum } from '@daytona/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  Table as ReactTable,
  RowData,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { MoreHorizontal, Users } from 'lucide-react'
import { useMemo, useState } from 'react'

type MemberTableMeta = {
  onUpdateMemberRole: (member: OrganizationUser) => void
  onUpdateAssignedRoles: (member: OrganizationUser) => void
  onRemove: (userId: string) => void
  ownerMode: boolean
  currentUserId?: string
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    member?: TData extends OrganizationUser ? MemberTableMeta : never
  }
}

const getMeta = (table: ReactTable<OrganizationUser>) => {
  return table.options.meta?.member as MemberTableMeta
}

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
  const [globalFilter, setGlobalFilter] = useState('')
  const [memberToUpdate, setMemberToUpdate] = useState<OrganizationUser | null>(null)
  const [isUpdateMemberAccessDialogOpen, setIsUpdateMemberAccessDialogOpen] = useState(false)
  const [memberToRemove, setMemberToRemove] = useState<string | null>(null)
  const [isRemoveDialogOpen, setIsRemoveDialogOpen] = useState(false)

  const table = useReactTable({
    data,
    columns,
    meta: {
      member: {
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
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const member = row.original
      const searchValue = String(filterValue).toLowerCase()

      return (
        member.email.toLowerCase().includes(searchValue) ||
        member.role.toLowerCase().includes(searchValue) ||
        member.assignedRoles.some((assignment) => assignment.name.toLowerCase().includes(searchValue))
      )
    },
    state: {
      globalFilter,
      sorting,
      columnVisibility: {
        assignedRoles: ownerMode,
        actions: ownerMode,
      },
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
  const hasSearch = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

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
      <div className="flex min-h-0 flex-col gap-3">
        <SearchInput
          debounced
          value={globalFilter}
          onValueChange={handleChangeFilter}
          placeholder="Search by Email, Role, or Assignment"
          containerClassName="max-w-sm"
        />
        <TableContainer
          className={cn('max-h-[550px]', {
            'min-h-[20rem]': isEmpty,
          })}
          empty={
            isEmpty ? (
              <TableEmptyState
                overlay
                colSpan={table.getVisibleLeafColumns().length}
                message={hasSearch ? 'No matching Members found.' : 'No Members found.'}
                icon={<Users />}
                description={hasSearch ? null : 'Invite people to collaborate in your organization.'}
                action={
                  hasSearch ? (
                    <Button variant="outline" onClick={() => handleChangeFilter('')}>
                      Clear filters
                    </Button>
                  ) : null
                }
              />
            ) : null
          }
        >
          <Table className="table-fixed" style={{ minWidth: table.getTotalSize() }}>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead
                      key={header.id}
                      sticky={header.column.getIsPinned()}
                      style={getColumnSizeStyles(header.column)}
                    >
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {loadingData ? (
                <>
                  {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                    <TableRow key={i} className="h-14">
                      {table.getVisibleLeafColumns().map((column) => (
                        <TableCell key={column.id} sticky={column.getIsPinned()} style={getColumnSizeStyles(column)}>
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
                    className={cn('h-14', {
                      'opacity-50 pointer-events-none': pendingMemberIds.has(row.original.userId),
                    })}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        sticky={cell.column.getIsPinned()}
                        style={getColumnSizeStyles(cell.column)}
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
        <Pagination table={table} entityName="Members" />
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
    cell: ({ row, table }) => {
      const { ownerMode, currentUserId, onUpdateMemberRole } = getMeta(table)
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
  {
    id: 'assignedRoles',
    accessorKey: 'assignedRoles',
    header: () => {
      return <div className="px-3 w-32">Assignments</div>
    },
    cell: ({ row, table }) => {
      const { currentUserId, onUpdateAssignedRoles } = getMeta(table)
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
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { currentUserId, onUpdateMemberRole, onUpdateAssignedRoles, onRemove } = getMeta(table)
      const canUpdateAccess = row.original.userId !== currentUserId

      return (
        <div className="text-right">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              {canUpdateAccess && (
                <DropdownMenuItem onClick={() => onUpdateMemberRole(row.original)}>Change Role</DropdownMenuItem>
              )}
              {canUpdateAccess && row.original.role !== OrganizationUserRoleEnum.OWNER && (
                <DropdownMenuItem onClick={() => onUpdateAssignedRoles(row.original)}>
                  Manage Assignments
                </DropdownMenuItem>
              )}
              <DropdownMenuItem variant="destructive" onClick={() => onRemove(row.original.userId)}>
                Remove
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]
