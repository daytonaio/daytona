/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from './queryKeys'

const TERMINAL_PORT = 22222
const SESSION_DURATION_SECONDS = 300

export type TerminalSession = {
  url: string
  expiresAt: number
}

export const useTerminalSessionQuery = (sandboxId: string, enabled: boolean) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()
  const queryKey = queryKeys.sandboxes.terminalSession(sandboxId)

  const query = useQuery({
    queryKey,
    queryFn: async (): Promise<TerminalSession> => {
      const url = (
        await sandboxApi.getSignedPortPreviewUrl(
          sandboxId,
          TERMINAL_PORT,
          selectedOrganization?.id,
          SESSION_DURATION_SECONDS,
        )
      ).data.url
      return { url, expiresAt: Date.now() + SESSION_DURATION_SECONDS * 1000 }
    },
    enabled: enabled && !!sandboxId,
    staleTime: Infinity,
  })

  const existingSession = queryClient.getQueryData<TerminalSession>(queryKey)

  const reset = () => {
    queryClient.removeQueries({ queryKey })
  }

  return { ...query, existingSession, reset }
}
