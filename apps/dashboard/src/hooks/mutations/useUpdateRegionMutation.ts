/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { UpdateRegion } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface UpdateRegionMutationVariables {
  regionId: string
  region: UpdateRegion
  organizationId?: string
}

export const useUpdateRegionMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, UpdateRegionMutationVariables>({
    mutationKey: mutationKeys.regions.update(),
    mutationFn: async ({ regionId, region, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      await organizationsApi.updateRegion(regionId, region, organizationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.regions.available(organizationId) })
      }
    },
  })
}
