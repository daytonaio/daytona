/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'

import { useWebhookService } from '@/services/webhookService'

interface WebhookAppPortalAccessResult {
  token: string
}

export const useWebhookAppPortalAccessQuery = (organizationId?: string) => {
  const { getAppPortalAccess } = useWebhookService()

  return useQuery<WebhookAppPortalAccessResult | null>({
    queryKey: ['webhooks', organizationId, 'app-portal-access'],
    queryFn: async () => {
      if (!organizationId) {
        return null
      }

      const token = await getAppPortalAccess(organizationId)
      return token ? { token } : null
    },
    enabled: Boolean(organizationId),
  })
}
