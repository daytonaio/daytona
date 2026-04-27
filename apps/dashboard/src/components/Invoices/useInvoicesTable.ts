/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Invoice } from '@/billing-api/types/Invoice'
import {
  ColumnFiltersState,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { useState } from 'react'
import { invoiceColumns } from './columns'

interface UseInvoicesTableProps {
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  data: Invoice[]
  onViewInvoice?: (invoice: Invoice) => void
  onVoidInvoice?: (invoice: Invoice) => void
  onPayInvoice?: (invoice: Invoice) => void
}

export function useInvoicesTable({
  data,
  pagination,
  pageCount,
  onPaginationChange,
  onViewInvoice,
  onVoidInvoice,
  onPayInvoice,
}: UseInvoicesTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'issuingDate',
      desc: true,
    },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const table = useReactTable({
    data,
    columns: invoiceColumns,
    meta: {
      invoices: {
        onViewInvoice,
        onVoidInvoice,
        onPayInvoice,
      },
    },
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    manualPagination: true,
    pageCount: pageCount,
    onPaginationChange: (updater) => {
      const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
      onPaginationChange(newPagination)
    },
    state: {
      sorting,
      columnFilters,
      columnPinning: {
        right: ['actions'],
      },
      pagination: {
        pageIndex: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    },
    defaultColumn: {
      size: 100,
    },
    getRowId: (row) => row.id,
  })

  return {
    table,
    sorting,
    columnFilters,
  }
}
