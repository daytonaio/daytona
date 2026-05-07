/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OtelConfig } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'

import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface UpdateOrganizationOtelConfigVariables {
  organizationId: string
  otelConfig: OtelConfig
}

export const useUpdateOrganizationOtelConfigMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, UpdateOrganizationOtelConfigVariables>({
    mutationFn: async ({ organizationId, otelConfig }) => {
      await organizationsApi.updateOrganizationOtelConfig(organizationId, otelConfig)
    },
    onSuccess: async (_data, { organizationId }) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() })
      await queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) })
    },
  })
}
