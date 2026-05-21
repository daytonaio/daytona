/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { WebhookInitializationStatus } from '@daytona/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export const useInitializeWebhooksMutation = () => {
  const { webhooksApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<WebhookInitializationStatus, unknown, string>({
    mutationFn: async (organizationId: string) => {
      const response = await webhooksApi.webhookControllerInitializeWebhooks(organizationId)
      return response.data
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all })
    },
  })
}
