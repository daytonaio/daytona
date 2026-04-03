/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateDockerRegistry, DockerRegistry } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CreateRegistryMutationVariables {
  registry: CreateDockerRegistry
  organizationId?: string
}

export const useCreateRegistryMutation = () => {
  const { dockerRegistryApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<DockerRegistry, unknown, CreateRegistryMutationVariables>({
    mutationFn: async ({ registry, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      const response = await dockerRegistryApi.createRegistry(registry, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.registries.list(organizationId) })
      }
    },
  })
}
