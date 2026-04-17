/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { handleApiError } from '@/lib/error-handling'
import { Organization } from '@daytona/api-client'
import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

type OrganizationsApiClient = ReturnType<typeof useApi>['organizationsApi']

export const getOrganizationsQueryOptions = (organizationsApi: OrganizationsApiClient) => ({
  queryKey: queryKeys.organization.list(),
  queryFn: async (): Promise<Organization[]> => {
    try {
      return (await organizationsApi.listOrganizations()).data
    } catch (error) {
      handleApiError(error, 'Failed to fetch your organizations')
      throw error
    }
  },
})

export function useOrganizationsQuery() {
  const { organizationsApi } = useApi()

  return useQuery<Organization[]>(getOrganizationsQueryOptions(organizationsApi))
}

export function useOrganizationsSuspenseQuery() {
  const { organizationsApi } = useApi()

  return useSuspenseQuery<Organization[]>(getOrganizationsQueryOptions(organizationsApi))
}
