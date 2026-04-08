/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { handleApiError } from '@/lib/error-handling'
import { OrganizationUser } from '@daytona/api-client'
import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

type OrganizationsApiClient = ReturnType<typeof useApi>['organizationsApi']

export const getOrganizationMembersQueryOptions = (
  organizationsApi: OrganizationsApiClient,
  organizationId?: string | null,
) => ({
  queryKey: queryKeys.organization.members(organizationId ?? ''),
  queryFn: async (): Promise<OrganizationUser[]> => {
    if (!organizationId) {
      return []
    }

    try {
      const response = await organizationsApi.listOrganizationMembers(organizationId)
      return response.data
    } catch (error) {
      handleApiError(error, 'Failed to fetch organization members')
      throw error
    }
  },
})

export function useOrganizationMembersQuery(organizationId?: string | null) {
  const { organizationsApi } = useApi()

  return useQuery<OrganizationUser[]>({
    ...getOrganizationMembersQueryOptions(organizationsApi, organizationId),
    enabled: !!organizationId,
  })
}

export function useOrganizationMembersSuspenseQuery(organizationId?: string | null) {
  const { organizationsApi } = useApi()

  return useSuspenseQuery<OrganizationUser[]>(getOrganizationMembersQueryOptions(organizationsApi, organizationId))
}
