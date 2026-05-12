/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useSvix } from 'svix-react'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useRefreshWebhookEndpointFlagMutation } from './useRefreshWebhookEndpointFlagMutation'

interface DeleteWebhookEndpointVariables {
  endpointId: string
}

export const useDeleteWebhookEndpointMutation = () => {
  const { svix, appId } = useSvix()
  const { selectedOrganization } = useSelectedOrganization()
  const refreshFlag = useRefreshWebhookEndpointFlagMutation()

  return useMutation({
    mutationFn: async ({ endpointId }: DeleteWebhookEndpointVariables) => {
      await svix.endpoint.delete(appId, endpointId)
      if (selectedOrganization?.id) {
        refreshFlag.mutate(selectedOrganization.id)
      }
    },
  })
}
