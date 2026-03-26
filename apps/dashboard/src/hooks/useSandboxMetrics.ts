/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
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
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<MetricsResponse>({
    queryKey: queryKeys.telemetry.metrics(sandboxId ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId || !api.analyticsTelemetryApi) {
        throw new Error('Missing required parameters')
      }
      const metricNames = params.metricNames?.length ? params.metricNames.join(',') : undefined

      const response = await api.analyticsTelemetryApi.organizationOrganizationIdSandboxSandboxIdTelemetryMetricsGet(
        selectedOrganization.id,
        sandboxId,
        params.from.toISOString(),
        params.to.toISOString(),
        metricNames,
      )

      // Group flat ModelsMetricPoint[] into MetricSeries[]
      const seriesMap = new Map<string, { timestamp: string; value: number }[]>()
      for (const point of response.data ?? []) {
        const name = point.metricName ?? ''
        if (!seriesMap.has(name)) {
          seriesMap.set(name, [])
        }
        seriesMap.get(name)!.push({
          timestamp: point.timestamp ?? '',
          value: point.value ?? 0,
        })
      }

      const series = Array.from(seriesMap.entries()).map(([metricName, dataPoints]) => ({
        metricName,
        dataPoints,
      }))

      return { series }
    },
    enabled: !!sandboxId && !!selectedOrganization && !!api.analyticsTelemetryApi && !!params.from && !!params.to,
    staleTime: 10_000,
    ...options,
  })
}
