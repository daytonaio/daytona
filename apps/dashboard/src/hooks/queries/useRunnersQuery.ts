/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Runner } from '@daytona/api-client'
import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

type RunnersQueryOptions = Omit<UseQueryOptions<Runner[]>, 'queryKey' | 'queryFn' | 'enabled'> & {
  enabled?: boolean
  regionId?: string
}

export function useRunnersQuery(options?: RunnersQueryOptions) {
  const { runnersApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const { enabled = true, regionId, ...queryOptions } = options ?? {}
  const normalizedRegionId = regionId || undefined

  return useQuery<Runner[]>({
    queryKey: queryKeys.runners.list(selectedOrganization?.id ?? '', normalizedRegionId),
    meta: {
      errorMessage: 'Failed to fetch runners',
    },
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const response = await runnersApi.listRunners(normalizedRegionId, selectedOrganization.id)
      return response.data ?? []
    },
    enabled: enabled && !!selectedOrganization,
    staleTime: 1000 * 10,
    gcTime: 1000 * 60 * 5,
    ...queryOptions,
  })
}
