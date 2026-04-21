/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { ChargeList } from '@daytona/billing-api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useChargesQuery = ({ organizationId, enabled = true }: { organizationId: string; enabled?: boolean }) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const v2 = useBillingV2Enabled()

  return useQuery<ChargeList>({
    queryKey: queryKeys.billing.charges(organizationId),
    queryFn: () => billingApi.listCharges(organizationId),
    enabled: Boolean(enabled && v2 && config.billingApiUrl && organizationId),
  })
}
