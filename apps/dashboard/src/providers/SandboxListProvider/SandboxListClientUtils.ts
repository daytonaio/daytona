/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import {
  ListSandboxesPaginatedOrderEnum,
  ListSandboxesPaginatedSortEnum,
  ListSandboxesPaginatedStatesEnum,
  Sandbox,
  SandboxState,
} from '@daytonaio/api-client'

const stateEnumToSandboxState: Record<ListSandboxesPaginatedStatesEnum, SandboxState> = {
  [ListSandboxesPaginatedStatesEnum.CREATING]: SandboxState.CREATING,
  [ListSandboxesPaginatedStatesEnum.RESTORING]: SandboxState.RESTORING,
  [ListSandboxesPaginatedStatesEnum.STARTING]: SandboxState.STARTING,
  [ListSandboxesPaginatedStatesEnum.STARTED]: SandboxState.STARTED,
  [ListSandboxesPaginatedStatesEnum.STOPPING]: SandboxState.STOPPING,
  [ListSandboxesPaginatedStatesEnum.STOPPED]: SandboxState.STOPPED,
  [ListSandboxesPaginatedStatesEnum.ARCHIVING]: SandboxState.ARCHIVING,
  [ListSandboxesPaginatedStatesEnum.ARCHIVED]: SandboxState.ARCHIVED,
  [ListSandboxesPaginatedStatesEnum.DESTROYING]: SandboxState.DESTROYING,
  [ListSandboxesPaginatedStatesEnum.ERROR]: SandboxState.ERROR,
  [ListSandboxesPaginatedStatesEnum.BUILD_FAILED]: SandboxState.BUILD_FAILED,
  [ListSandboxesPaginatedStatesEnum.PENDING_BUILD]: SandboxState.PENDING_BUILD,
  [ListSandboxesPaginatedStatesEnum.BUILDING_SNAPSHOT]: SandboxState.BUILDING_SNAPSHOT,
  [ListSandboxesPaginatedStatesEnum.UNKNOWN]: SandboxState.UNKNOWN,
  [ListSandboxesPaginatedStatesEnum.PULLING_SNAPSHOT]: SandboxState.PULLING_SNAPSHOT,
  [ListSandboxesPaginatedStatesEnum.RESIZING]: SandboxState.RESIZING,
}

export function matchesSandboxFilters(sandbox: Sandbox, filters: SandboxFilters): boolean {
  if (filters.idOrName) {
    const search = filters.idOrName.toLowerCase()
    const matchesId = sandbox.id.toLowerCase().includes(search)
    const matchesName = sandbox.name?.toLowerCase().includes(search)
    if (!matchesId && !matchesName) return false
  }

  if (filters.states && filters.states.length > 0) {
    const sandboxStates = filters.states.map((s) => stateEnumToSandboxState[s]).filter(Boolean)
    if (!sandbox.state || !sandboxStates.includes(sandbox.state)) return false
  }

  if (filters.snapshots && filters.snapshots.length > 0) {
    if (!sandbox.snapshot || !filters.snapshots.includes(sandbox.snapshot)) return false
  }

  if (filters.regions && filters.regions.length > 0) {
    if (!sandbox.target || !filters.regions.includes(sandbox.target)) return false
  }

  if (filters.labels) {
    const sandboxLabels = sandbox.labels || {}
    for (const [key, value] of Object.entries(filters.labels)) {
      if (sandboxLabels[key] !== value) return false
    }
  }

  if (filters.minCpu !== undefined && sandbox.cpu < filters.minCpu) return false
  if (filters.maxCpu !== undefined && sandbox.cpu > filters.maxCpu) return false
  if (filters.minMemoryGiB !== undefined && sandbox.memory < filters.minMemoryGiB) return false
  if (filters.maxMemoryGiB !== undefined && sandbox.memory > filters.maxMemoryGiB) return false
  if (filters.minDiskGiB !== undefined && sandbox.disk < filters.minDiskGiB) return false
  if (filters.maxDiskGiB !== undefined && sandbox.disk > filters.maxDiskGiB) return false

  if (filters.lastEventAfter || filters.lastEventBefore) {
    const lastEventTimestamp = sandbox.updatedAt ? new Date(sandbox.updatedAt).getTime() : Number.NaN
    if (Number.isNaN(lastEventTimestamp)) return false

    if (filters.lastEventAfter && lastEventTimestamp < filters.lastEventAfter.getTime()) return false
    if (filters.lastEventBefore && lastEventTimestamp > filters.lastEventBefore.getTime()) return false
  }

  return true
}

function getSortValue(sandbox: Sandbox, field?: ListSandboxesPaginatedSortEnum): string | number | undefined {
  switch (field) {
    case ListSandboxesPaginatedSortEnum.ID:
      return sandbox.id
    case ListSandboxesPaginatedSortEnum.NAME:
      return sandbox.name?.toLowerCase() ?? ''
    case ListSandboxesPaginatedSortEnum.STATE:
      return sandbox.state ?? ''
    case ListSandboxesPaginatedSortEnum.SNAPSHOT:
      return sandbox.snapshot ?? ''
    case ListSandboxesPaginatedSortEnum.REGION:
      return sandbox.target ?? ''
    case ListSandboxesPaginatedSortEnum.CREATED_AT:
      return sandbox.createdAt ? new Date(sandbox.createdAt).getTime() : 0
    case ListSandboxesPaginatedSortEnum.UPDATED_AT:
    default:
      return sandbox.updatedAt ? new Date(sandbox.updatedAt).getTime() : 0
  }
}

export function compareSandboxesBySorting(a: Sandbox, b: Sandbox, sorting: SandboxSorting): number {
  const aVal = getSortValue(a, sorting.field)
  const bVal = getSortValue(b, sorting.field)
  const direction = sorting.direction === ListSandboxesPaginatedOrderEnum.ASC ? 1 : -1

  if (aVal === undefined && bVal === undefined) return 0
  if (aVal === undefined) return direction
  if (bVal === undefined) return -direction

  if (typeof aVal === 'string' && typeof bVal === 'string') {
    return aVal.localeCompare(bVal) * direction
  }

  if (typeof aVal === 'number' && typeof bVal === 'number') {
    return (aVal - bVal) * direction
  }

  return 0
}
