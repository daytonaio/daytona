/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect, useCallback } from 'react'
import { useWebhookService } from '@/services/webhookService'
import { useSelectedOrganization } from './useSelectedOrganization'

export function useWebhooks() {
  const { selectedOrganization } = useSelectedOrganization()
  const { isWebhookInitialized, getAppPortalAccess } = useWebhookService()
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [appPortalUrl, setAppPortalUrl] = useState<string | null>(null)

  const checkInitializationStatus = useCallback(async () => {
    if (!selectedOrganization?.id) {
      setIsInitialized(false)
      setIsLoading(false)
      return
    }

    try {
      setIsLoading(true)
      const initialized = await isWebhookInitialized(selectedOrganization.id)
      setIsInitialized(initialized)

      if (initialized) {
        const url = await getAppPortalAccess(selectedOrganization.id)
        setAppPortalUrl(url)
      }
    } catch (error) {
      console.error('Failed to check webhook initialization status:', error)
      setIsInitialized(false)
    } finally {
      setIsLoading(false)
    }
  }, [selectedOrganization?.id, isWebhookInitialized, getAppPortalAccess])

  useEffect(() => {
    checkInitializationStatus()
  }, [checkInitializationStatus])

  const openAppPortal = useCallback(() => {
    if (appPortalUrl) {
      window.open(appPortalUrl, '_blank', 'noopener,noreferrer')
    }
  }, [appPortalUrl])

  return {
    isInitialized,
    isLoading,
    appPortalUrl,
    openAppPortal,
    refresh: checkInitializationStatus,
  }
}
