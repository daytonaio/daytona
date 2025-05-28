/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { Check, X } from 'lucide-react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { OrganizationInvitation } from '@daytonaio/api-client'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { DeclineOrganizationInvitationDialog } from '@/components/UserOrganizationInvitations/DeclineOrganizationInvitationDialog'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from '../TableEmptyState'

interface DataTableProps {
  data: OrganizationInvitation[]
  loadingData: boolean
  onAcceptInvitation: (invitation: OrganizationInvitation) => Promise<boolean>
  onDeclineInvitation: (invitation: OrganizationInvitation) => Promise<boolean>
  loadingInvitationAction: Record<string, boolean>
}

export function UserOrganizationInvitationTable({
  data,
  loadingData,
  onAcceptInvitation,
  onDeclineInvitation,
  loadingInvitationAction,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [invitationToDecline, setInvitationToDecline] = useState<OrganizationInvitation | null>(null)
  const [isDeclineDialogOpen, setIsDeclineDialogOpen] = useState(false)

  const handleDecline = (invitation: OrganizationInvitation) => {
    setInvitationToDecline(invitation)
    setIsDeclineDialogOpen(true)
  }

  const handleConfirmDecline = async () => {
    if (invitationToDecline) {
      const success = await onDeclineInvitation(invitationToDecline)
      if (success) {
        setInvitationToDecline(null)
        setIsDeclineDialogOpen(false)
        return success
      }
    }
    return false
  }

  const columns = getColumns({ onAccept: onAcceptInvitation, onDecline: handleDecline })

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

      {invitationToDecline && (
        <DeclineOrganizationInvitationDialog
          open={isDeclineDialogOpen}
          onOpenChange={(open) => {
            setIsDeclineDialogOpen(open)
            if (!open) {
              setInvitationToDecline(null)
            }
          }}
          onDeclineInvitation={handleConfirmDecline}
          loading={loadingInvitationAction[invitationToDecline.id]}
        />
      )}
    </>
  )
}

const getColumns = ({
  onAccept,
  onDecline,
}: {
  onAccept: (invitation: OrganizationInvitation) => void
  onDecline: (invitation: OrganizationInvitation) => void
}): ColumnDef<OrganizationInvitation>[] => {
  const columns: ColumnDef<OrganizationInvitation>[] = [
    {
      accessorKey: 'organizationName',
      header: 'Organization',
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
      id: 'actions',
      cell: ({ row }) => {
        return (
          <div className="flex justify-end gap-2">
            <Button variant="ghost" className="h-8 w-8 p-0" onClick={() => onAccept(row.original)}>
              <span className="sr-only">Accept invitation</span>
              <Check className="h-4 w-4" />
            </Button>
            <Button variant="ghost" className="h-8 w-8 p-0" onClick={() => onDecline(row.original)}>
              <span className="sr-only">Decline invitation</span>
              <X className="h-4 w-4" />
            </Button>
          </div>
        )
      },
    },
  ]

  return columns
}
