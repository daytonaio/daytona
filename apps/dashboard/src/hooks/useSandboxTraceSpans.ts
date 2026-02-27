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
      if (!selectedOrganization || !sandboxId || !traceId || !api.analyticsTelemetryApi) {
        throw new Error('Missing required parameters')
      }
      const response =
        await api.analyticsTelemetryApi.organizationOrganizationIdSandboxSandboxIdTelemetryTracesTraceIdGet(
          selectedOrganization.id,
          sandboxId,
          traceId,
        )

      return (response.data ?? []).map((span) => ({
        traceId: span.traceId ?? '',
        spanId: span.spanId ?? '',
        parentSpanId: span.parentSpanId,
        spanName: span.spanName ?? '',
        timestamp: span.timestamp ?? '',
        durationNs: (span.durationMs ?? 0) * 1_000_000,
        spanAttributes: span.spanAttributes ?? {},
        statusCode: span.statusCode,
        statusMessage: span.statusMessage,
      }))
    },
    enabled: !!sandboxId && !!traceId && !!selectedOrganization && !!api.analyticsTelemetryApi,
    staleTime: 30_000,
    ...options,
  })
}
