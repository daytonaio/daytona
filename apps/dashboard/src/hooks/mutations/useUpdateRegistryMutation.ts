/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DockerRegistry, UpdateDockerRegistry } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface UpdateRegistryMutationVariables {
  registryId: string
  registry: UpdateDockerRegistry
  organizationId?: string
}

export const useUpdateRegistryMutation = () => {
  const { dockerRegistryApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<DockerRegistry, unknown, UpdateRegistryMutationVariables>({
    mutationFn: async ({ registryId, registry, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      const response = await dockerRegistryApi.updateRegistry(registryId, registry, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.registries.list(organizationId) })
      }
    },
  })
}
