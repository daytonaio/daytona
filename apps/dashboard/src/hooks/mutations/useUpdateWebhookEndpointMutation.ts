/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { EndpointPatch } from 'svix'
import { useSvix } from 'svix-react'

interface UpdateWebhookEndpointVariables {
  endpointId: string
  update: EndpointPatch
}

export const useUpdateWebhookEndpointMutation = () => {
  const { svix, appId } = useSvix()

  return useMutation({
    mutationFn: async ({ endpointId, update }: UpdateWebhookEndpointVariables) => {
      return svix.endpoint.patch(appId, endpointId, update)
    },
  })
}
