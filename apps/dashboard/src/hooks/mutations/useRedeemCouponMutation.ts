/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface RedeemCouponVariables {
  organizationId: string
  couponCode: string
}

export const useRedeemCouponMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<string, unknown, RedeemCouponVariables>({
    mutationFn: ({ organizationId, couponCode }) => billingApi.redeemCoupon(organizationId, couponCode),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId) })

      // a coupon can upgrade the tier
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.tier(organizationId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.usage.overview(organizationId) })
    },
  })
}
