/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'

import { useWebhookService } from '@/services/webhookService'

export const useWebhookInitializationStatusQuery = (organizationId?: string) => {
  const { getInitializationStatus } = useWebhookService()

  return useQuery({
    queryKey: ['webhooks', organizationId, 'initialization-status'],
    queryFn: async () => {
      if (!organizationId) {
        return null
      }

      return getInitializationStatus(organizationId)
    },
    enabled: Boolean(organizationId),
  })
}
