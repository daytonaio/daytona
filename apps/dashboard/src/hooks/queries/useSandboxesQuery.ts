/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { QueryKey } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  type SandboxListItem,
  SandboxClass,
  SandboxListSortDirection,
  SandboxListSortField,
  SandboxState,
  ListSandboxesResponse,
} from '@daytona/api-client'
import { queryKeys } from './queryKeys'

type ListSandboxesQueryResponse = ListSandboxesResponse | SandboxListItem[]

export interface SandboxFilters {
  name?: string
  labels?: Record<string, string>
  includeErroredDeleted?: boolean
  states?: SandboxState[]
  snapshots?: string[]
  regions?: string[]
  sandboxClasses?: SandboxClass[]
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
  return queryKeys.sandboxes.list(organizationId ?? '', params)
}

function normalizeListSandboxesResponse(data: ListSandboxesQueryResponse): ListSandboxesResponse {
  if (Array.isArray(data)) {
    return {
      items: data,
      nextCursor: null,
    }
  }

  return {
    items: data.items ?? [],
    nextCursor: data.nextCursor ?? null,
  }
}

export function useSandboxesQuery(params: SandboxQueryParams) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ListSandboxesResponse>({
    queryKey: queryKeys.sandboxes.list(selectedOrganization?.id ?? '', params),
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
        filters.sandboxClasses,
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

      return normalizeListSandboxesResponse(listResponse.data)
    },
    enabled: !!selectedOrganization,
    placeholderData: keepPreviousData,
    staleTime: 1000 * 10, // 10 seconds
    gcTime: 1000 * 60 * 5, // 5 minutes,
  })
}
