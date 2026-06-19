/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateRegion, CreateRegionResponse } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface CreateRegionMutationVariables {
  region: CreateRegion
  organizationId?: string
}

export const useCreateRegionMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<CreateRegionResponse, unknown, CreateRegionMutationVariables>({
    mutationKey: mutationKeys.regions.create(),
    mutationFn: async ({ region, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.createRegion(region, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.regions.available(organizationId) })
      }
    },
  })
}
