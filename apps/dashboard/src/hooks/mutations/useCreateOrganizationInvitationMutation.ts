/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationInvitation, OrganizationInvitation } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface CreateOrganizationInvitationMutationVariables {
  organizationId?: string
  invitation: CreateOrganizationInvitation
}

export const useCreateOrganizationInvitationMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<OrganizationInvitation, unknown, CreateOrganizationInvitationMutationVariables>({
    mutationKey: mutationKeys.organization.invitations.create(),
    mutationFn: async ({ organizationId, invitation }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.createOrganizationInvitation(organizationId, invitation)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.organization.invitations(organizationId) })
      }
    },
  })
}
