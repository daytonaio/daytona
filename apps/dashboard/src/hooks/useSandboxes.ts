/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  ListSandboxesPaginatedDeprecatedOrderEnum,
  ListSandboxesPaginatedDeprecatedSortEnum,
  ListSandboxesPaginatedDeprecatedStatesEnum,
  PaginatedSandboxes,
} from '@daytonaio/api-client'
import { isValidUUID } from '@/lib/utils'

export interface SandboxFilters {
  idOrName?: string
  labels?: Record<string, string>
  includeErroredDeleted?: boolean
  states?: ListSandboxesPaginatedDeprecatedStatesEnum[]
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
  field?: ListSandboxesPaginatedDeprecatedSortEnum
  direction?: ListSandboxesPaginatedDeprecatedOrderEnum
}

export const DEFAULT_SANDBOX_SORTING: SandboxSorting = {
  field: ListSandboxesPaginatedDeprecatedSortEnum.UPDATED_AT,
  direction: ListSandboxesPaginatedDeprecatedOrderEnum.DESC,
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

      const listResponse = await sandboxApi.listSandboxesPaginatedDeprecated(
        selectedOrganization.id,
        page,
        pageSize,
        undefined,
        filters.idOrName,
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

      let paginatedData = listResponse.data

      // TODO: this will be obsolete once we introduce the search API
      if (filters.idOrName && isValidUUID(filters.idOrName) && page === 1) {
        // Attempt to fetch sandbox by ID if the search value is a valid UUID
        try {
          const sandbox = (await sandboxApi.getSandbox(filters.idOrName, selectedOrganization.id)).data
          const existsInPaginatedData = paginatedData.items.some((item) => item.id === sandbox.id)

          if (!existsInPaginatedData) {
            paginatedData = {
              ...paginatedData,
              // This is an exact UUID match, ignore sorting
              items: [sandbox, ...paginatedData.items],
              total: paginatedData.total + 1,
            }
          }
        } catch (error) {
          // TODO: rethrow if not 4xx
        }
      }

      return paginatedData
    },
    enabled: !!selectedOrganization,
    staleTime: 1000 * 10, // 10 seconds
    gcTime: 1000 * 60 * 5, // 5 minutes,
  })
}
