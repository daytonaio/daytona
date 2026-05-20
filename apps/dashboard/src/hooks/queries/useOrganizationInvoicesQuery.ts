/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { PaginatedTInvoice } from '@daytona/billing-api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useOrganizationInvoicesQuery = ({
  organizationId,
  page,
  perPage,
  enabled = true,
}: {
  organizationId: string
  page?: number
  perPage?: number
  enabled?: boolean
}) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const v2 = useBillingV2Enabled()

  return useQuery<PaginatedTInvoice>({
    queryKey: queryKeys.billing.invoices(organizationId, v2, page, perPage),
    queryFn: () => billingApi.listInvoices(organizationId, page, perPage, { v2 }),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
