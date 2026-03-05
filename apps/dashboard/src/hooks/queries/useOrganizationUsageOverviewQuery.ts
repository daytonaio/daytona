/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { queryKeys } from './queryKeys'
import { OrganizationUsageOverview } from '@daytonaio/api-client'

interface UsageOverviewParams {
  organizationId: string
}

export function useOrganizationUsageOverviewQuery(
  params: UsageOverviewParams,
  options?: Omit<UseQueryOptions<OrganizationUsageOverview>, 'queryKey' | 'queryFn'>,
) {
  const api = useApi()

  return useQuery<OrganizationUsageOverview>({
    queryKey: queryKeys.organization.usage.overview(params.organizationId),
    queryFn: async () => {
      const response = await api.organizationsApi.getOrganizationUsageOverview(params.organizationId)
      return response.data
    },
    enabled: !!params.organizationId,
    ...options,
  })
}
