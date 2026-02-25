/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useQuery } from '@tanstack/react-query'
import { queryKeys } from './queryKeys'

export const useVncInitialStatusQuery = (sandboxId: string, enabled: boolean) => {
  const { toolboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery({
    queryKey: queryKeys.sandboxes.vncInitialStatus(sandboxId),
    queryFn: async () => {
      const { data } = await toolboxApi.getComputerUseStatusDeprecated(sandboxId, selectedOrganization?.id)
      return data.status as string
    },
    enabled: enabled && !!sandboxId,
    retry: false,
    staleTime: 0,
  })
}

export const useVncPollStatusQuery = (sandboxId: string, enabled: boolean) => {
  const { toolboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery({
    queryKey: queryKeys.sandboxes.vncPollStatus(sandboxId),
    queryFn: async () => {
      const { data } = await toolboxApi.getComputerUseStatusDeprecated(sandboxId, selectedOrganization?.id)
      if (data.status !== 'active') throw new Error(`VNC not ready: ${data.status}`)
      return data.status as string
    },
    enabled: enabled && !!sandboxId,
    retry: 10,
    retryDelay: 2000,
  })
}
