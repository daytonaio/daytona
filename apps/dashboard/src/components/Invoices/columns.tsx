/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Invoice } from '@/billing-api/types/Invoice'
import { formatAmount } from '@/lib/utils'
import { ColumnDef, RowData, Table } from '@tanstack/react-table'
import React from 'react'
import { SortOrderIcon } from '../SortIcon'
import { Badge } from '../ui/badge'
import { InvoicesTableActions } from './InvoicesTableActions'

export type InvoicesTableMeta = {
  onViewInvoice?: (invoice: Invoice) => void
  onVoidInvoice?: (invoice: Invoice) => void
  onPayInvoice?: (invoice: Invoice) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    invoices?: TData extends Invoice ? InvoicesTableMeta : never
  }
}

interface SortableHeaderProps {
  column: any
  label: string
  dataState?: string
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label, dataState }) => {
  const sortDirection = column.getIsSorted()

  return (
    <button
      type="button"
      onClick={() => column.toggleSorting(sortDirection === 'asc')}
      className="group/sort-header flex h-full w-full items-center gap-2"
      {...(dataState && { 'data-state': dataState })}
    >
      {label}
      <SortOrderIcon sort={sortDirection || null} />
    </button>
  )
}

const getMeta = (table: Table<Invoice>) => {
  return table.options.meta?.invoices as InvoicesTableMeta
}

export const invoiceColumns: ColumnDef<Invoice>[] = [
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
      const statusOrder: Record<string, number> = { succeeded: 0, pending: 1, failed: 2 }
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
    header: () => null,
    size: 48,
    minSize: 48,
    maxSize: 48,
    enableHiding: false,
    cell: ({ row, table }) => {
      const { onViewInvoice, onVoidInvoice, onPayInvoice } = getMeta(table)
      const isViewable = Boolean(row.original.fileUrl)
      const isVoidable =
        row.original.status === 'finalized' &&
        ['pending', 'failed'].includes(row.original.paymentStatus) &&
        row.original.type === 'one_off'
      const isPayable =
        row.original.status === 'finalized' && ['pending', 'failed'].includes(row.original.paymentStatus)

      return (
        <div className="flex justify-center">
          <InvoicesTableActions
            invoice={row.original}
            onView={isViewable ? onViewInvoice : undefined}
            onVoid={isVoidable ? onVoidInvoice : undefined}
            onPay={isPayable ? onPayInvoice : undefined}
          />
        </div>
      )
    },
  },
]
