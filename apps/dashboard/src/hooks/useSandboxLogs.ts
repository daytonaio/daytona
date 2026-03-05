/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
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
  return useQuery<PaginatedLogs>({
    queryKey: queryKeys.telemetry.logs(sandboxId ?? '', params),
    queryFn: async () => {
      return { items: [], total: 0, page: params.page ?? 1, totalPages: 0 }
    },
    enabled: !!sandboxId && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
