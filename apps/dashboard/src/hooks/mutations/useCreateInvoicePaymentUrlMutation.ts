/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PaymentUrl } from '@daytona/billing-api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'

interface CreateInvoicePaymentUrlVariables {
  organizationId: string
  invoiceId: string
}

export const useCreateInvoicePaymentUrlMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()
  const v2 = useBillingV2Enabled()

  return useMutation<PaymentUrl, unknown, CreateInvoicePaymentUrlVariables>({
    mutationFn: ({ organizationId, invoiceId }) =>
      billingApi.createInvoicePaymentUrl(organizationId, invoiceId, { v2 }),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId, v2) })
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.invoices(organizationId, v2) })
    },
  })
}
