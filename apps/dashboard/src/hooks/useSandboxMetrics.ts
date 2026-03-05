/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
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
  return useQuery<MetricsResponse>({
    queryKey: queryKeys.telemetry.metrics(sandboxId ?? '', params),
    queryFn: async () => {
      return { series: [] }
    },
    enabled: !!sandboxId && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
