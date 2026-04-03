/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CancelOrganizationInvitationDialog } from '@/components/OrganizationMembers/CancelOrganizationInvitationDialog'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn } from '@/lib/utils'
import { OrganizationInvitation, UpdateOrganizationInvitationRoleEnum } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { MailPlus, MoreHorizontal } from 'lucide-react'
import { useMemo, useState } from 'react'
import { TableEmptyState } from '../TableEmptyState'

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

  const columns = getColumns({ onCancel: handleCancel, onUpdate: handleUpdate })

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

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col pt-2">
        <TableContainer
          className={isEmpty ? 'min-h-64' : undefined}
          empty={
            isEmpty ? (
              <TableEmptyState
                overlay
                colSpan={columns.length}
                message="No Invitations found."
                icon={<MailPlus className="h-5 w-5" />}
                description="No pending invitations for this organization."
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
                  {Array.from({ length: 5 }).map((_, i) => (
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
                    className={cn('h-14', {
                      'opacity-50 pointer-events-none': pendingInvitationIds.has(row.original.id),
                    })}
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
          <Pagination table={table} entityName="Invitations" />
        </PageFooterPortal>
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

const getColumns = ({
  onCancel,
  onUpdate,
}: {
  onCancel: (invitationId: string) => void
  onUpdate: (invitation: OrganizationInvitation) => void
}): ColumnDef<OrganizationInvitation>[] => {
  const columns: ColumnDef<OrganizationInvitation>[] = [
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
        return new Date(row.original.expiresAt).toLocaleDateString()
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const isExpired = new Date(row.original.expiresAt) < new Date()
        return isExpired ? <span className="text-red-600 dark:text-red-400">Expired</span> : 'Pending'
      },
    },
    {
      id: 'actions',
      size: 48,
      minSize: 48,
      maxSize: 48,
      cell: ({ row }) => {
        const isExpired = new Date(row.original.expiresAt) < new Date()
        if (isExpired) {
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
                  onClick={() => onCancel(row.original.id)}
                >
                  Cancel
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
