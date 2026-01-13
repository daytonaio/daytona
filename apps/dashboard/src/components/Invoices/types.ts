/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Invoice } from '@/billing-api'
import { Table } from '@tanstack/react-table'

export interface InvoicesTableProps {
  data: Invoice[]
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  loading: boolean
  onViewInvoice?: (invoice: Invoice) => void
  onVoidInvoice?: (invoice: Invoice) => void
  onRowClick?: (invoice: Invoice) => void
}

export interface InvoicesTableActionsProps {
  invoice: Invoice
  onView?: (invoice: Invoice) => void
  onVoid?: (invoice: Invoice) => void
}

export interface InvoicesTableHeaderProps {
  table: Table<Invoice>
}
