/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from './queryKeys'

const VNC_PORT = 6080
const SESSION_DURATION_SECONDS = 300

export type VncSession = {
  url: string
  expiresAt: number
}

export const useVncSessionQuery = (sandboxId: string, enabled: boolean) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()
  const queryKey = queryKeys.sandboxes.vncSession(sandboxId)

  const query = useQuery({
    queryKey,
    queryFn: async (): Promise<VncSession> => {
      const url = (
        await sandboxApi.getSignedPortPreviewUrl(
          sandboxId,
          VNC_PORT,
          selectedOrganization?.id,
          SESSION_DURATION_SECONDS,
        )
      ).data.url
      return { url, expiresAt: Date.now() + SESSION_DURATION_SECONDS * 1000 }
    },
    enabled: enabled && !!sandboxId,
    staleTime: Infinity,
  })

  const existingSession = queryClient.getQueryData<VncSession>(queryKey)

  const reset = () => {
    queryClient.removeQueries({ queryKey })
  }

  return { ...query, existingSession, reset }
}
