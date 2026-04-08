/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateVolume, VolumeDto } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CreateVolumeMutationVariables {
  volume: CreateVolume
  organizationId?: string
}

export const useCreateVolumeMutation = () => {
  const { volumeApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<VolumeDto, unknown, CreateVolumeMutationVariables>({
    mutationFn: async ({ volume, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      const response = await volumeApi.createVolume(volume, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.volumes.list(organizationId) })
      }
    },
  })
}
