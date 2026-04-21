/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationUsage } from '@daytona/billing-api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const usePastOrganizationUsageQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const v2 = useBillingV2Enabled()

  return useQuery<OrganizationUsage[]>({
    queryKey: queryKeys.organization.usage.past(organizationId, v2),
    queryFn: () => billingApi.getPastOrganizationUsage(organizationId, undefined, { v2 }),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
