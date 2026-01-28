/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKeyList } from '@daytonaio/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useApiKeysQuery = (organizationId?: string) => {
  const { apiKeyApi } = useApi()

  return useQuery<ApiKeyList[]>({
    queryKey: organizationId ? queryKeys.apiKeys.list(organizationId) : queryKeys.apiKeys.all,
    enabled: Boolean(organizationId),
    queryFn: async () => {
      if (!organizationId) {
        return []
      }
      const response = await apiKeyApi.listApiKeys(organizationId)
      return response.data
    },
  })
}
