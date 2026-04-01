/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CancelOrganizationInvitationMutationVariables {
  organizationId?: string
  invitationId: string
}

export const useCancelOrganizationInvitationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, CancelOrganizationInvitationMutationVariables>({
    mutationKey: mutationKeys.organization.invitations.cancel(),
    mutationFn: async ({ organizationId, invitationId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      await organizationsApi.cancelOrganizationInvitation(organizationId, invitationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.organization.invitations(organizationId) })
      }
    },
  })
}
