/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AutomaticTopUp } from '@daytona/billing-api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'

interface SetAutomaticTopUpVariables {
  organizationId: string
  automaticTopUp?: AutomaticTopUp
}

export const useSetAutomaticTopUpMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()
  const v2 = useBillingV2Enabled()

  return useMutation({
    mutationFn: ({ organizationId, automaticTopUp }: SetAutomaticTopUpVariables) =>
      billingApi.setAutomaticTopUp(organizationId, automaticTopUp, { v2 }),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId, v2) })
    },
  })
}
