/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, QueryKey } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

export const getSandboxLogsQueryKey = (organizationId: string | undefined, sandboxId: string): QueryKey => {
  return ['sandbox-logs' as const, organizationId, sandboxId]
}

export function useSandboxLogs(sandboxId: string) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<string>({
    queryKey: getSandboxLogsQueryKey(selectedOrganization?.id, sandboxId),
    queryFn: async () => {
      if (!selectedOrganization || !sandboxId) {
        throw new Error('No organization selected or sandbox ID missing')
      }

      const response = await sandboxApi.getSandboxLogs(sandboxId, selectedOrganization.id)
      return response.data
    },
    enabled: !!selectedOrganization && !!sandboxId,
    staleTime: 1000 * 5, // 5 seconds
    gcTime: 1000 * 30, // 30 seconds
    refetchInterval: 1000 * 10, // Refetch every 10 seconds for live logs
  })
}
