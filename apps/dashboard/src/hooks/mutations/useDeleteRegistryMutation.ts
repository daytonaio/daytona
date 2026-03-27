/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface DeleteRegistryMutationVariables {
  registryId: string
  organizationId?: string
}

export const useDeleteRegistryMutation = () => {
  const { dockerRegistryApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeleteRegistryMutationVariables>({
    mutationFn: async ({ registryId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      await dockerRegistryApi.deleteRegistry(registryId, organizationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.registries.list(organizationId) })
      }
    },
  })
}
