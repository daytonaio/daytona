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
      if (!selectedOrganization || !sandboxId || !api.analyticsTelemetryApi) {
        throw new Error('Missing required parameters')
      }
      const limit = params.limit ?? 50
      const page = params.page ?? 1
      const offset = (page - 1) * limit

      const response = await api.analyticsTelemetryApi.organizationOrganizationIdSandboxSandboxIdTelemetryTracesGet(
        selectedOrganization.id,
        sandboxId,
        params.from.toISOString(),
        params.to.toISOString(),
        limit,
        offset,
      )

      const items = (response.data ?? []).map((trace) => ({
        traceId: trace.traceId ?? '',
        rootSpanName: trace.rootSpanName ?? '',
        startTime: trace.startTime ?? '',
        endTime: trace.endTime ?? '',
        durationMs: trace.totalDurationMs ?? 0,
        spanCount: trace.spanCount ?? 0,
        statusCode: trace.statusCode,
      }))

      return {
        items,
        total: items.length < limit ? offset + items.length : offset + items.length + 1,
        page,
        totalPages: items.length < limit ? page : page + 1,
      }
    },
    enabled: !!sandboxId && !!selectedOrganization && !!api.analyticsTelemetryApi && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
