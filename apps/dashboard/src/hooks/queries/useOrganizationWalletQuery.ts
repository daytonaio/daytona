/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationWalletQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<OrganizationWallet>({
    queryKey: queryKeys.organization.wallet(organizationId),
    queryFn: () => billingApi.getOrganizationWallet(organizationId),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
