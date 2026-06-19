/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotManagerCredentials } from '@daytona/api-client'
import { useMutation } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface RegenerateRegionSnapshotManagerCredentialsMutationVariables {
  regionId: string
  organizationId?: string
}

export const useRegenerateRegionSnapshotManagerCredentialsMutation = () => {
  const { organizationsApi } = useApi()

  return useMutation<SnapshotManagerCredentials, unknown, RegenerateRegionSnapshotManagerCredentialsMutationVariables>({
    mutationKey: mutationKeys.regions.regenerateSnapshotManagerCredentials(),
    mutationFn: async ({ regionId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.regenerateSnapshotManagerCredentials(regionId, organizationId)
      return response.data
    },
  })
}
