/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { TraceSpan } from '@daytonaio/api-client'

export function useSandboxTraceSpans(
  sandboxId: string | undefined,
  traceId: string | undefined,
  options?: Omit<UseQueryOptions<TraceSpan[]>, 'queryKey' | 'queryFn'>,
) {
  return useQuery<TraceSpan[]>({
    queryKey: queryKeys.telemetry.traceSpans(sandboxId ?? '', traceId ?? ''),
    queryFn: async () => {
      return []
    },
    enabled: !!sandboxId && !!traceId,
    staleTime: 30_000,
    ...options,
  })
}
