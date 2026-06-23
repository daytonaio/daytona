/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePendingMutationKeys } from '@/hooks/usePendingMutationKeys'
import { mutationKeys } from './mutationKeys'

type RegionMutationVariables = {
  regionId?: string
}

const regionMutationSelectors = [
  {
    mutationKey: mutationKeys.regions.update(),
    getKey: (variables?: RegionMutationVariables) => variables?.regionId,
  },
  {
    mutationKey: mutationKeys.regions.remove(),
    getKey: (variables?: RegionMutationVariables) => variables?.regionId,
  },
  {
    mutationKey: mutationKeys.regions.regenerateProxyApiKey(),
    getKey: (variables?: RegionMutationVariables) => variables?.regionId,
  },
  {
    mutationKey: mutationKeys.regions.regenerateSshGatewayApiKey(),
    getKey: (variables?: RegionMutationVariables) => variables?.regionId,
  },
  {
    mutationKey: mutationKeys.regions.regenerateSnapshotManagerCredentials(),
    getKey: (variables?: RegionMutationVariables) => variables?.regionId,
  },
]

export function useMutatingRegions(): Set<string> {
  return usePendingMutationKeys<string, RegionMutationVariables>(regionMutationSelectors)
}
