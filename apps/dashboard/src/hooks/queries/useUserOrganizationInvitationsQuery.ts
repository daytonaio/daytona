/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationInvitation } from '@daytona/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export function useUserOrganizationInvitationsQuery() {
  const { organizationsApi } = useApi()

  return useQuery<OrganizationInvitation[]>({
    queryKey: queryKeys.user.invitations(),
    queryFn: async () => {
      const response = await organizationsApi.listOrganizationInvitationsForAuthenticatedUser()
      return response.data
    },
  })
}
