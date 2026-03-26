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
import { OrganizationInvitation, OrganizationRole, UpdateOrganizationInvitationRoleEnum } from '@daytonaio/api-client'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { CancelOrganizationInvitationDialog } from '@/components/OrganizationMembers/CancelOrganizationInvitationDialog'
import { UpdateOrganizationInvitationDialog } from './UpdateOrganizationInvitationDialog'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from '../TableEmptyState'

interface DataTableProps {
  data: OrganizationInvitation[]
  loadingData: boolean
  availableRoles: OrganizationRole[]
  loadingAvailableRoles: boolean
  onCancelInvitation: (invitationId: string) => Promise<boolean>
  onUpdateInvitation: (
    invitationId: string,
    role: UpdateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ) => Promise<boolean>
  loadingInvitationAction: Record<string, boolean>
}

export function OrganizationInvitationTable({
  data,
  loadingData,
  availableRoles,
  loadingAvailableRoles,
  onCancelInvitation,
  onUpdateInvitation,
  loadingInvitationAction,
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
    },
  })

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
                    className={`h-14 ${loadingInvitationAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''}`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableEmptyState colSpan={columns.length} message="No Invitations found." />
              )}
            </TableBody>
          </Table>
        </div>
        <Pagination table={table} className="mt-4" entityName="Invitations" />
      </div>

      {invitationToUpdate && (
        <UpdateOrganizationInvitationDialog
          open={isUpdateDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateDialogOpen(open)
            if (!open) {
              setInvitationToUpdate(null)
            }
          }}
          invitation={invitationToUpdate}
          availableRoles={availableRoles}
          loadingAvailableRoles={loadingAvailableRoles}
          onUpdateInvitation={handleConfirmUpdate}
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
          loading={loadingInvitationAction[invitationToCancel]}
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
