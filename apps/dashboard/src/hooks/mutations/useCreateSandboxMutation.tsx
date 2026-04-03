/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateSandboxFromImageParams, CreateSandboxFromSnapshotParams, Daytona, Sandbox } from '@daytona/sdk'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from 'react-oidc-context'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { getSandboxesQueryKey } from '../useSandboxes'

export type CreateSandboxParams = (CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams) & {
  target?: string
}

export const useCreateSandboxMutation = () => {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  return useMutation<Sandbox, unknown, CreateSandboxParams>({
    mutationFn: async (params) => {
      if (!user?.access_token || !selectedOrganization?.id) {
        throw new Error('Missing authentication or organization')
      }

      const { target, ...createParams } = params
      const client = new Daytona({
        jwtToken: user.access_token,
        apiUrl: import.meta.env.VITE_API_URL,
        organizationId: selectedOrganization.id,
        target,
      })

      if ('image' in createParams) {
        return await client.create(createParams as CreateSandboxFromImageParams)
      }
      return await client.create(createParams as CreateSandboxFromSnapshotParams)
    },
    onSuccess: async () => {
      if (selectedOrganization?.id) {
        await queryClient.invalidateQueries({ queryKey: getSandboxesQueryKey(selectedOrganization.id) })
      }
    },
  })
}
