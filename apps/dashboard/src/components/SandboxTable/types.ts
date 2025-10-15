/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Sandbox,
  SandboxState,
  SnapshotDto,
  ListSandboxesPaginatedSortEnum,
  ListSandboxesPaginatedOrderEnum,
  ListSandboxesPaginatedStatesEnum,
  Region,
} from '@daytonaio/api-client'
import { Table, SortingState, ColumnFiltersState } from '@tanstack/react-table'
import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'

export interface SandboxTableProps {
  data: Sandbox[]
  sandboxIsLoading: Record<string, boolean>
  sandboxStateIsTransitioning: Record<string, boolean>
  loading: boolean
  snapshots: SnapshotDto[]
  snapshotsDataIsLoading: boolean
  snapshotsDataHasMore?: boolean
  onChangeSnapshotSearchValue: (name?: string) => void
  regionsData: Region[]
  regionsDataIsLoading: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleBulkDelete: (ids: string[]) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  handleRefresh: () => void
  isRefreshing?: boolean
  onRowClick?: (sandbox: Sandbox) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  sorting: SandboxSorting
  onSortingChange: (sorting: SandboxSorting) => void
  filters: SandboxFilters
  onFiltersChange: (filters: SandboxFilters) => void
}

export interface SandboxTableActionsProps {
  sandbox: Sandbox
  writePermitted: boolean
  deletePermitted: boolean
  isLoading: boolean
  onStart: (id: string) => void
  onStop: (id: string) => void
  onDelete: (id: string) => void
  onArchive: (id: string) => void
  onVnc: (id: string) => void
  onOpenWebTerminal: (id: string) => void
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
}

export interface SandboxTableHeaderProps {
  table: Table<Sandbox>
  regionOptions: FacetedFilterOption[]
  regionsDataIsLoading: boolean
  snapshots: SnapshotDto[]
  snapshotsDataIsLoading: boolean
  snapshotsDataHasMore?: boolean
  onChangeSnapshotSearchValue: (name?: string) => void
  onRefresh: () => void
  isRefreshing?: boolean
}

export interface FacetedFilterOption {
  label: string
  value: string | SandboxState
  icon?: any
}

export const convertTableSortingToApiSorting = (sorting: SortingState): SandboxSorting => {
  if (!sorting.length) {
    return DEFAULT_SANDBOX_SORTING
  }

  const sort = sorting[0]
  let field: ListSandboxesPaginatedSortEnum

  switch (sort.id) {
    case 'lastEvent':
      field = ListSandboxesPaginatedSortEnum.LAST_ACTIVITY_AT
      break
    case 'createdAt':
    default:
      field = ListSandboxesPaginatedSortEnum.CREATED_AT
      break
  }

  return {
    field,
    direction: sort.desc ? ListSandboxesPaginatedOrderEnum.DESC : ListSandboxesPaginatedOrderEnum.ASC,
  }
}

export const convertTableFiltersToApiFilters = (columnFilters: ColumnFiltersState): SandboxFilters => {
  const filters: SandboxFilters = {}

  columnFilters.forEach((filter) => {
    switch (filter.id) {
      case 'name':
        if (filter.value && typeof filter.value === 'string') {
          filters.idOrName = filter.value
        }
        break
      case 'state':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.states = filter.value as ListSandboxesPaginatedStatesEnum[]
        }
        break
      case 'snapshot':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.snapshots = filter.value as string[]
        }
        break
      case 'region':
      case 'target':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.regions = filter.value as string[]
        }
        break
      case 'labels':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          const labelObj: Record<string, string> = {}
          filter.value.forEach((label: string) => {
            const [key, value] = label.split(': ')
            if (key && value) {
              labelObj[key] = value
            }
          })
          if (Object.keys(labelObj).length > 0) {
            filters.labels = labelObj
          }
        }
        break
    }
  })

  return filters
}

export const convertApiSortingToTableSorting = (sorting: SandboxSorting): SortingState => {
  if (!sorting.field || !sorting.direction) {
    return [{ id: 'lastEvent', desc: true }]
  }

  let id: string
  switch (sorting.field) {
    case ListSandboxesPaginatedSortEnum.LAST_ACTIVITY_AT:
      id = 'lastEvent'
      break
    case ListSandboxesPaginatedSortEnum.CREATED_AT:
    default:
      id = 'createdAt'
      break
  }

  return [{ id, desc: sorting.direction === ListSandboxesPaginatedOrderEnum.DESC }]
}

export const convertApiFiltersToTableFilters = (filters: SandboxFilters): ColumnFiltersState => {
  const columnFilters: ColumnFiltersState = []

  if (filters.idOrName) {
    columnFilters.push({ id: 'name', value: filters.idOrName })
  }

  if (filters.states && filters.states.length > 0) {
    columnFilters.push({ id: 'state', value: filters.states })
  }

  if (filters.snapshots && filters.snapshots.length > 0) {
    columnFilters.push({ id: 'snapshot', value: filters.snapshots })
  }

  if (filters.regions && filters.regions.length > 0) {
    columnFilters.push({ id: 'region', value: filters.regions })
  }

  if (filters.labels && Object.keys(filters.labels).length > 0) {
    const labelArray = Object.entries(filters.labels).map(([key, value]) => `${key}: ${value}`)
    columnFilters.push({ id: 'labels', value: labelArray })
  }

  return columnFilters
}
