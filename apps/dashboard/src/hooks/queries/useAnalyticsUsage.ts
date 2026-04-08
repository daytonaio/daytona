/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import {
  ModelsAggregatedUsage,
  ModelsSandboxUsage,
  ModelsUsageChartPoint,
  ModelsUsagePeriod,
} from '@daytona/analytics-api-client'

export interface AnalyticsUsageParams {
  from: Date
  to: Date
  enabled?: boolean
}

export function useAggregatedUsage(params: AnalyticsUsageParams) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ModelsAggregatedUsage>({
    queryKey: queryKeys.analytics.aggregatedUsage(selectedOrganization?.id ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !api.analyticsUsageApi) {
        throw new Error('Missing required parameters')
      }
      const response = await api.analyticsUsageApi.organizationOrganizationIdUsageAggregatedGet(
        selectedOrganization.id,
        params.from.toISOString(),
        params.to.toISOString(),
      )
      return response.data
    },
    enabled: !!selectedOrganization && !!api.analyticsUsageApi && params.enabled !== false,
    staleTime: 10_000,
  })
}

export function useSandboxesUsage(params: AnalyticsUsageParams) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ModelsSandboxUsage[]>({
    queryKey: queryKeys.analytics.sandboxesUsage(selectedOrganization?.id ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !api.analyticsUsageApi) {
        throw new Error('Missing required parameters')
      }
      const response = await api.analyticsUsageApi.organizationOrganizationIdUsageSandboxGet(
        selectedOrganization.id,
        params.from.toISOString(),
        params.to.toISOString(),
      )
      return response.data
    },
    enabled: !!selectedOrganization && !!api.analyticsUsageApi && params.enabled !== false,
    staleTime: 10_000,
  })
}

export interface UsageChartParams extends AnalyticsUsageParams {
  region?: string
}

export function useUsageChart(params: UsageChartParams) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ModelsUsageChartPoint[]>({
    queryKey: queryKeys.analytics.usageChart(selectedOrganization?.id ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !api.analyticsUsageApi) {
        throw new Error('Missing required parameters')
      }
      const response = await api.analyticsUsageApi.organizationOrganizationIdUsageChartGet(
        selectedOrganization.id,
        params.from.toISOString(),
        params.to.toISOString(),
        params.region,
      )
      return response.data
    },
    enabled: !!selectedOrganization && !!api.analyticsUsageApi && params.enabled !== false,
    staleTime: 10_000,
  })
}

export function useSandboxUsagePeriods(sandboxId: string | undefined, params: AnalyticsUsageParams) {
  const api = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<ModelsUsagePeriod[]>({
    queryKey: queryKeys.analytics.sandboxUsagePeriods(selectedOrganization?.id ?? '', sandboxId ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId || !api.analyticsUsageApi) {
        throw new Error('Missing required parameters')
      }
      const response = await api.analyticsUsageApi.organizationOrganizationIdSandboxSandboxIdUsageGet(
        selectedOrganization.id,
        sandboxId,
        params.from.toISOString(),
        params.to.toISOString(),
      )
      return response.data
    },
    enabled: !!sandboxId && !!selectedOrganization && !!api.analyticsUsageApi && params.enabled !== false,
    staleTime: 10_000,
  })
}
