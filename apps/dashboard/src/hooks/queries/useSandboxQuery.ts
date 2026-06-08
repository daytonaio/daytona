/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useQuery } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { queryKeys } from './queryKeys'

export function getSandboxQueryErrorStatus(error: unknown) {
  const cause = error instanceof Error ? error.cause : undefined

  if (!isAxiosError(cause)) {
    return undefined
  }

  return cause.response?.status ?? cause.status
}

export const useSandboxQuery = (sandboxId: string, { enabled = true }: { enabled?: boolean } = {}) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery({
    queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
    queryFn: async () => {
      const response = await sandboxApi.getSandbox(sandboxId, selectedOrganization?.id)
      return response.data
    },
    enabled: enabled && !!sandboxId && !!selectedOrganization?.id,
    staleTime: 1000 * 10,
    retry: (failureCount, error) => {
      if (getSandboxQueryErrorStatus(error) === 404) return false
      return failureCount < 3
    },
  })
}
