/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
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
  return useQuery<PaginatedTraces>({
    queryKey: queryKeys.telemetry.traces(sandboxId ?? '', params),
    queryFn: async () => {
      return { items: [], total: 0, page: params.page ?? 1, totalPages: 0 }
    },
    enabled: !!sandboxId && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
