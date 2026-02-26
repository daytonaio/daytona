/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useMutation } from '@tanstack/react-query'

interface CreateSshAccessVariables {
  sandboxId: string
  expiresInMinutes: number
}

export const useCreateSshAccessMutation = () => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useMutation({
    mutationFn: async ({ sandboxId, expiresInMinutes }: CreateSshAccessVariables) => {
      const response = await sandboxApi.createSshAccess(sandboxId, selectedOrganization?.id, expiresInMinutes)
      return response.data
    },
  })
}
