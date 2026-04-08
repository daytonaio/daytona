/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotDto } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface ActivateSnapshotMutationVariables {
  snapshotId: string
  organizationId?: string
}

interface UseActivateSnapshotMutationOptions {
  invalidateOnSuccess?: boolean
}

export const useActivateSnapshotMutation = ({
  invalidateOnSuccess = true,
}: UseActivateSnapshotMutationOptions = {}) => {
  const { snapshotApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<SnapshotDto, unknown, ActivateSnapshotMutationVariables>({
    mutationFn: async ({ snapshotId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      const response = await snapshotApi.activateSnapshot(snapshotId, organizationId)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (invalidateOnSuccess && organizationId) {
        await queryClient.invalidateQueries({
          queryKey: queryKeys.snapshots.list(organizationId),
        })
      }
    },
  })
}
