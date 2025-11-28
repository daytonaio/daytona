/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface UnlinkAccountParams {
  provider: string
  userId: string
}

export const useUnlinkAccountMutation = () => {
  const { userApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ provider, userId }: UnlinkAccountParams) => userApi.unlinkAccount(provider, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.user.accountProviders() })
    },
  })
}
