/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  ListSandboxesPaginatedOrderEnum,
  ListSandboxesPaginatedSortEnum,
  ListSandboxesPaginatedStatesEnum,
  PaginatedSandboxes,
} from '@daytonaio/api-client'

export interface SandboxFilters {
  id?: string
  name?: string
  labels?: Record<string, string>
  includeErroredDeleted?: boolean
  states?: ListSandboxesPaginatedStatesEnum[]
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
}

export interface SandboxSorting {
  field?: ListSandboxesPaginatedSortEnum
  direction?: ListSandboxesPaginatedOrderEnum
}

export const DEFAULT_SANDBOX_SORTING: SandboxSorting = {
  field: ListSandboxesPaginatedSortEnum.UPDATED_AT,
  direction: ListSandboxesPaginatedOrderEnum.DESC,
}

export interface SandboxQueryParams {
  page: number
  pageSize: number
  filters?: SandboxFilters
  sorting?: SandboxSorting
}

export const getSandboxesQueryKey = (organizationId: string | undefined, params?: SandboxQueryParams): QueryKey => {
  const baseKey = ['sandboxes' as const, organizationId]

  if (!params) {
    return baseKey
  }

  const normalizedParams = {
    page: params.page,
    pageSize: params.pageSize,
    ...(params.filters && { filters: params.filters }),
    ...(params.sorting && { sorting: params.sorting }),
  }

  return [...baseKey, normalizedParams]
}

export function useSandboxes(queryKey: QueryKey, params: SandboxQueryParams) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedSandboxes>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const { page, pageSize, filters = {}, sorting = {} } = params

      const response = await sandboxApi.listSandboxesPaginated(
        selectedOrganization.id,
        page,
        pageSize,
        filters.id,
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
        filters.lastEventAfter,
        filters.lastEventBefore,
        sorting.field,
        sorting.direction,
      )

      return response.data
    },
    enabled: !!selectedOrganization,
    staleTime: 1000 * 30, // 30 seconds
    gcTime: 1000 * 60 * 5, // 5 minutes,
  })
}
