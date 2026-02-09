/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox, Region } from '@daytonaio/api-client'
import {
  useReactTable,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getPaginationRowModel,
  VisibilityState,
} from '@tanstack/react-table'
import { useMemo, useState, useEffect } from 'react'
import { FacetedFilterOption } from './types'
import { getColumns } from './columns'
import {
  convertApiSortingToTableSorting,
  convertApiFiltersToTableFilters,
  convertTableSortingToApiSorting,
  convertTableFiltersToApiFilters,
} from './types'
import { SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { getRegionFullDisplayName } from '@/lib/utils'

interface UseSandboxTableProps {
  data: Sandbox[]
  sandboxIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  handleScreenRecordings: (id: string) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  sorting: SandboxSorting
  onSortingChange: (sorting: SandboxSorting) => void
  filters: SandboxFilters
  onFiltersChange: (filters: SandboxFilters) => void
  regionsData: Region[]
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
}

export function useSandboxTable({
  data,
  sandboxIsLoading,
  writePermitted,
  deletePermitted,
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  handleCreateSshAccess,
  handleRevokeSshAccess,
  handleScreenRecordings,
  pagination,
  pageCount,
  onPaginationChange,
  sorting,
  onSortingChange,
  filters,
  onFiltersChange,
  regionsData,
  handleRecover,
  getRegionName,
}: UseSandboxTableProps) {
  // Column visibility state management with persistence
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(() => {
    const saved = getLocalStorageItem(LocalStorageKey.SandboxTableColumnVisibility)
    if (saved) {
      try {
        return JSON.parse(saved)
      } catch {
        return { id: false, labels: false }
      }
    }
    return { id: false, labels: false }
  })

  useEffect(() => {
    setLocalStorageItem(LocalStorageKey.SandboxTableColumnVisibility, JSON.stringify(columnVisibility))
  }, [columnVisibility])

  // Convert API sorting and filters to table format for internal use
  const tableSorting = useMemo(() => convertApiSortingToTableSorting(sorting), [sorting])
  const tableFilters = useMemo(() => convertApiFiltersToTableFilters(filters), [filters])

  const regionOptions: FacetedFilterOption[] = useMemo(() => {
    return regionsData.map((region) => ({
      label: getRegionFullDisplayName(region),
      value: region.id,
    }))
  }, [regionsData])

  const columns = useMemo(
    () =>
      getColumns({
        handleStart,
        handleStop,
        handleDelete,
        handleArchive,
        handleVnc,
        getWebTerminalUrl,
        sandboxIsLoading,
        writePermitted,
        deletePermitted,
        handleCreateSshAccess,
        handleRevokeSshAccess,
        handleRecover,
        getRegionName,
        handleScreenRecordings,
      }),
    [
      handleStart,
      handleStop,
      handleDelete,
      handleArchive,
      handleVnc,
      getWebTerminalUrl,
      sandboxIsLoading,
      writePermitted,
      deletePermitted,
      handleCreateSshAccess,
      handleRevokeSshAccess,
      handleRecover,
      getRegionName,
      handleScreenRecordings,
    ],
  )

  const table = useReactTable({
    data,
    columns,
    manualFiltering: true,
    onColumnFiltersChange: (updater) => {
      const newTableFilters = typeof updater === 'function' ? updater(table.getState().columnFilters) : updater
      const newApiFilters = convertTableFiltersToApiFilters(newTableFilters)
      onFiltersChange(newApiFilters)
    },
    getCoreRowModel: getCoreRowModel(),
    manualSorting: true,
    onSortingChange: (updater) => {
      const newTableSorting = typeof updater === 'function' ? updater(table.getState().sorting) : updater
      const newApiSorting = convertTableSortingToApiSorting(newTableSorting)
      onSortingChange(newApiSorting)
    },
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    manualPagination: true,
    pageCount: pageCount,
    onPaginationChange: (updater) => {
      const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
      onPaginationChange(newPagination)
    },
    getPaginationRowModel: getPaginationRowModel(),
    state: {
      sorting: tableSorting,
      columnFilters: tableFilters,
      columnVisibility,
      pagination: {
        pageIndex: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    },
    onColumnVisibilityChange: setColumnVisibility,
    defaultColumn: {
      size: 100,
    },
    enableRowSelection: deletePermitted,
    getRowId: (row) => row.id,
  })

  return {
    table,
    regionOptions,
  }
}
