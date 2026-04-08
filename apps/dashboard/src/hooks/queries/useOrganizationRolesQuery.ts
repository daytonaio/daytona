/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRole } from '@daytona/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

export const useOrganizationRolesQuery = () => {
  const { organizationsApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<OrganizationRole[]>({
    queryKey: queryKeys.organization.roles(selectedOrganization?.id ?? ''),
    enabled: Boolean(selectedOrganization),
    queryFn: async () => {
      if (!selectedOrganization) {
        return []
      }

      const response = await organizationsApi.listOrganizationRoles(selectedOrganization.id)
      return response.data
    },
  })
}
