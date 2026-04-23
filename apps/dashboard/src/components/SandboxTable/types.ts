/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/queries/useSandboxesQuery'
import {
  Region,
  SandboxClass,
  SandboxListItem,
  SandboxListSortDirection,
  SandboxListSortField,
  SandboxState,
  SnapshotDto,
} from '@daytona/api-client'
import { ColumnFiltersState, SortingState, Table } from '@tanstack/react-table'
import type { Ref } from 'react'
import { ResourceFilterValue } from './filters/ResourceFilter'

export interface SandboxTableRef {
  table: Table<SandboxListItem>
}

export interface SandboxTableProps {
  ref?: Ref<SandboxTableRef>
  data: SandboxListItem[]
  sandboxIsLoading: Record<string, boolean>
  activeSandboxId?: string
  loading: boolean
  isShowingPreviousData?: boolean
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
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  onRowClick?: (sandbox: SandboxListItem) => void
  handleRecover: (id: string) => void
  handleScreenRecordings: (id: string) => void
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
  handlePause: (id: string) => void
  handleRefresh: () => void
  isRefreshing?: boolean
  sorting: SandboxSorting
  onSortingChange: (sorting: SandboxSorting) => void
  filters: SandboxFilters
  onFiltersChange: (filters: SandboxFilters) => void
  handleOpenTerminal: (sandbox: SandboxListItem) => void
}

export interface SandboxTableActionsProps {
  sandbox: SandboxListItem
  writePermitted: boolean
  deletePermitted: boolean
  isLoading: boolean
  onStart: (id: string) => void
  onStop: (id: string) => void
  onDelete: (id: string) => void
  onArchive: (id: string) => void
  onVnc: (id: string) => void
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
  onFork?: () => void
  onCreateSnapshot?: () => void
  onViewForks?: () => void
  onOpenTerminal?: () => void
  onPause: (id: string) => void
  onRecover: (id: string) => void
  onScreenRecordings: (id: string) => void
}

export interface SandboxTableHeaderProps {
  table: Table<SandboxListItem>
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
  [SandboxListSortField.NAME]: 'name',
  [SandboxListSortField.LAST_ACTIVITY_AT]: 'lastEvent',
  [SandboxListSortField.CREATED_AT]: 'createdAt',
  [SandboxListSortField.CPU]: 'cpu',
  [SandboxListSortField.MEMORY_GIB]: 'memory',
  [SandboxListSortField.DISK_GIB]: 'disk',
}

const COLUMN_TO_SORT_FIELD: Record<string, SandboxListSortField> = {
  name: SandboxListSortField.NAME,
  lastEvent: SandboxListSortField.LAST_ACTIVITY_AT,
  createdAt: SandboxListSortField.CREATED_AT,
}

export function convertApiSortingToTableSorting(sorting: SandboxSorting): SortingState {
  if (!sorting.field) return []
  const columnId = SORT_FIELD_TO_COLUMN[sorting.field]
  if (!columnId) return []
  return [{ id: columnId, desc: sorting.direction === SandboxListSortDirection.DESC }]
}

export function convertTableSortingToApiSorting(tableSorting: SortingState): SandboxSorting {
  if (tableSorting.length === 0) return DEFAULT_SANDBOX_SORTING
  const { id, desc } = tableSorting[0]
  const field = COLUMN_TO_SORT_FIELD[id]
  if (!field) return DEFAULT_SANDBOX_SORTING
  return {
    field,
    direction: desc ? SandboxListSortDirection.DESC : SandboxListSortDirection.ASC,
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
  if (filters.sandboxClasses?.length) {
    columnFilters.push({ id: 'sandboxClass', value: filters.sandboxClasses })
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
  if (filters.minMemoryGib !== undefined || filters.maxMemoryGib !== undefined) {
    resourceValue.memory = { min: filters.minMemoryGib, max: filters.maxMemoryGib }
  }
  if (filters.minDiskGib !== undefined || filters.maxDiskGib !== undefined) {
    resourceValue.disk = { min: filters.minDiskGib, max: filters.maxDiskGib }
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
        if (typeof filter.value === 'string' && filter.value.trim()) {
          filters.name = filter.value
        }
        break
      case 'state':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.states = filter.value as SandboxFilters['states']
        }
        break
      case 'snapshot':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.snapshots = filter.value as string[]
        }
        break
      case 'region':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.regions = filter.value as string[]
        }
        break
      case 'sandboxClass':
        if (Array.isArray(filter.value) && filter.value.length > 0) {
          filters.sandboxClasses = filter.value as SandboxClass[]
        }
        break
      case 'labels': {
        if (!Array.isArray(filter.value) || filter.value.length === 0) {
          break
        }
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
        if (!Array.isArray(filter.value)) {
          break
        }
        const [after, before] = filter.value as (Date | undefined)[]
        filters.lastEventAfter = after
        filters.lastEventBefore = before
        break
      }
      case 'createdAt': {
        if (!Array.isArray(filter.value)) {
          break
        }
        const [createdAfter, createdBefore] = filter.value as (Date | undefined)[]
        filters.createdAtAfter = createdAfter
        filters.createdAtBefore = createdBefore
        break
      }
      case 'resources': {
        if (!filter.value) {
          break
        }
        const resourceValue = filter.value as ResourceFilterValue
        if (resourceValue.cpu) {
          filters.minCpu = resourceValue.cpu.min
          filters.maxCpu = resourceValue.cpu.max
        }
        if (resourceValue.memory) {
          filters.minMemoryGib = resourceValue.memory.min
          filters.maxMemoryGib = resourceValue.memory.max
        }
        if (resourceValue.disk) {
          filters.minDiskGib = resourceValue.disk.min
          filters.maxDiskGib = resourceValue.disk.max
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
