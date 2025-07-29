/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { OrganizationEmail } from '@/billing-api/types/OrganizationEmail'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowUp, ArrowDown, CheckCircle, Clock } from 'lucide-react'
import { getRelativeTimeString } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { OrganizationEmailsTableActions } from './OrganizationEmailsTableActions'

interface SortableHeaderProps {
  column: any
  label: string
  dataState?: string
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label, dataState }) => {
  return (
    <div
      role="button"
      onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
      className="flex items-center"
      {...(dataState && { 'data-state': dataState })}
    >
      {label}
      {column.getIsSorted() === 'asc' ? (
        <ArrowUp className="ml-2 h-4 w-4" />
      ) : column.getIsSorted() === 'desc' ? (
        <ArrowDown className="ml-2 h-4 w-4" />
      ) : (
        <div className="ml-2 w-4 h-4" />
      )}
    </div>
  )
}

interface GetColumnsProps {
  handleDelete: (email: string) => void
  handleResendVerification: (email: string) => void
  loadingEmails: Record<string, boolean>
}

export function getColumns({
  handleDelete,
  handleResendVerification,
  loadingEmails,
}: GetColumnsProps): ColumnDef<OrganizationEmail>[] {
  const columns: ColumnDef<OrganizationEmail>[] = [
    {
      id: 'email',
      header: ({ column }) => {
        return <SortableHeader column={column} label="Email" />
      },
      accessorKey: 'email',
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span className="font-medium">
              {row.original.email}
              {row.original.owner && (
                <Badge variant="secondary" className="ml-2">
                  Owner
                </Badge>
              )}
            </span>
          </div>
        )
      },
      sortingFn: (rowA, rowB) => {
        if (rowA.original.owner && !rowB.original.owner) {
          return -1
        }
        return rowA.original.email.localeCompare(rowB.original.email)
      },
    },
    {
      id: 'verified',
      size: 120,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Status" />
      },
      cell: ({ row }) => (
        <div className="max-w-[120px]">
          <Badge
            variant={row.original.verified ? 'default' : 'secondary'}
            className={`flex items-center gap-1 ${
              row.original.verified
                ? 'bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-200'
                : 'bg-gray-100 text-gray-800 dark:bg-gray-950 dark:text-gray-200'
            }`}
          >
            {row.original.verified ? <CheckCircle className="w-3 h-3" /> : <Clock className="w-3 h-3" />}
            {row.original.verified ? 'Verified' : 'Unverified'}
          </Badge>
        </div>
      ),
      accessorKey: 'verified',
      sortingFn: (rowA, rowB) => {
        if (rowA.original.owner && !rowB.original.owner) {
          return -1
        }
        return rowA.original.verified ? 1 : -1
      },
    },
    {
      id: 'verifiedAt',
      size: 160,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Verified At" />
      },
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            {row.original.verifiedAt ? (
              <span>{getRelativeTimeString(row.original.verifiedAt).relativeTimeString}</span>
            ) : (
              <span className="text-muted-foreground">-</span>
            )}
          </div>
        )
      },
      accessorFn: (row) => row.verifiedAt,
      sortingFn: (rowA, rowB) => {
        if (rowA.original.owner && !rowB.original.owner) {
          return -1
        }
        return (rowA.original.verifiedAt?.getTime() ?? 0) - (rowB.original.verifiedAt?.getTime() ?? 0)
      },
    },
    {
      id: 'actions',
      size: 100,
      enableHiding: false,
      cell: ({ row }) => {
        if (row.original.owner) {
          return null
        }

        return (
          <div>
            <OrganizationEmailsTableActions
              email={row.original}
              isLoading={loadingEmails[row.original.email]}
              onDelete={handleDelete}
              onResendVerification={handleResendVerification}
            />
          </div>
        )
      },
    },
  ]

  return columns
}
