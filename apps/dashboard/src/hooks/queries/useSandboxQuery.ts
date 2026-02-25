/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useQuery } from '@tanstack/react-query'
import { queryKeys } from './queryKeys'

export const useSandboxQuery = (sandboxId: string) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery({
    queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
    queryFn: async () => {
      const response = await sandboxApi.getSandbox(sandboxId, selectedOrganization?.id)
      return response.data
    },
    enabled: !!sandboxId && !!selectedOrganization?.id,
    staleTime: 1000 * 10,
  })
}
