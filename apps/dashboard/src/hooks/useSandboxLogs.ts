/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { PaginatedLogs } from '@daytonaio/api-client'

export interface LogsQueryParams {
  from: Date
  to: Date
  page?: number
  limit?: number
  severities?: string[]
  search?: string
}

export function useSandboxLogs(
  sandboxId: string | undefined,
  params: LogsQueryParams,
  options?: Omit<UseQueryOptions<PaginatedLogs>, 'queryKey' | 'queryFn'>,
) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedLogs>({
    queryKey: queryKeys.telemetry.logs(sandboxId ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId) {
        throw new Error('Missing required parameters')
      }
      const response = await api.sandboxApi.getSandboxLogs(
        sandboxId,
        params.from,
        params.to,
        selectedOrganization.id,
        params.page,
        params.limit,
        params.severities,
        params.search,
      )
      return response.data
    },
    enabled: !!sandboxId && !!selectedOrganization && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
