/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useSvix } from 'svix-react'

interface DeleteWebhookEndpointVariables {
  endpointId: string
}

export const useDeleteWebhookEndpointMutation = () => {
  const { svix, appId } = useSvix()

  return useMutation({
    mutationFn: async ({ endpointId }: DeleteWebhookEndpointVariables) => {
      await svix.endpoint.delete(appId, endpointId)
    },
  })
}
