/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

interface RevokeApiKeyVariables {
  userId: string
  name: string
  organizationId?: string
}

export const useRevokeApiKeyMutation = () => {
  const { apiKeyApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ userId, name, organizationId }: RevokeApiKeyVariables) =>
      apiKeyApi.deleteApiKeyForUser(userId, name, organizationId),
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.list(organizationId) })
      }
    },
  })
}
