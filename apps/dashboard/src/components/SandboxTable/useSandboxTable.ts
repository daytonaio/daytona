/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { getRegionFullDisplayName } from '@/lib/utils'
import { Region, Sandbox } from '@daytona/api-client'
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
  VisibilityState,
} from '@tanstack/react-table'
import { useEffect, useMemo, useState } from 'react'
import { getColumns } from './columns'
import { FacetedFilterOption } from './types'

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
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  handleScreenRecordings: (id: string) => void
  regionsData: Region[]
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
  handleOpenTerminal: (sandbox: Sandbox) => void
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
  handleCreateSshAccess,
  handleRevokeSshAccess,
  handleScreenRecordings,
  regionsData,
  handleRecover,
  getRegionName,
  handleCreateSnapshot,
  handleFork,
  handleViewForks,
  handleOpenTerminal,
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
  const [globalFilter, setGlobalFilter] = useState('')

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
        handleOpenTerminal,
      }),
    [
      handleStart,
      handleStop,
      handleDelete,
      handleArchive,
      handleVnc,
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
      handleOpenTerminal,
    ],
  )

  const table = useReactTable({
    data,
    columns,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    globalFilterFn: (row, _columnId, filterValue) => {
      const sandbox = row.original
      const searchValue = String(filterValue).trim().toLowerCase()

      if (!searchValue) {
        return true
      }

      const regionName = getRegionName(sandbox.target) ?? sandbox.target
      const labels = Object.entries(sandbox.labels ?? {})
        .map(([key, value]) => `${key}: ${value}`)
        .join(' ')

      return [sandbox.name, sandbox.id, sandbox.state, sandbox.snapshot ?? '', regionName, labels].some((value) =>
        String(value).toLowerCase().includes(searchValue),
      )
    },
    state: {
      globalFilter,
      sorting,
      columnFilters,
      columnVisibility,
      columnPinning: {
        left: ['select', 'name'],
        right: ['actions'],
      },
    },
    onColumnVisibilityChange: setColumnVisibility,
    defaultColumn: {
      minSize: 0,
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
    globalFilter,
    labelOptions,
    regionOptions,
    sorting,
    columnFilters,
  }
}
