/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { BillingInfo } from '@daytona/billing-api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useBillingInfoQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<BillingInfo>({
    queryKey: queryKeys.billing.billingInfo(organizationId),
    queryFn: () => billingApi.getBillingInfo(organizationId),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
