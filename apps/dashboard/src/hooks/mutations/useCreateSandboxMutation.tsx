/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateSandbox, Sandbox } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { getSandboxesQueryKey } from '../useSandboxes'
import { useApi } from '../useApi'

export interface CreateSandboxMutationVariables {
  sandbox: CreateSandbox
  organizationId?: string
}

export const useCreateSandboxMutation = () => {
  const { sandboxApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<Sandbox, unknown, CreateSandboxMutationVariables>({
    mutationFn: async ({ sandbox, organizationId }) => {
      const response = await sandboxApi.createSandbox(sandbox, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: getSandboxesQueryKey(organizationId) })
      }
    },
  })
}
