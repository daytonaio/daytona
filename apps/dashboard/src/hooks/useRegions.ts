/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { QueryKey, useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Region } from '@daytonaio/api-client'

export const getRegionsQueryKey = (organizationId: string | undefined): QueryKey => {
  return ['regions' as const, organizationId]
}

export function useRegions(queryKey: QueryKey) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<Region[]>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const response = await sandboxApi.getSandboxRegions(selectedOrganization.id)

      return response.data
    },
    enabled: !!selectedOrganization,
  })
}
