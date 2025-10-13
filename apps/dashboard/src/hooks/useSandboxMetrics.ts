/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { MetricsResponse } from '@daytonaio/api-client'

export interface MetricsQueryParams {
  from: Date
  to: Date
  metricNames?: string[]
}

export function useSandboxMetrics(
  sandboxId: string | undefined,
  params: MetricsQueryParams,
  options?: Omit<UseQueryOptions<MetricsResponse>, 'queryKey' | 'queryFn'>,
) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<MetricsResponse>({
    queryKey: queryKeys.telemetry.metrics(sandboxId ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId) {
        throw new Error('Missing required parameters')
      }
      const response = await api.sandboxApi.getSandboxMetrics(
        sandboxId,
        params.from,
        params.to,
        selectedOrganization.id,
        params.metricNames,
      )
      return response.data
    },
    enabled: !!sandboxId && !!selectedOrganization && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
