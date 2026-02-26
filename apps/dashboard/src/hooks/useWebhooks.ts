/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useWebhookInitializationStatusQuery } from './queries/useWebhookInitializationStatusQuery'
import { useSelectedOrganization } from './useSelectedOrganization'

export function useWebhooks() {
  const { selectedOrganization } = useSelectedOrganization()
  const { data, isLoading, refetch } = useWebhookInitializationStatusQuery(selectedOrganization?.id)

  const isInitialized = data?.svixApplicationId != null

  return {
    isInitialized,
    isLoading,
    refresh: refetch,
  }
}
