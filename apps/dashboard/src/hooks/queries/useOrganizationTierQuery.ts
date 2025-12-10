/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationTier } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationTierQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<OrganizationTier | null>({
    queryKey: queryKeys.organization.tier(organizationId),
    queryFn: () => billingApi.getOrganizationTier(organizationId),
    enabled: Boolean(enabled && organizationId && config.billingApiUrl),
  })
}
