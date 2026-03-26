/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AutomaticTopUp } from '@/billing-api/types/OrganizationWallet'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface SetAutomaticTopUpVariables {
  organizationId: string
  automaticTopUp?: AutomaticTopUp
}

export const useSetAutomaticTopUpMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, automaticTopUp }: SetAutomaticTopUpVariables) =>
      billingApi.setAutomaticTopUp(organizationId, automaticTopUp),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.wallet(organizationId) })
    },
  })
}
