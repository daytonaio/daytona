/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKeyResponse, CreateApiKeyPermissionsEnum } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CreateApiKeyMutationVariables {
  name: string
  permissions: CreateApiKeyPermissionsEnum[]
  expiresAt: Date | null
  organizationId?: string
}

export const useCreateApiKeyMutation = () => {
  const { apiKeyApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<ApiKeyResponse, unknown, CreateApiKeyMutationVariables>({
    mutationFn: async ({ organizationId, name, permissions, expiresAt }) => {
      const response = await apiKeyApi.createApiKey(
        {
          name,
          permissions,
          expiresAt: expiresAt ?? undefined,
        },
        organizationId,
      )

      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.list(organizationId) })
      }
    },
  })
}
