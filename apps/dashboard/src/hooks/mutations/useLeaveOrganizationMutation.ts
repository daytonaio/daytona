/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'

import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface LeaveOrganizationVariables {
  organizationId: string
}

export const useLeaveOrganizationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId }: LeaveOrganizationVariables) => organizationsApi.leaveOrganization(organizationId),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() })
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) })
    },
  })
}
