/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox, Region } from '@daytona/api-client'
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
  VisibilityState,
} from '@tanstack/react-table'
import { useMemo, useState, useEffect } from 'react'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { FacetedFilterOption } from './types'
import { getColumns } from './columns'
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
  regionsData: Region[]
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
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
  regionsData,
  handleRecover,
  getRegionName,
  handleCreateSnapshot,
  handleFork,
  handleViewForks,
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

  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'lastEvent',
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
        handleCreateSnapshot,
        handleFork,
        handleViewForks,
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
      handleCreateSnapshot,
      handleFork,
      handleViewForks,
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
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
    onColumnVisibilityChange: setColumnVisibility,
    defaultColumn: {
      size: 100,
    },
    enableRowSelection: deletePermitted,
    getRowId: (row) => row.id,
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  return {
    table,
    labelOptions,
    regionOptions,
    sorting,
    columnFilters,
  }
}
