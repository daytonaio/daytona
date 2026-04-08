/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationInvitation } from '@daytona/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

export function useOrganizationInvitationsQuery() {
  const { organizationsApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<OrganizationInvitation[]>({
    queryKey: queryKeys.organization.invitations(selectedOrganization?.id ?? ''),
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.listOrganizationInvitations(selectedOrganization.id)
      return response.data
    },
    enabled: !!selectedOrganization,
  })
}
