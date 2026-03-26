/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Tier } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useTiersQuery = ({ enabled = true }: { enabled?: boolean } = {}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<Tier[]>({
    queryKey: queryKeys.billing.tiers(),
    queryFn: () => billingApi.listTiers(),
    enabled: Boolean(enabled && config.billingApiUrl),
  })
}
