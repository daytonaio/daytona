/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { mutationKeys, SandboxMutationVariables } from './mutationKeys'
import { useMutation, useQueryClient } from '@tanstack/react-query'

export const useRecoverSandboxMutation = () => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  return useMutation({
    mutationKey: mutationKeys.sandboxes.recover,
    mutationFn: async ({ sandboxId }: SandboxMutationVariables) => {
      await sandboxApi.recoverSandbox(sandboxId, selectedOrganization?.id)
    },
    onSuccess: async (_, { sandboxId }) => {
      await queryClient.invalidateQueries({
        queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
      })
    },
  })
}
