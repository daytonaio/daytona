/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Runner } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface UpdateRunnerSchedulingMutationVariables {
  runnerId: string
  organizationId?: string
  unschedulable: boolean
}

interface UseUpdateRunnerSchedulingMutationOptions {
  invalidateOnSuccess?: boolean
}

export const useUpdateRunnerSchedulingMutation = ({
  invalidateOnSuccess = true,
}: UseUpdateRunnerSchedulingMutationOptions = {}) => {
  const { runnersApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<Runner, unknown, UpdateRunnerSchedulingMutationVariables>({
    mutationKey: mutationKeys.runners.updateScheduling(),
    mutationFn: async ({ runnerId, organizationId, unschedulable }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await runnersApi.updateRunnerScheduling(runnerId, organizationId, {
        data: { unschedulable },
      })
      return response.data
    },
    onSuccess: async (runner, { organizationId }) => {
      if (!organizationId) {
        return
      }

      const queryKey = queryKeys.runners.list(organizationId)
      queryClient.setQueriesData<Runner[]>({ queryKey }, (previousRunners) => {
        if (!previousRunners) return previousRunners

        return previousRunners.map((previousRunner) => (previousRunner.id === runner.id ? runner : previousRunner))
      })

      if (invalidateOnSuccess) {
        await queryClient.invalidateQueries({ queryKey })
      }
    },
  })
}
