/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import {
  ListSandboxesPaginatedOrderEnum,
  ListSandboxesPaginatedSortEnum,
  ListSandboxesPaginatedStatesEnum,
  Region,
  Sandbox,
  SandboxState,
  SnapshotDto,
} from '@daytonaio/api-client'
import { ColumnFiltersState, SortingState, Table } from '@tanstack/react-table'

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
  getRegionName: (regionId: string) => string | undefined
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleBulkDelete: (ids: string[]) => void
  handleBulkStart: (ids: string[]) => void
  handleBulkStop: (ids: string[]) => void
  handleBulkArchive: (ids: string[]) => void
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
  handleRecover: (id: string) => void
  handleScreenRecordings: (id: string) => void
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
  onRecover: (id: string) => void
  onScreenRecordings: (id: string) => void
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
    case 'name':
      field = ListSandboxesPaginatedSortEnum.NAME
      break
    case 'state':
      field = ListSandboxesPaginatedSortEnum.STATE
      break
    case 'snapshot':
      field = ListSandboxesPaginatedSortEnum.SNAPSHOT
      break
    case 'region':
    case 'target':
      field = ListSandboxesPaginatedSortEnum.REGION
      break
    case 'lastEvent':
    case 'updatedAt':
      field = ListSandboxesPaginatedSortEnum.UPDATED_AT
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
      case 'resources':
        if (filter.value && typeof filter.value === 'object') {
          const resourceValue = filter.value as {
            cpu?: { min?: number; max?: number }
            memory?: { min?: number; max?: number }
            disk?: { min?: number; max?: number }
          }

          if (resourceValue.cpu?.min !== undefined) {
            filters.minCpu = resourceValue.cpu.min
          }
          if (resourceValue.cpu?.max !== undefined) {
            filters.maxCpu = resourceValue.cpu.max
          }
          if (resourceValue.memory?.min !== undefined) {
            filters.minMemoryGiB = resourceValue.memory.min
          }
          if (resourceValue.memory?.max !== undefined) {
            filters.maxMemoryGiB = resourceValue.memory.max
          }
          if (resourceValue.disk?.min !== undefined) {
            filters.minDiskGiB = resourceValue.disk.min
          }
          if (resourceValue.disk?.max !== undefined) {
            filters.maxDiskGiB = resourceValue.disk.max
          }
        }
        break
      case 'lastEvent':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          const dateRange = filter.value as (Date | undefined)[]
          if (dateRange[0]) {
            filters.lastEventAfter = dateRange[0]
          }
          if (dateRange[1]) {
            filters.lastEventBefore = dateRange[1]
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
    case ListSandboxesPaginatedSortEnum.NAME:
      id = 'name'
      break
    case ListSandboxesPaginatedSortEnum.STATE:
      id = 'state'
      break
    case ListSandboxesPaginatedSortEnum.SNAPSHOT:
      id = 'snapshot'
      break
    case ListSandboxesPaginatedSortEnum.REGION:
      id = 'region'
      break
    case ListSandboxesPaginatedSortEnum.UPDATED_AT:
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

  // Convert resource filters back to table format
  const resourceValue: {
    cpu?: { min?: number; max?: number }
    memory?: { min?: number; max?: number }
    disk?: { min?: number; max?: number }
  } = {}

  if (filters.minCpu !== undefined || filters.maxCpu !== undefined) {
    resourceValue.cpu = {}
    if (filters.minCpu !== undefined) resourceValue.cpu.min = filters.minCpu
    if (filters.maxCpu !== undefined) resourceValue.cpu.max = filters.maxCpu
  }

  if (filters.minMemoryGiB !== undefined || filters.maxMemoryGiB !== undefined) {
    resourceValue.memory = {}
    if (filters.minMemoryGiB !== undefined) resourceValue.memory.min = filters.minMemoryGiB
    if (filters.maxMemoryGiB !== undefined) resourceValue.memory.max = filters.maxMemoryGiB
  }

  if (filters.minDiskGiB !== undefined || filters.maxDiskGiB !== undefined) {
    resourceValue.disk = {}
    if (filters.minDiskGiB !== undefined) resourceValue.disk.min = filters.minDiskGiB
    if (filters.maxDiskGiB !== undefined) resourceValue.disk.max = filters.maxDiskGiB
  }

  if (Object.keys(resourceValue).length > 0) {
    columnFilters.push({ id: 'resources', value: resourceValue })
  }

  // Convert date range filters back to table format
  if (filters.lastEventAfter || filters.lastEventBefore) {
    const dateRange: (Date | undefined)[] = [undefined, undefined]
    if (filters.lastEventAfter) dateRange[0] = filters.lastEventAfter
    if (filters.lastEventBefore) dateRange[1] = filters.lastEventBefore
    columnFilters.push({ id: 'lastEvent', value: dateRange })
  }

  return columnFilters
}
