/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationUsage } from '@daytona/billing-api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled, useBillingV2FlagLoaded } from '../useBillingV2Enabled'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationUsageQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const v2 = useBillingV2Enabled()
  const v2FlagLoaded = useBillingV2FlagLoaded()

  return useQuery<OrganizationUsage>({
    queryKey: queryKeys.organization.usage.current(organizationId, v2),
    queryFn: () => billingApi.getOrganizationUsage(organizationId, { v2 }),
    enabled: Boolean(enabled && v2FlagLoaded && config.billingApiUrl && organizationId),
  })
}
