/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { mutationKeys, SandboxMutationVariables } from './mutationKeys'
import { useMutation, useQueryClient } from '@tanstack/react-query'

interface SandboxMutationContext {
  organizationId?: string
}

export const useDeleteSandboxMutation = () => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, SandboxMutationVariables, SandboxMutationContext>({
    mutationKey: mutationKeys.sandboxes.delete,
    mutationFn: async ({ sandboxId }: SandboxMutationVariables) => {
      await sandboxApi.deleteSandbox(sandboxId, selectedOrganization?.id)
    },
    onMutate: () => ({
      organizationId: selectedOrganization?.id,
    }),
    onSuccess: async (_, { sandboxId }, context) => {
      if (!context?.organizationId) {
        return
      }

      await queryClient.invalidateQueries({
        queryKey: queryKeys.sandboxes.detail(context.organizationId, sandboxId),
      })
    },
  })
}
