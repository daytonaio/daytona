/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useSvix } from 'svix-react'

interface ReplayWebhookEventVariables {
  endpointId: string
  msgId: string
}

export const useReplayWebhookEventMutation = () => {
  const { svix, appId } = useSvix()

  return useMutation({
    mutationFn: async ({ endpointId, msgId }: ReplayWebhookEventVariables) => {
      return svix.messageAttempt.resend(appId, msgId, endpointId)
    },
  })
}
