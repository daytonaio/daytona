/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CancelOrganizationInvitationDialog } from '@/components/OrganizationMembers/CancelOrganizationInvitationDialog'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge } from '@/components/ui/badge'
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
import { cn, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationInvitation, UpdateOrganizationInvitationRoleEnum } from '@daytona/api-client'
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
import { MailPlus, MoreHorizontal } from 'lucide-react'
import { useMemo, useState } from 'react'

type OrganizationInvitationTableMeta = {
  onCancel: (invitationId: string) => void
  onUpdate: (invitation: OrganizationInvitation) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    organizationInvitation?: TData extends OrganizationInvitation ? OrganizationInvitationTableMeta : never
  }
}

const getMeta = (table: ReactTable<OrganizationInvitation>) => {
  return table.options.meta?.organizationInvitation as OrganizationInvitationTableMeta
}

interface DataTableProps {
  data: OrganizationInvitation[]
  loadingData: boolean
  onCancelInvitation: (invitationId: string) => Promise<boolean>
  onUpdateInvitation: (
    invitationId: string,
    role: UpdateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ) => Promise<boolean>
  pendingInvitationIds: Set<string>
}

export function OrganizationInvitationTable({
  data,
  loadingData,
  onCancelInvitation,
  onUpdateInvitation,
  pendingInvitationIds,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [invitationToCancel, setInvitationToCancel] = useState<string | null>(null)
  const [isCancelDialogOpen, setIsCancelDialogOpen] = useState(false)
  const [invitationToUpdate, setInvitationToUpdate] = useState<OrganizationInvitation | null>(null)
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = useState(false)

  const handleCancel = (invitationId: string) => {
    setInvitationToCancel(invitationId)
    setIsCancelDialogOpen(true)
  }

  const handleUpdate = (invitation: OrganizationInvitation) => {
    setInvitationToUpdate(invitation)
    setIsUpdateDialogOpen(true)
  }

  const handleConfirmCancel = async () => {
    if (invitationToCancel) {
      const success = await onCancelInvitation(invitationToCancel)
      if (success) {
        setInvitationToCancel(null)
        setIsCancelDialogOpen(false)
      }
      return success
    }
    return false
  }

  const handleConfirmUpdate = async (role: UpdateOrganizationInvitationRoleEnum, assignedRoleIds: string[]) => {
    if (invitationToUpdate) {
      const success = await onUpdateInvitation(invitationToUpdate.id, role, assignedRoleIds)
      if (success) {
        setInvitationToUpdate(null)
        setIsUpdateDialogOpen(false)
      }
      return success
    }
    return false
  }

  const initialInvitationMember = useMemo(
    () =>
      invitationToUpdate
        ? {
            email: invitationToUpdate.email,
            role: invitationToUpdate.role,
            assignedRoleIds: invitationToUpdate.assignedRoles.map((role) => role.id),
          }
        : undefined,
    [invitationToUpdate],
  )

  const table = useReactTable({
    data,
    columns: organizationInvitationColumns,
    meta: {
      organizationInvitation: {
        onCancel: handleCancel,
        onUpdate: handleUpdate,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const invitation = row.original
      const searchValue = String(filterValue).toLowerCase()
      const status = new Date(invitation.expiresAt) < new Date() ? 'expired' : 'pending'

      return (
        invitation.email.toLowerCase().includes(searchValue) ||
        invitation.invitedBy.toLowerCase().includes(searchValue) ||
        status.includes(searchValue)
      )
    },
    state: {
      globalFilter,
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
  const hasSearch = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <>
      <div className="flex min-h-0 flex-col gap-3">
        <SearchInput
          debounced
          value={globalFilter}
          onValueChange={handleChangeFilter}
          placeholder="Search by Email, Inviter, or Status"
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
                colSpan={organizationInvitationColumns.length}
                message={hasSearch ? 'No matching Invitations found.' : 'No Invitations found.'}
                icon={<MailPlus />}
                description={hasSearch ? null : 'No pending invitations for this organization.'}
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
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead
                        key={header.id}
                        sticky={header.column.getIsPinned()}
                        style={getColumnSizeStyles(header.column)}
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
                      'opacity-50 pointer-events-none': pendingInvitationIds.has(row.original.id),
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
        <Pagination table={table} entityName="Invitations" />
      </div>

      {invitationToUpdate && (
        <UpsertOrganizationAccessSheet
          mode="edit"
          open={isUpdateDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateDialogOpen(open)
            if (!open) {
              setInvitationToUpdate(null)
            }
          }}
          trigger={null}
          initialMember={initialInvitationMember}
          title="Update Invitation"
          description="Modify organization access for the invited member."
          onSubmit={({ role, assignedRoleIds }) => handleConfirmUpdate(role, assignedRoleIds)}
          reducedRoleWarning="Removing assignments will reduce the invited member's access when they accept this invitation."
        />
      )}

      {invitationToCancel && (
        <CancelOrganizationInvitationDialog
          open={isCancelDialogOpen}
          onOpenChange={(open) => {
            setIsCancelDialogOpen(open)
            if (!open) {
              setInvitationToCancel(null)
            }
          }}
          onCancelInvitation={handleConfirmCancel}
          loading={invitationToCancel ? pendingInvitationIds.has(invitationToCancel) : false}
        />
      )}
    </>
  )
}

const organizationInvitationColumns: ColumnDef<OrganizationInvitation>[] = [
  {
    accessorKey: 'email',
    header: 'Email',
  },
  {
    accessorKey: 'invitedBy',
    header: 'Invited by',
  },
  {
    accessorKey: 'expiresAt',
    header: 'Expires',
    cell: ({ row }) => {
      const expiresAt = row.original.expiresAt.toString()
      const { relativeTimeString } = getRelativeTimeString(expiresAt)

      return (
        <TimestampTooltip timestamp={expiresAt}>
          <span className="cursor-default">{relativeTimeString}</span>
        </TimestampTooltip>
      )
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => {
      const isExpired = new Date(row.original.expiresAt) < new Date()
      return <Badge variant={isExpired ? 'destructive' : 'secondary'}>{isExpired ? 'Expired' : 'Pending'}</Badge>
    },
  },
  {
    id: 'actions',
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { onCancel, onUpdate } = getMeta(table)
      const isExpired = new Date(row.original.expiresAt) < new Date()

      if (isExpired) {
        return null
      }

      return (
        <div className="text-right">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onUpdate(row.original)}>Edit</DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onClick={() => onCancel(row.original.id)}>
                Cancel
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]
