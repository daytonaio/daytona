/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface DeleteRegionMutationVariables {
  regionId: string
  organizationId?: string
}

export const useDeleteRegionMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeleteRegionMutationVariables>({
    mutationKey: mutationKeys.regions.remove(),
    mutationFn: async ({ regionId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      await organizationsApi.deleteRegion(regionId, organizationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.regions.available(organizationId) })
      }
    },
  })
}
