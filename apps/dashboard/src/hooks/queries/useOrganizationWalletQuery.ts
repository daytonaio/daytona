/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationWallet } from '@daytona/billing-api-client'
import { useQuery, UseQueryOptions } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationWalletQuery = ({
  organizationId,
  enabled = true,
  ...queryOptions
}: {
  organizationId: string
  enabled?: boolean
} & Omit<UseQueryOptions<OrganizationWallet>, 'queryKey' | 'queryFn'>) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const v2 = useBillingV2Enabled()

  return useQuery<OrganizationWallet>({
    queryKey: queryKeys.organization.wallet(organizationId, v2),
    queryFn: () => billingApi.getOrganizationWallet(organizationId, { v2 }),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
    refetchOnWindowFocus: true,
    ...queryOptions,
  })
}
