/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface LinkAccountParams {
  provider: string
  userId: string
}

export const useLinkAccountMutation = () => {
  const { userApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ provider, userId }: LinkAccountParams) =>
      userApi.linkAccount({
        provider,
        userId,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.user.accountProviders() })
    },
  })
}
