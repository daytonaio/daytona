/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useApi } from '../useApi'

interface ResendOrganizationEmailVerificationVariables {
  organizationId: string
  email: string
}

export const useResendOrganizationEmailVerificationMutation = () => {
  const { billingApi } = useApi()

  return useMutation({
    mutationFn: ({ organizationId, email }: ResendOrganizationEmailVerificationVariables) =>
      billingApi.resendOrganizationEmailVerification(organizationId, email),
  })
}
