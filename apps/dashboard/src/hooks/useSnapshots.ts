/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { PaginatedSnapshots } from '@daytonaio/api-client'

export interface SnapshotFilters {
  name?: string
}

export interface SnapshotQueryParams {
  page: number
  pageSize: number
  filters?: SnapshotFilters
}

export const getSnapshotsQueryKey = (organizationId: string | undefined, params?: SnapshotQueryParams): QueryKey => {
  const baseKey = ['snapshots' as const, organizationId]

  if (!params) {
    return baseKey
  }

  const normalizedParams = {
    page: params.page,
    pageSize: params.pageSize,
    ...(params.filters && { filters: params.filters }),
  }

  return [...baseKey, normalizedParams]
}

export function useSnapshots(queryKey: QueryKey, params: SnapshotQueryParams) {
  const { snapshotApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedSnapshots>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const { page, pageSize, filters = {} } = params

      const response = await snapshotApi.getAllSnapshots(selectedOrganization.id, page, pageSize, filters.name)

      return response.data
    },
    enabled: !!selectedOrganization,
  })
}
