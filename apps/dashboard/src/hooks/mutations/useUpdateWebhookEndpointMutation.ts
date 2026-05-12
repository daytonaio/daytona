/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { EndpointPatch } from 'svix'
import { useSvix } from 'svix-react'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useRefreshWebhookEndpointFlagMutation } from './useRefreshWebhookEndpointFlagMutation'

interface UpdateWebhookEndpointVariables {
  endpointId: string
  update: EndpointPatch
}

export const useUpdateWebhookEndpointMutation = () => {
  const { svix, appId } = useSvix()
  const { selectedOrganization } = useSelectedOrganization()
  const refreshFlag = useRefreshWebhookEndpointFlagMutation()

  return useMutation({
    mutationFn: async ({ endpointId, update }: UpdateWebhookEndpointVariables) => {
      const result = await svix.endpoint.patch(appId, endpointId, update)
      if (selectedOrganization?.id) {
        refreshFlag.mutate(selectedOrganization.id)
      }
      return result
    },
  })
}
