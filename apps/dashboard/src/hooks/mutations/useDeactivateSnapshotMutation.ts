/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface DeactivateSnapshotMutationVariables {
  snapshotId: string
  organizationId?: string
}

interface UseDeactivateSnapshotMutationOptions {
  invalidateOnSuccess?: boolean
}

export const useDeactivateSnapshotMutation = ({
  invalidateOnSuccess = true,
}: UseDeactivateSnapshotMutationOptions = {}) => {
  const { snapshotApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeactivateSnapshotMutationVariables>({
    mutationFn: async ({ snapshotId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }
      await snapshotApi.deactivateSnapshot(snapshotId, organizationId)
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
