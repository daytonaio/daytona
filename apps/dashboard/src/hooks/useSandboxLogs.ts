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
      if (!selectedOrganization || !sandboxId || !api.analyticsTelemetryApi) {
        throw new Error('Missing required parameters')
      }
      const limit = params.limit ?? 50
      const page = params.page ?? 1
      const offset = (page - 1) * limit
      const severity = params.severities?.length ? params.severities.join(',') : undefined

      const response = await api.analyticsTelemetryApi.organizationOrganizationIdSandboxSandboxIdTelemetryLogsGet(
        selectedOrganization.id,
        sandboxId,
        params.from.toISOString(),
        params.to.toISOString(),
        severity,
        params.search,
        limit,
        offset,
      )

      const items = (response.data ?? []).map((entry) => ({
        timestamp: entry.timestamp ?? '',
        body: entry.body ?? '',
        severityText: entry.severityText ?? '',
        severityNumber: entry.severityNumber,
        serviceName: entry.serviceName ?? '',
        resourceAttributes: entry.resourceAttributes ?? {},
        logAttributes: entry.logAttributes ?? {},
        traceId: entry.traceId,
        spanId: entry.spanId,
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
