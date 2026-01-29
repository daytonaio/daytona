/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface VoidInvoiceVariables {
  organizationId: string
  invoiceId: string
}

export const useVoidInvoiceMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, VoidInvoiceVariables>({
    mutationFn: ({ organizationId, invoiceId }) => billingApi.voidInvoice(organizationId, invoiceId),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.invoices(organizationId) })
    },
  })
}
