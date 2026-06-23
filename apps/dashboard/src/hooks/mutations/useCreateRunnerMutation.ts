/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateRunner, CreateRunnerResponse } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface CreateRunnerMutationVariables {
  runner: CreateRunner
  organizationId?: string
}

export const useCreateRunnerMutation = () => {
  const { runnersApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<CreateRunnerResponse, unknown, CreateRunnerMutationVariables>({
    mutationKey: mutationKeys.runners.create(),
    mutationFn: async ({ runner, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await runnersApi.createRunner(runner, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.runners.list(organizationId) })
      }
    },
  })
}
