/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useMutation } from '@tanstack/react-query'

export const useStartVncMutation = (sandboxId: string) => {
  const { toolboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useMutation({
    mutationFn: async () => {
      await toolboxApi.startComputerUseDeprecated(sandboxId, selectedOrganization?.id)
    },
  })
}
