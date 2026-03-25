/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface DeleteVolumeMutationVariables {
  volumeId: string
  organizationId?: string
}

interface UseDeleteVolumeMutationOptions {
  invalidateOnSuccess?: boolean
}

export const useDeleteVolumeMutation = ({ invalidateOnSuccess = true }: UseDeleteVolumeMutationOptions = {}) => {
  const { volumeApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeleteVolumeMutationVariables>({
    mutationFn: async ({ volumeId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      await volumeApi.deleteVolume(volumeId, organizationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      if (invalidateOnSuccess && organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.volumes.list(organizationId) })
      }
    },
  })
}
