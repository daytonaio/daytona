/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ListSandboxesOrderEnum,
  ListSandboxesSortEnum,
  Region,
  Sandbox,
  SandboxState,
  SnapshotDto,
} from '@daytona/api-client'
import { ColumnFiltersState, SortingState, Table } from '@tanstack/react-table'
import { SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { ResourceFilterValue } from './filters/ResourceFilter'

export interface SandboxTableProps {
  data: Sandbox[]
  sandboxIsLoading: Record<string, boolean>
  sandboxStateIsTransitioning: Record<string, boolean>
  loading: boolean
  snapshots: SnapshotDto[]
  snapshotsDataIsLoading: boolean
  snapshotsDataHasMore?: boolean
  onChangeSnapshotSearchValue?: (name?: string) => void
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
  onRowClick?: (sandbox: Sandbox) => void
  handleRecover: (id: string) => void
  handleScreenRecordings: (id: string) => void
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
  handleRefresh: () => void
  isRefreshing?: boolean
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
}

export interface SandboxTableActionsProps {
  sandbox: Sandbox
  writePermitted: boolean
  deletePermitted: boolean
  isLoading: boolean
  runnerClass?: string
  onStart: (id: string) => void
  onStop: (id: string) => void
  onDelete: (id: string) => void
  onArchive: (id: string) => void
  onVnc: (id: string) => void
  onOpenWebTerminal: (id: string) => void
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
  onFork?: () => void
  onCreateSnapshot?: () => void
  onViewForks?: () => void
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
  onChangeSnapshotSearchValue?: (name?: string) => void
  onRefresh?: () => void
  isRefreshing?: boolean
}

export interface FacetedFilterOption {
  label: string
  value: string | SandboxState
  icon?: any
}

const SORT_FIELD_TO_COLUMN: Record<string, string> = {
  [ListSandboxesSortEnum.NAME]: 'name',
  [ListSandboxesSortEnum.LAST_ACTIVITY_AT]: 'lastEvent',
  [ListSandboxesSortEnum.CREATED_AT]: 'createdAt',
  [ListSandboxesSortEnum.CPU]: 'cpu',
  [ListSandboxesSortEnum.MEMORY_GI_B]: 'memory',
  [ListSandboxesSortEnum.DISK_GI_B]: 'disk',
}

const COLUMN_TO_SORT_FIELD: Record<string, ListSandboxesSortEnum> = {
  name: ListSandboxesSortEnum.NAME,
  lastEvent: ListSandboxesSortEnum.LAST_ACTIVITY_AT,
  createdAt: ListSandboxesSortEnum.CREATED_AT,
}

export function convertApiSortingToTableSorting(sorting: SandboxSorting): SortingState {
  if (!sorting.field) return []
  const columnId = SORT_FIELD_TO_COLUMN[sorting.field]
  if (!columnId) return []
  return [{ id: columnId, desc: sorting.direction === ListSandboxesOrderEnum.DESC }]
}

export function convertTableSortingToApiSorting(tableSorting: SortingState): SandboxSorting {
  if (tableSorting.length === 0) return {}
  const { id, desc } = tableSorting[0]
  const field = COLUMN_TO_SORT_FIELD[id]
  if (!field) return {}
  return {
    field,
    direction: desc ? ListSandboxesOrderEnum.DESC : ListSandboxesOrderEnum.ASC,
  }
}

export function convertApiFiltersToTableFilters(filters: SandboxFilters): ColumnFiltersState {
  const columnFilters: ColumnFiltersState = []

  if (filters.name) {
    columnFilters.push({ id: 'name', value: filters.name })
  }
  if (filters.states?.length) {
    columnFilters.push({ id: 'state', value: filters.states })
  }
  if (filters.snapshots?.length) {
    columnFilters.push({ id: 'snapshot', value: filters.snapshots })
  }
  if (filters.regions?.length) {
    columnFilters.push({ id: 'region', value: filters.regions })
  }
  if (filters.labels && Object.keys(filters.labels).length > 0) {
    const labelStrings = Object.entries(filters.labels).map(([key, value]) => `${key}: ${value}`)
    columnFilters.push({ id: 'labels', value: labelStrings })
  }
  if (filters.lastEventAfter || filters.lastEventBefore) {
    columnFilters.push({ id: 'lastEvent', value: [filters.lastEventAfter, filters.lastEventBefore] })
  }
  if (filters.createdAtAfter || filters.createdAtBefore) {
    columnFilters.push({ id: 'createdAt', value: [filters.createdAtAfter, filters.createdAtBefore] })
  }

  const resourceValue: ResourceFilterValue = {}
  if (filters.minCpu !== undefined || filters.maxCpu !== undefined) {
    resourceValue.cpu = { min: filters.minCpu, max: filters.maxCpu }
  }
  if (filters.minMemoryGiB !== undefined || filters.maxMemoryGiB !== undefined) {
    resourceValue.memory = { min: filters.minMemoryGiB, max: filters.maxMemoryGiB }
  }
  if (filters.minDiskGiB !== undefined || filters.maxDiskGiB !== undefined) {
    resourceValue.disk = { min: filters.minDiskGiB, max: filters.maxDiskGiB }
  }
  if (Object.keys(resourceValue).length > 0) {
    columnFilters.push({ id: 'resources', value: resourceValue })
  }

  if (filters.isPublic !== undefined) {
    columnFilters.push({ id: 'isPublic', value: filters.isPublic })
  }
  if (filters.isRecoverable !== undefined) {
    columnFilters.push({ id: 'isRecoverable', value: filters.isRecoverable })
  }

  return columnFilters
}

export function convertTableFiltersToApiFilters(tableFilters: ColumnFiltersState): SandboxFilters {
  const filters: SandboxFilters = {}

  for (const filter of tableFilters) {
    switch (filter.id) {
      case 'name':
        filters.name = filter.value as string
        break
      case 'state':
        filters.states = filter.value as SandboxFilters['states']
        break
      case 'snapshot':
        filters.snapshots = filter.value as string[]
        break
      case 'region':
        filters.regions = filter.value as string[]
        break
      case 'labels': {
        const labelStrings = filter.value as string[]
        const labels: Record<string, string> = {}
        for (const labelStr of labelStrings) {
          const separatorIndex = labelStr.indexOf(': ')
          if (separatorIndex !== -1) {
            labels[labelStr.substring(0, separatorIndex)] = labelStr.substring(separatorIndex + 2)
          }
        }
        if (Object.keys(labels).length > 0) {
          filters.labels = labels
        }
        break
      }
      case 'lastEvent': {
        const [after, before] = filter.value as (Date | undefined)[]
        filters.lastEventAfter = after
        filters.lastEventBefore = before
        break
      }
      case 'createdAt': {
        const [createdAfter, createdBefore] = filter.value as (Date | undefined)[]
        filters.createdAtAfter = createdAfter
        filters.createdAtBefore = createdBefore
        break
      }
      case 'resources': {
        const resourceValue = filter.value as ResourceFilterValue
        if (resourceValue.cpu) {
          filters.minCpu = resourceValue.cpu.min
          filters.maxCpu = resourceValue.cpu.max
        }
        if (resourceValue.memory) {
          filters.minMemoryGiB = resourceValue.memory.min
          filters.maxMemoryGiB = resourceValue.memory.max
        }
        if (resourceValue.disk) {
          filters.minDiskGiB = resourceValue.disk.min
          filters.maxDiskGiB = resourceValue.disk.max
        }
        break
      }
      case 'isPublic':
        filters.isPublic = filter.value as boolean
        break
      case 'isRecoverable':
        filters.isRecoverable = filter.value as boolean
        break
    }
  }

  return filters
}
