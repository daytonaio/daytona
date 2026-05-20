/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'

interface VoidInvoiceVariables {
  organizationId: string
  invoiceId: string
}

export const useVoidInvoiceMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()
  const v2 = useBillingV2Enabled()

  return useMutation<void, unknown, VoidInvoiceVariables>({
    mutationFn: ({ organizationId, invoiceId }) => billingApi.voidInvoice(organizationId, invoiceId, { v2 }),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId, v2) })
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.invoices(organizationId, v2) })
    },
  })
}
