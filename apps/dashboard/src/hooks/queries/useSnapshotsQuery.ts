/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GetAllSnapshotsOrderEnum, GetAllSnapshotsSortEnum, PaginatedSnapshots, SnapshotDto } from '@daytona/api-client'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
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

export function getSnapshotQueryErrorStatus(error: unknown) {
  const cause = error instanceof Error ? error.cause : undefined

  if (!isAxiosError(cause)) {
    return undefined
  }

  return cause.response?.status ?? cause.status
}

export function useSnapshotQuery(
  snapshotId: string | null | undefined,
  { enabled = true }: { enabled?: boolean } = {},
) {
  const { snapshotApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<SnapshotDto>({
    queryKey: queryKeys.snapshots.detail(selectedOrganization?.id ?? '', snapshotId ?? ''),
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      if (!snapshotId) {
        throw new Error('No snapshot selected')
      }

      const response = await snapshotApi.getSnapshot(snapshotId, selectedOrganization.id)
      return response.data
    },
    enabled: enabled && !!snapshotId && !!selectedOrganization,
    staleTime: 1000 * 10,
    retry: (failureCount, error) => {
      const status = getSnapshotQueryErrorStatus(error)

      if (status === 404) {
        return failureCount < 1
      }

      if (status && status >= 400 && status < 500) {
        return false
      }

      return failureCount < 3
    },
  })
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
