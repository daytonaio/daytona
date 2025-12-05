/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface AddOrganizationEmailVariables {
  organizationId: string
  email: string
}

export const useAddOrganizationEmailMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, email }: AddOrganizationEmailVariables) =>
      billingApi.addOrganizationEmail(organizationId, email),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.emails(organizationId) })
    },
  })
}
