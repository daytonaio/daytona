/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateSnapshot, SnapshotDto } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CreateSnapshotMutationVariables {
  snapshot: CreateSnapshot
  organizationId?: string
}

export const useCreateSnapshotMutation = () => {
  const { snapshotApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<SnapshotDto, unknown, CreateSnapshotMutationVariables>({
    mutationFn: async ({ snapshot, organizationId }) => {
      const response = await snapshotApi.createSnapshot(snapshot, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.snapshots.all })
      }
    },
  })
}
