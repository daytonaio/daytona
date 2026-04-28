/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Button } from '@/components/ui/button'
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
import { DeclineOrganizationInvitationDialog } from '@/components/UserOrganizationInvitations/DeclineOrganizationInvitationDialog'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationInvitation } from '@daytona/api-client'
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
import { Check, Mail, X } from 'lucide-react'
import { useState } from 'react'

type UserOrganizationInvitationTableMeta = {
  onAccept: (invitation: OrganizationInvitation) => Promise<boolean> | void
  onDecline: (invitation: OrganizationInvitation) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    userOrganizationInvitation?: TData extends OrganizationInvitation ? UserOrganizationInvitationTableMeta : never
  }
}

const getMeta = (table: ReactTable<OrganizationInvitation>) => {
  return table.options.meta?.userOrganizationInvitation as UserOrganizationInvitationTableMeta
}

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
  const [globalFilter, setGlobalFilter] = useState('')
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

  const table = useReactTable({
    data,
    columns: userOrganizationInvitationColumns,
    meta: {
      userOrganizationInvitation: {
        onAccept: onAcceptInvitation,
        onDecline: handleDecline,
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

      return (
        invitation.organizationName.toLowerCase().includes(searchValue) ||
        invitation.invitedBy.toLowerCase().includes(searchValue) ||
        new Date(invitation.expiresAt).toLocaleDateString().toLowerCase().includes(searchValue)
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
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div>
        <SearchInput
          debounced
          value={globalFilter}
          onValueChange={handleChangeFilter}
          placeholder="Search by Organization or Inviter"
          containerClassName="max-w-sm"
        />
      </div>
      <TableContainer
        className={cn({
          'min-h-64': isEmpty,
        })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={userOrganizationInvitationColumns.length}
              message={hasSearch ? 'No matching Invitations found.' : 'No Invitations found.'}
              icon={<Mail />}
              description={hasSearch ? null : 'You have no pending organization invitations.'}
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
                  className={`h-14 ${loadingInvitationAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''}`}
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
      <PageFooterPortal>
        <Pagination table={table} entityName="Invitations" />
      </PageFooterPortal>

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
    </div>
  )
}

const userOrganizationInvitationColumns: ColumnDef<OrganizationInvitation>[] = [
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
    id: 'actions',
    size: 80,
    minSize: 80,
    maxSize: 80,
    cell: ({ row, table }) => {
      const { onAccept, onDecline } = getMeta(table)

      return (
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="icon-sm" aria-label="Accept invitation" onClick={() => onAccept(row.original)}>
            <Check className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Decline invitation"
            onClick={() => onDecline(row.original)}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      )
    },
  },
]
