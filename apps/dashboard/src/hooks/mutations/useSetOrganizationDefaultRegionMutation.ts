/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'

import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface SetOrganizationDefaultRegionVariables {
  organizationId: string
  defaultRegionId: string
}

export const useSetOrganizationDefaultRegionMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, defaultRegionId }: SetOrganizationDefaultRegionVariables) =>
      organizationsApi.setOrganizationDefaultRegion(organizationId, { defaultRegionId }),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() })
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) })
    },
  })
}
