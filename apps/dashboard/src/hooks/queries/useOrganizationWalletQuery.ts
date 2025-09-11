/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useOrganizationWalletQuery = ({ organizationId }: { organizationId: string }) => {
  const { billingApi } = useApi()

  return useQuery<OrganizationWallet>({
    queryKey: queryKeys.organization.wallet(organizationId),
    queryFn: () => billingApi.getOrganizationWallet(organizationId),
    enabled: !!organizationId && !!import.meta.env.VITE_BILLING_API_URL,
  })
}
