/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationBillingPortalUrlQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<string | null>({
    queryKey: queryKeys.billing.portalUrl(organizationId),
    queryFn: () => billingApi.getOrganizationBillingPortalUrl(organizationId),
    enabled: Boolean(enabled && organizationId && config.billingApiUrl),
  })
}
