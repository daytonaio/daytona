/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Region } from '@daytona/api-client'
import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useMemo } from 'react'
import { createRegionNameGetter, EMPTY_REGIONS } from '@/lib/regions'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

type RegionsQueryOptions = Omit<UseQueryOptions<Region[]>, 'queryKey' | 'queryFn' | 'enabled'> & {
  enabled?: boolean
}

export function useSharedRegionsQuery(options?: RegionsQueryOptions) {
  const { regionsApi } = useApi()
  const { enabled = true, ...queryOptions } = options ?? {}

  return useQuery<Region[]>({
    queryKey: queryKeys.regions.shared(),
    meta: {
      errorMessage: 'Failed to fetch shared regions',
    },
    queryFn: async () => {
      const response = await regionsApi.listSharedRegions()
      return response.data
    },
    enabled,
    ...queryOptions,
  })
}

export function useAvailableRegionsQuery(organizationId?: string, options?: RegionsQueryOptions) {
  const { organizationsApi } = useApi()
  const { enabled = true, ...queryOptions } = options ?? {}

  return useQuery<Region[]>({
    queryKey: queryKeys.regions.available(organizationId ?? ''),
    meta: {
      errorMessage: 'Failed to fetch available regions',
    },
    queryFn: async () => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.listAvailableRegions(organizationId)
      return response.data
    },
    enabled: enabled && !!organizationId,
    ...queryOptions,
  })
}

export function useRegionLookup(organizationId?: string) {
  const availableRegionsQuery = useAvailableRegionsQuery(organizationId)
  const sharedRegionsQuery = useSharedRegionsQuery()

  const availableRegions = availableRegionsQuery.data ?? EMPTY_REGIONS
  const sharedRegions = sharedRegionsQuery.data ?? EMPTY_REGIONS

  const getRegionName = useMemo(
    () => createRegionNameGetter(availableRegions, sharedRegions),
    [availableRegions, sharedRegions],
  )

  return useMemo(
    () => ({
      getRegionName,
      isLoading: availableRegionsQuery.isLoading || sharedRegionsQuery.isLoading,
      isFetching: availableRegionsQuery.isFetching || sharedRegionsQuery.isFetching,
    }),
    [
      availableRegionsQuery.isFetching,
      availableRegionsQuery.isLoading,
      getRegionName,
      sharedRegionsQuery.isFetching,
      sharedRegionsQuery.isLoading,
    ],
  )
}
