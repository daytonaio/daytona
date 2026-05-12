/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useApi } from '../useApi'

export const useRefreshWebhookEndpointFlagMutation = () => {
  const { axiosInstance } = useApi()

  return useMutation({
    mutationFn: async (organizationId: string) => {
      await axiosInstance.post(`/webhooks/organizations/${organizationId}/refresh-endpoints`)
    },
  })
}
