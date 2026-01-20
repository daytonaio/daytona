/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  SearchSandboxesOrderEnum,
  SearchSandboxesSortEnum,
  SearchSandboxesStatesEnum,
  SearchSandboxesResult,
} from '@daytonaio/api-client'

export interface SandboxFilters {
  name?: string
  labels?: Record<string, string>
  includeErroredDeleted?: boolean
  states?: SearchSandboxesStatesEnum[]
  snapshots?: string[]
  regions?: string[]
  minCpu?: number
  maxCpu?: number
  minMemoryGiB?: number
  maxMemoryGiB?: number
  minDiskGiB?: number
  maxDiskGiB?: number
  lastEventAfter?: Date
  lastEventBefore?: Date
  createdAtAfter?: Date
  createdAtBefore?: Date
  isPublic?: boolean
  isRecoverable?: boolean
}

export interface SandboxSorting {
  field?: SearchSandboxesSortEnum
  direction?: SearchSandboxesOrderEnum
}

export const DEFAULT_SANDBOX_SORTING: SandboxSorting = {
  field: SearchSandboxesSortEnum.LAST_ACTIVITY_AT,
  direction: SearchSandboxesOrderEnum.DESC,
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

  return useQuery<SearchSandboxesResult>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const { cursor, limit, filters = {}, sorting = {} } = params

      const searchResponse = await sandboxApi.searchSandboxes(
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
        filters.minMemoryGiB,
        filters.maxMemoryGiB,
        filters.minDiskGiB,
        filters.maxDiskGiB,
        filters.isPublic,
        filters.isRecoverable,
        filters.createdAtAfter,
        filters.createdAtBefore,
        filters.lastEventAfter,
        filters.lastEventBefore,
        sorting.field,
        sorting.direction,
      )

      return searchResponse.data
    },
    enabled: !!selectedOrganization,
    staleTime: 1000 * 10, // 10 seconds
    gcTime: 1000 * 60 * 5, // 5 minutes,
  })
}
