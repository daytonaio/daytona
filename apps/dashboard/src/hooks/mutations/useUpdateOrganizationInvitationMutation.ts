/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationInvitation, UpdateOrganizationInvitation } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface UpdateOrganizationInvitationMutationVariables {
  organizationId?: string
  invitationId: string
  invitation: UpdateOrganizationInvitation
}

export const useUpdateOrganizationInvitationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<OrganizationInvitation, unknown, UpdateOrganizationInvitationMutationVariables>({
    mutationKey: mutationKeys.organization.invitations.update(),
    mutationFn: async ({ organizationId, invitationId, invitation }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.updateOrganizationInvitation(organizationId, invitationId, invitation)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.organization.invitations(organizationId) })
      }
    },
  })
}
