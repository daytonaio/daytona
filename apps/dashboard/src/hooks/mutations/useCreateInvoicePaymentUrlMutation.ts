/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PaymentUrl } from '@/billing-api/types/Invoice'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface CreateInvoicePaymentUrlVariables {
  organizationId: string
  invoiceId: string
}

export const useCreateInvoicePaymentUrlMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<PaymentUrl, unknown, CreateInvoicePaymentUrlVariables>({
    mutationFn: ({ organizationId, invoiceId }) => billingApi.createInvoicePaymentUrl(organizationId, invoiceId),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.invoices(organizationId) })
    },
  })
}
