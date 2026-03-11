/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'

import { useApi } from '../useApi'

export const useDeleteAccountMutation = () => {
  const { userApi } = useApi()

  return useMutation({
    mutationFn: async () => {
      await userApi.deleteAuthenticatedUser()
    },
  })
}
