/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DockerRegistry } from '@daytona/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

export function useRegistriesQuery() {
  const { dockerRegistryApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<DockerRegistry[]>({
    queryKey: queryKeys.registries.list(selectedOrganization?.id ?? ''),
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const response = await dockerRegistryApi.listRegistries(selectedOrganization.id)
      return response.data
    },
    enabled: !!selectedOrganization,
  })
}
