/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useAuth } from 'react-oidc-context'
import { useCallback } from 'react'
import { WebhookInitializationStatus } from '@daytonaio/api-client'

export function useWebhookService() {
  const { user } = useAuth()

  const getInitializationStatus = useCallback(
    async (organizationId: string): Promise<WebhookInitializationStatus | null> => {
      try {
        // Create a simple fetch request with the access token
        // Note: We don't need to include /api in the URL since the Vite dev server proxy handles it
        const response = await fetch(`/api/webhooks/organizations/${organizationId}/initialization-status`, {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${user?.access_token || ''}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          return null
        }

        return await response.json()
      } catch (error) {
        console.error('Failed to get webhook initialization status:', error)
        return null
      }
    },
    [user?.access_token],
  )

  const getAppPortalAccess = useCallback(
    async (organizationId: string): Promise<string | null> => {
      try {
        // Note: We don't need to include /api in the URL since the Vite dev server proxy handles it
        const response = await fetch(`/api/webhooks/organizations/${organizationId}/app-portal-access`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${user?.access_token || ''}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          return null
        }

        const data = await response.json()
        return data.url
      } catch (error) {
        console.error('Failed to get app portal access:', error)
        return null
      }
    },
    [user?.access_token],
  )

  const isWebhookInitialized = useCallback(
    async (organizationId: string): Promise<boolean> => {
      const status = await getInitializationStatus(organizationId)
      return status !== null && status.svixApplicationId !== null
    },
    [getInitializationStatus],
  )

  return {
    getInitializationStatus,
    getAppPortalAccess,
    isWebhookInitialized,
  }
}
