/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { UpdateOrganizationDefaultVolumeBackendDefaultVolumeBackendEnum } from '@daytona/api-client'

import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface SetOrganizationDefaultVolumeBackendVariables {
  organizationId: string
  defaultVolumeBackend: string
}

export const useSetOrganizationDefaultVolumeBackendMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, defaultVolumeBackend }: SetOrganizationDefaultVolumeBackendVariables) =>
      organizationsApi.setDefaultVolumeBackend(organizationId, {
        defaultVolumeBackend: defaultVolumeBackend as UpdateOrganizationDefaultVolumeBackendDefaultVolumeBackendEnum,
      }),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() })
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) })
    },
  })
}
