/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

interface WebhookTokenResponse {
  token: string
  url: string
}

export const useWebhookTokenQuery = (organizationId?: string) => {
  const { axiosInstance } = useApi()

  return useQuery<WebhookTokenResponse>({
    queryKey: organizationId ? queryKeys.webhooks.token(organizationId) : queryKeys.webhooks.all,
    enabled: Boolean(organizationId),
    queryFn: async () => {
      if (!organizationId) {
        throw new Error('Organization ID is required')
      }
      const response = await axiosInstance.post<WebhookTokenResponse>(`/webhooks/organizations/${organizationId}/token`)
      return response.data
    },
    staleTime: 1000 * 60 * 5, // Token is valid for some time, cache for 5 minutes
    retry: 1,
  })
}
