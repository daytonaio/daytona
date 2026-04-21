/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
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
  const v2 = useBillingV2Enabled()

  return useQuery<string | null>({
    queryKey: queryKeys.billing.portalUrl(organizationId, v2),
    queryFn: () => billingApi.getOrganizationBillingPortalUrl(organizationId, { v2 }),
    enabled: Boolean(enabled && organizationId && config.billingApiUrl),
  })
}
