/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { PaginatedTraces } from '@daytonaio/api-client'

export interface TracesQueryParams {
  from: Date
  to: Date
  page?: number
  limit?: number
}

export function useSandboxTraces(
  sandboxId: string | undefined,
  params: TracesQueryParams,
  options?: Omit<UseQueryOptions<PaginatedTraces>, 'queryKey' | 'queryFn'>,
) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedTraces>({
    queryKey: queryKeys.telemetry.traces(sandboxId ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId) {
        throw new Error('Missing required parameters')
      }
      const response = await api.sandboxApi.getSandboxTraces(
        sandboxId,
        params.from,
        params.to,
        selectedOrganization.id,
        params.page,
        params.limit,
      )
      return response.data
    },
    enabled: !!sandboxId && !!selectedOrganization && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
