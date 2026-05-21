/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { WebhookInitializationStatus } from '@daytona/api-client'
import { useQuery } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
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
      } catch (error) {
        // 404 means not initialized; let other errors surface.
        if (error instanceof Error && isAxiosError(error.cause) && error.cause.status === 404) {
          return null
        }
        throw error
      }
    },
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
    retry: false,
  })
}
