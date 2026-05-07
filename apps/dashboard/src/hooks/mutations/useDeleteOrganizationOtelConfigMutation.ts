/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'

import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface DeleteOrganizationOtelConfigVariables {
  organizationId: string
}

export const useDeleteOrganizationOtelConfigMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeleteOrganizationOtelConfigVariables>({
    mutationFn: async ({ organizationId }) => {
      await organizationsApi.deleteOrganizationOtelConfig(organizationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() })
      await queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) })
    },
  })
}
