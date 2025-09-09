/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Tier } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useTiersQuery = () => {
  const { billingApi } = useApi()

  return useQuery<Tier[]>({
    queryKey: queryKeys.billing.tiers(),
    queryFn: () => billingApi.listTiers(),
    enabled: !!import.meta.env.VITE_BILLING_API_URL,
  })
}
