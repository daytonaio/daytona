/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Invoice } from '@/billing-api/types/Invoice'
import { formatAmount } from '@/lib/utils'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowDown, ArrowUp } from 'lucide-react'
import React from 'react'
import { Badge } from '../ui/badge'
import { InvoicesTableActions } from './InvoicesTableActions'

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
  onViewInvoice?: (invoice: Invoice) => void
  onVoidInvoice?: (invoice: Invoice) => void
}

export function getColumns({ onViewInvoice, onVoidInvoice }: GetColumnsProps): ColumnDef<Invoice>[] {
  const columns: ColumnDef<Invoice>[] = [
    {
      id: 'number',
      header: ({ column }) => {
        return <SortableHeader column={column} label="Invoice" />
      },
      accessorKey: 'number',
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span className="font-medium">{row.original.number}</span>
          </div>
        )
      },
      sortingFn: (rowA, rowB) => {
        return rowA.original.number.localeCompare(rowB.original.number)
      },
    },
    {
      id: 'issuingDate',
      size: 140,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Date" />
      },
      cell: ({ row }) => {
        const date = new Date(row.original.issuingDate)
        return (
          <div className="w-full truncate">
            <span>{date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}</span>
          </div>
        )
      },
      accessorFn: (row) => new Date(row.issuingDate).getTime(),
      sortingFn: (rowA, rowB) => {
        return new Date(rowA.original.issuingDate).getTime() - new Date(rowB.original.issuingDate).getTime()
      },
    },
    {
      id: 'paymentDueDate',
      size: 140,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Due Date" />
      },
      cell: ({ row }) => {
        const date = new Date(row.original.paymentDueDate)
        return (
          <div className="w-full truncate">
            <span>{date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}</span>
          </div>
        )
      },
      accessorFn: (row) => new Date(row.paymentDueDate).getTime(),
      sortingFn: (rowA, rowB) => {
        return new Date(rowA.original.paymentDueDate).getTime() - new Date(rowB.original.paymentDueDate).getTime()
      },
    },
    {
      id: 'totalAmountCents',
      size: 120,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Amount" />
      },
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span>{formatAmount(row.original.totalAmountCents)}</span>
          </div>
        )
      },
      accessorKey: 'totalAmountCents',
      sortingFn: (rowA, rowB) => {
        return rowA.original.totalAmountCents - rowB.original.totalAmountCents
      },
    },
    {
      id: 'paymentStatus',
      size: 120,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Status" />
      },
      cell: ({ row }) => {
        const invoice = row.original
        const isSucceeded = invoice.paymentStatus === 'succeeded'
        const isFailed = invoice.paymentStatus === 'failed'
        const isOverdue = invoice.paymentOverdue

        let variant: 'success' | 'destructive' | 'secondary' = 'secondary'
        let label = 'Pending'

        if (isSucceeded) {
          variant = 'success'
          label = 'Paid'
        } else if (isOverdue || isFailed) {
          variant = 'destructive'
          label = isOverdue ? 'Overdue' : 'Failed'
        }

        if (invoice.status === 'voided') {
          label = 'Voided'
        }

        return (
          <div className="max-w-[120px] flex">
            <Badge variant={variant}>{label}</Badge>
          </div>
        )
      },
      accessorKey: 'paymentStatus',
      sortingFn: (rowA, rowB) => {
        const statusOrder = { succeeded: 0, pending: 1, failed: 2 }
        return (statusOrder[rowA.original.paymentStatus] ?? 3) - (statusOrder[rowB.original.paymentStatus] ?? 3)
      },
    },
    {
      id: 'type',
      size: 120,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Type" />
      },
      cell: ({ row }) => {
        const type = row.original.type
        const displayType = type === 'subscription' ? 'Subscription' : 'One Time'
        return (
          <div className="w-full truncate">
            <span>{displayType}</span>
          </div>
        )
      },
      accessorKey: 'type',
      sortingFn: (rowA, rowB) => {
        return rowA.original.type.localeCompare(rowB.original.type)
      },
    },
    {
      id: 'actions',
      size: 100,
      enableHiding: false,
      cell: ({ row }) => {
        return (
          <div>
            <InvoicesTableActions invoice={row.original} onView={onViewInvoice} onVoid={onVoidInvoice} />
          </div>
        )
      },
    },
  ]

  return columns
}
