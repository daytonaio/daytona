/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationEmail } from '@/billing-api'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationEmailsQuery = ({
  organizationId,
  enabled = true,
}: {
  organizationId: string
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()

  return useQuery<OrganizationEmail[]>({
    queryKey: queryKeys.billing.emails(organizationId),
    queryFn: () => billingApi.listOrganizationEmails(organizationId),
    enabled: Boolean(enabled && organizationId && config.billingApiUrl),
  })
}
