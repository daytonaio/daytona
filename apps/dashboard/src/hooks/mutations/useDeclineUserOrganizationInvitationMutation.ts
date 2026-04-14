/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface DeclineUserOrganizationInvitationMutationVariables {
  invitationId: string
}

export const useDeclineUserOrganizationInvitationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, DeclineUserOrganizationInvitationMutationVariables>({
    mutationKey: mutationKeys.user.invitations.decline(),
    mutationFn: async ({ invitationId }) => {
      await organizationsApi.declineOrganizationInvitation(invitationId)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.user.invitations() })
    },
  })
}
