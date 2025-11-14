/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface DowngradeTierParams {
  organizationId: string
  tier: number
}

export const useDowngradeTierMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, tier }: DowngradeTierParams) => billingApi.downgradeTier(organizationId, tier),
    onSuccess: async (_data, { organizationId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.organization.tier(organizationId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.organization.usage.overview(organizationId) }),
      ])
    },
  })
}
