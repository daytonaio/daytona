/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { useMutation, useQueryClient } from '@tanstack/react-query'

interface StartSandboxVariables {
  sandboxId: string
}

export const useStartSandboxMutation = () => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ sandboxId }: StartSandboxVariables) => {
      await sandboxApi.startSandbox(sandboxId, selectedOrganization?.id)
    },
    onSuccess: (_, { sandboxId }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
      })
    },
  })
}
