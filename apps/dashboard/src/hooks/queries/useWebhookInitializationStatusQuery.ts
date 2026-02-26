/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { WebhookInitializationStatus } from '@daytonaio/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useWebhookInitializationStatusQuery = (organizationId?: string) => {
  const { webhooksApi } = useApi()

  return useQuery<WebhookInitializationStatus | null>({
    queryKey: organizationId ? queryKeys.webhooks.initializationStatus(organizationId) : queryKeys.webhooks.all,
    enabled: Boolean(organizationId),
    queryFn: async () => {
      if (!organizationId) {
        return null
      }
      try {
        const response = await webhooksApi.webhookControllerGetInitializationStatus(organizationId)
        return response.data
      } catch {
        // If the endpoint returns 404, webhooks are not initialized
        return null
      }
    },
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
    retry: false, // Don't retry on 404
  })
}
