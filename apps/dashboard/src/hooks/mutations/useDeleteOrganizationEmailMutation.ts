/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface DeleteOrganizationEmailVariables {
  organizationId: string
  email: string
}

export const useDeleteOrganizationEmailMutation = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ organizationId, email }: DeleteOrganizationEmailVariables) =>
      billingApi.deleteOrganizationEmail(organizationId, email),
    onSuccess: (_data, { organizationId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.emails(organizationId) })
    },
  })
}
