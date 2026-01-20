/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import {
  Region,
  Sandbox,
  SandboxState,
  SnapshotDto,
  SearchSandboxesSortEnum,
  SearchSandboxesOrderEnum,
  SearchSandboxesStatesEnum,
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
  pageSize: number
  hasNextPage: boolean
  hasPreviousPage: boolean
  onNextPage: () => void
  onPreviousPage: () => void
  onPageSizeChange: (pageSize: number) => void
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
  let field: SearchSandboxesSortEnum

  switch (sort.id) {
    case 'name':
      field = SearchSandboxesSortEnum.NAME
      break
    case 'lastEvent':
      field = SearchSandboxesSortEnum.LAST_ACTIVITY_AT
      break
    case 'createdAt':
    default:
      field = SearchSandboxesSortEnum.CREATED_AT
      break
  }

  return {
    field,
    direction: sort.desc ? SearchSandboxesOrderEnum.DESC : SearchSandboxesOrderEnum.ASC,
  }
}

export const convertTableFiltersToApiFilters = (columnFilters: ColumnFiltersState): SandboxFilters => {
  const filters: SandboxFilters = {}

  columnFilters.forEach((filter) => {
    switch (filter.id) {
      case 'name':
        if (filter.value && typeof filter.value === 'string') {
          filters.name = filter.value
        }
        break
      case 'state':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.states = filter.value as SearchSandboxesStatesEnum[]
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
      case 'createdAt':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          const dateRange = filter.value as (Date | undefined)[]
          if (dateRange[0]) {
            filters.createdAtAfter = dateRange[0]
          }
          if (dateRange[1]) {
            filters.createdAtBefore = dateRange[1]
          }
        }
        break
      case 'isPublic':
        if (typeof filter.value === 'boolean') {
          filters.isPublic = filter.value
        }
        break
      case 'isRecoverable':
        if (typeof filter.value === 'boolean') {
          filters.isRecoverable = filter.value
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
    case SearchSandboxesSortEnum.NAME:
      id = 'name'
      break
    case SearchSandboxesSortEnum.LAST_ACTIVITY_AT:
      id = 'lastEvent'
      break
    case SearchSandboxesSortEnum.CREATED_AT:
    default:
      id = 'createdAt'
      break
  }

  return [{ id, desc: sorting.direction === SearchSandboxesOrderEnum.DESC }]
}

export const convertApiFiltersToTableFilters = (filters: SandboxFilters): ColumnFiltersState => {
  const columnFilters: ColumnFiltersState = []

  if (filters.name) {
    columnFilters.push({ id: 'name', value: filters.name })
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

  // Convert createdAt date range filters
  if (filters.createdAtAfter || filters.createdAtBefore) {
    const dateRange: (Date | undefined)[] = [undefined, undefined]
    if (filters.createdAtAfter) dateRange[0] = filters.createdAtAfter
    if (filters.createdAtBefore) dateRange[1] = filters.createdAtBefore
    columnFilters.push({ id: 'createdAt', value: dateRange })
  }

  // Convert boolean filters
  if (filters.isPublic !== undefined) {
    columnFilters.push({ id: 'isPublic', value: filters.isPublic })
  }

  if (filters.isRecoverable !== undefined) {
    columnFilters.push({ id: 'isRecoverable', value: filters.isRecoverable })
  }

  return columnFilters
}
