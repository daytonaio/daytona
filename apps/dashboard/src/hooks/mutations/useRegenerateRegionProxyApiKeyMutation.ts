/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegenerateApiKeyResponse } from '@daytona/api-client'
import { useMutation } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface RegenerateRegionProxyApiKeyMutationVariables {
  regionId: string
  organizationId?: string
}

export const useRegenerateRegionProxyApiKeyMutation = () => {
  const { organizationsApi } = useApi()

  return useMutation<RegenerateApiKeyResponse, unknown, RegenerateRegionProxyApiKeyMutationVariables>({
    mutationKey: mutationKeys.regions.regenerateProxyApiKey(),
    mutationFn: async ({ regionId, organizationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.regenerateProxyApiKey(regionId, organizationId)
      return response.data
    },
  })
}
