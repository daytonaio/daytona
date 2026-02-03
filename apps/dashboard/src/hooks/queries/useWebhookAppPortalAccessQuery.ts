/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

interface WebhookAppPortalAccess {
  token: string
  url: string
}

export const useWebhookAppPortalAccessQuery = (organizationId?: string) => {
  const { axiosInstance } = useApi()

  return useQuery<WebhookAppPortalAccess>({
    queryKey: organizationId ? queryKeys.webhooks.appPortalAccess(organizationId) : queryKeys.webhooks.all,
    enabled: Boolean(organizationId),
    queryFn: async () => {
      if (!organizationId) {
        throw new Error('Organization ID is required')
      }
      const response = await axiosInstance.post<WebhookAppPortalAccess>(
        `/webhooks/organizations/${organizationId}/app-portal-access`,
      )
      return response.data
    },
    staleTime: 1000 * 60 * 5, // Token is valid for some time, cache for 5 minutes
    retry: 1,
  })
}
