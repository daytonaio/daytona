/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUser, UpdateOrganizationMemberAccess } from '@daytonaio/api-client'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'
import { queryKeys } from '../queries/queryKeys'
import { useApi } from '../useApi'

export interface UpdateOrganizationMemberAccessMutationVariables {
  organizationId?: string
  userId: string
  access: UpdateOrganizationMemberAccess
}

export const useUpdateOrganizationMemberAccessMutation = () => {
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()

  return useMutation<OrganizationUser, unknown, UpdateOrganizationMemberAccessMutationVariables>({
    mutationKey: mutationKeys.organization.members.updateAccess(),
    mutationFn: async ({ organizationId, userId, access }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      const response = await organizationsApi.updateAccessForOrganizationMember(organizationId, userId, access)
      return response.data
    },
    onSuccess: async (_data, { organizationId }) => {
      if (organizationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.organization.members(organizationId) })
      }
    },
  })
}
