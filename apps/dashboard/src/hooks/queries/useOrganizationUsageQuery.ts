/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationUsage } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationUsageQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<OrganizationUsage>({
    queryKey: queryKeys.organization.usage.current(organizationId),
    queryFn: () => billingApi.getOrganizationUsage(organizationId),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
