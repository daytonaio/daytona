/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { PaginatedInvoices } from '@/billing-api/types/Invoice'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
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

  return useQuery<PaginatedInvoices>({
    queryKey: queryKeys.billing.invoices(organizationId, page, perPage),
    queryFn: () => billingApi.listInvoices(organizationId, page, perPage),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })
}
