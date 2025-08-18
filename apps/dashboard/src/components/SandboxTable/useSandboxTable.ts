/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '@daytonaio/api-client'
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
import { FacetedFilterOption } from './types'
import { getColumns } from './columns'

interface UseSandboxTableProps {
  data: Sandbox[]
  loadingSandboxes: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
}

export function useSandboxTable({
  data,
  loadingSandboxes,
  writePermitted,
  deletePermitted,
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  pagination,
  pageCount,
  onPaginationChange,
}: UseSandboxTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'createdAt',
      desc: true,
    },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const labelOptions: FacetedFilterOption[] = useMemo(() => {
    const labels = new Set<string>()
    data.forEach((sandbox) => {
      Object.entries(sandbox.labels ?? {}).forEach(([key, value]) => {
        labels.add(`${key}: ${value}`)
      })
    })
    return Array.from(labels).map((label) => ({ label, value: label }))
  }, [data])

  const regionOptions: FacetedFilterOption[] = useMemo(() => {
    const regions = new Set<string>()
    data.forEach((sandbox) => {
      if (sandbox.target) {
        regions.add(sandbox.target)
      }
    })
    return Array.from(regions).map((region) => ({ label: region, value: region }))
  }, [data])

  const columns = useMemo(
    () =>
      getColumns({
        handleStart,
        handleStop,
        handleDelete,
        handleArchive,
        handleVnc,
        getWebTerminalUrl,
        loadingSandboxes,
        writePermitted,
        deletePermitted,
      }),
    [
      handleStart,
      handleStop,
      handleDelete,
      handleArchive,
      handleVnc,
      getWebTerminalUrl,
      loadingSandboxes,
      writePermitted,
      deletePermitted,
    ],
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
    manualPagination: true,
    pageCount: pageCount,
    onPaginationChange: (updater) => {
      const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
      onPaginationChange(newPagination)
    },
    state: {
      sorting,
      columnFilters,
      pagination: {
        pageIndex: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    },
    defaultColumn: {
      size: 100,
    },
    enableRowSelection: deletePermitted,
    getRowId: (row) => row.id,
  })

  return {
    table,
    labelOptions,
    regionOptions,
    sorting,
    columnFilters,
  }
}
