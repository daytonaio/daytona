/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'
import { mutationKeys } from './mutationKeys'

export interface AcceptUserOrganizationInvitationMutationVariables {
  invitationId: string
  organizationId: string
}

export const useAcceptUserOrganizationInvitationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<void, unknown, AcceptUserOrganizationInvitationMutationVariables>({
    mutationKey: mutationKeys.user.invitations.accept(),
    mutationFn: async ({ invitationId }) => {
      await organizationsApi.acceptOrganizationInvitation(invitationId)
    },
    onSuccess: async (_data, { organizationId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.user.invitations() }),
        queryClient.invalidateQueries({ queryKey: queryKeys.organization.list() }),
        queryClient.invalidateQueries({ queryKey: queryKeys.organization.detail(organizationId) }),
      ])
    },
  })
}
