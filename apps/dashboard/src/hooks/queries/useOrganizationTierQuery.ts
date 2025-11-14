/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationTier } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useOrganizationTierQuery = ({ organizationId }: { organizationId: string }) => {
  const { billingApi } = useApi()

  return useQuery<OrganizationTier | null>({
    queryKey: queryKeys.organization.tier(organizationId),
    queryFn: () => billingApi.getOrganizationTier(organizationId),
    enabled: !!organizationId && !!import.meta.env.VITE_BILLING_API_URL,
  })
}
