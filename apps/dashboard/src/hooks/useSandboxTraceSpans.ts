/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { TraceSpan } from '@daytonaio/api-client'

export function useSandboxTraceSpans(
  sandboxId: string | undefined,
  traceId: string | undefined,
  options?: Omit<UseQueryOptions<TraceSpan[]>, 'queryKey' | 'queryFn'>,
) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<TraceSpan[]>({
    queryKey: queryKeys.telemetry.traceSpans(sandboxId ?? '', traceId ?? ''),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId || !traceId) {
        throw new Error('Missing required parameters')
      }
      const response = await api.sandboxApi.getSandboxTraceSpans(sandboxId, traceId, selectedOrganization.id)
      return response.data
    },
    enabled: !!sandboxId && !!traceId && !!selectedOrganization,
    staleTime: 30_000,
    ...options,
  })
}
