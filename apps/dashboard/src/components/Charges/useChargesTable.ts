/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_TABLE_COLUMN } from '@/lib/utils/table'
import { Charge } from '@daytona/billing-api-client'
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
import { useMemo, useState } from 'react'
import { getColumns } from './columns'

interface UseChargesTableProps {
  data: Charge[]
}

export function useChargesTable({ data }: UseChargesTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'createdAt',
      desc: true,
    },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 10 })

  const columns = useMemo(() => getColumns(), [])
  const table = useReactTable({
    columnResizeMode: 'onEnd',
    data,
    columns,
    onColumnFiltersChange: setColumnFilters,
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
      columnPinning: {
        right: ['actions'],
      },
      pagination,
    },
    defaultColumn: {
      ...DEFAULT_TABLE_COLUMN,
      size: 100,
    },
    getRowId: (row, index) => row.id ?? String(index),
  })

  return {
    table,
    sorting,
    columnFilters,
  }
}
