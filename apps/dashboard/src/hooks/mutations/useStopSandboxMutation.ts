/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'

interface StopSandboxVariables {
  sandboxId: string
}

interface UseStopSandboxMutationOptions {
  invalidate?: boolean
}

export const useStopSandboxMutation = ({ invalidate = true }: UseStopSandboxMutationOptions = {}) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  return useMutation({
    mutationKey: mutationKeys.sandboxes.stop(),
    mutationFn: async ({ sandboxId }: StopSandboxVariables) => {
      await sandboxApi.stopSandbox(sandboxId, selectedOrganization?.id)
    },
    onSuccess: (_, { sandboxId }) => {
      if (!invalidate) {
        return
      }

      queryClient.invalidateQueries({
        queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
      })
    },
  })
}
