/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PaymentUrl } from '@/billing-api/types/Invoice'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface TopUpWalletVariables {
  organizationId: string
  amountCents: number
}

export const useTopUpWalletMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<PaymentUrl, unknown, TopUpWalletVariables>({
    mutationFn: ({ organizationId, amountCents }) => billingApi.topUpWallet(organizationId, amountCents),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId) })
    },
  })
}
