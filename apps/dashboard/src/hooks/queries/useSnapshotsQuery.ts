/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GetAllSnapshotsOrderEnum, GetAllSnapshotsSortEnum, PaginatedSnapshots } from '@daytonaio/api-client'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

export interface SnapshotFilters {
  name?: string
}

export interface SnapshotSorting {
  field: GetAllSnapshotsSortEnum
  direction: GetAllSnapshotsOrderEnum
}

export const DEFAULT_SNAPSHOT_SORTING: SnapshotSorting = {
  field: GetAllSnapshotsSortEnum.LAST_USED_AT,
  direction: GetAllSnapshotsOrderEnum.DESC,
}

export interface SnapshotQueryParams {
  page: number
  pageSize: number
  filters?: SnapshotFilters
  sorting?: SnapshotSorting
}

export function useSnapshotsQuery(params: SnapshotQueryParams) {
  const { snapshotApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedSnapshots>({
    queryKey: queryKeys.snapshots.list(selectedOrganization?.id ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const { page, pageSize, filters = {}, sorting = DEFAULT_SNAPSHOT_SORTING } = params

      const response = await snapshotApi.getAllSnapshots(
        selectedOrganization.id,
        page,
        pageSize,
        filters.name,
        sorting.field,
        sorting.direction,
      )

      return response.data
    },
    enabled: !!selectedOrganization,
    placeholderData: keepPreviousData,
    staleTime: 1000 * 10,
    gcTime: 1000 * 60 * 5,
  })
}
