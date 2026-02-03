/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useSvix } from 'svix-react'

interface RotateWebhookSecretVariables {
  endpointId: string
}

export const useRotateWebhookSecretMutation = () => {
  const { svix, appId } = useSvix()

  return useMutation({
    mutationFn: async ({ endpointId }: RotateWebhookSecretVariables) => {
      return svix.endpoint.rotateSecret(appId, endpointId, {})
    },
  })
}
