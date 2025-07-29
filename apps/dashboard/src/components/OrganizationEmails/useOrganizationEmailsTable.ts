/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationEmail } from '@/billing-api/types/OrganizationEmail'
import {
  ColumnFiltersState,
  SortingState,
  useReactTable,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
} from '@tanstack/react-table'
import { useState, useMemo } from 'react'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { getColumns } from './columns'

interface UseOrganizationEmailsTableProps {
  data: OrganizationEmail[]
  loadingEmails: Record<string, boolean>
  handleDelete: (email: string) => void
  handleResendVerification: (email: string) => void
}

export function useOrganizationEmailsTable({
  data,
  loadingEmails,
  handleDelete,
  handleResendVerification,
}: UseOrganizationEmailsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'email',
      desc: false,
    },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const columns = useMemo(
    () =>
      getColumns({
        handleDelete,
        handleResendVerification,
        loadingEmails,
      }),
    [handleDelete, handleResendVerification, loadingEmails],
  )

  const table = useReactTable({
    data,
    columns,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
    },
    defaultColumn: {
      size: 100,
    },
    getRowId: (row) => row.email,
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  return {
    table,
    sorting,
    columnFilters,
  }
}
