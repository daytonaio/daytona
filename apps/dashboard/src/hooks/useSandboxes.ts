/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  SandboxListSortDirection,
  SandboxListSortField,
  SandboxState,
  ListSandboxesResponse,
} from '@daytona/api-client'

export interface SandboxFilters {
  name?: string
  labels?: Record<string, string>
  includeErroredDeleted?: boolean
  states?: SandboxState[]
  snapshots?: string[]
  regions?: string[]
  minCpu?: number
  maxCpu?: number
  minMemoryGib?: number
  maxMemoryGib?: number
  minDiskGib?: number
  maxDiskGib?: number
  lastEventAfter?: Date
  lastEventBefore?: Date
  createdAtAfter?: Date
  createdAtBefore?: Date
  isPublic?: boolean
  isRecoverable?: boolean
}

export interface SandboxSorting {
  field?: SandboxListSortField
  direction?: SandboxListSortDirection
}

export const DEFAULT_SANDBOX_SORTING: SandboxSorting = {
  field: SandboxListSortField.LAST_ACTIVITY_AT,
  direction: SandboxListSortDirection.DESC,
}

export interface SandboxQueryParams {
  cursor?: string
  limit: number
  filters?: SandboxFilters
  sorting?: SandboxSorting
}

export const getSandboxesQueryKey = (organizationId: string | undefined, params?: SandboxQueryParams): QueryKey => {
  const baseKey = ['sandboxes' as const, organizationId]

  if (!params) {
    return baseKey
  }

  const normalizedParams = {
    cursor: params.cursor,
    limit: params.limit,
    ...(params.filters && { filters: params.filters }),
    ...(params.sorting && { sorting: params.sorting }),
  }

  return [...baseKey, normalizedParams]
}

export function useSandboxes(queryKey: QueryKey, params: SandboxQueryParams) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ListSandboxesResponse>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const { cursor, limit, filters = {}, sorting = {} } = params

      const listResponse = await sandboxApi.listSandboxes(
        selectedOrganization.id,
        cursor,
        limit,
        undefined,
        filters.name,
        filters.labels ? JSON.stringify(filters.labels) : undefined,
        filters.includeErroredDeleted,
        filters.states,
        filters.snapshots,
        filters.regions,
        filters.minCpu,
        filters.maxCpu,
        filters.minMemoryGib,
        filters.maxMemoryGib,
        filters.minDiskGib,
        filters.maxDiskGib,
        filters.isPublic,
        filters.isRecoverable,
        filters.createdAtAfter,
        filters.createdAtBefore,
        filters.lastEventAfter,
        filters.lastEventBefore,
        sorting.field,
        sorting.direction,
      )

      return listResponse.data
    },
    enabled: !!selectedOrganization,
    staleTime: 1000 * 10, // 10 seconds
    gcTime: 1000 * 60 * 5, // 5 minutes,
  })
}
