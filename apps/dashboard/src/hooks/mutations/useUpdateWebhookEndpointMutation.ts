/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useSvix } from 'svix-react'

interface UpdateWebhookEndpointVariables {
  endpointId: string
  update: {
    url: string
    description?: string
    filterTypes?: string[]
  }
}

export const useUpdateWebhookEndpointMutation = () => {
  const { svix, appId } = useSvix()

  return useMutation({
    mutationFn: ({ endpointId, update }: UpdateWebhookEndpointVariables) =>
      svix.endpoint.update(appId, endpointId, update),
  })
}
