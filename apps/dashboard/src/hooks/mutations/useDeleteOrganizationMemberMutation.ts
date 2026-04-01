/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { mutationKeys } from './mutationKeys'
import { useApi } from '../useApi'

export interface DeleteOrganizationMemberMutationVariables {
  organizationId?: string
  userId: string
}

export const useDeleteOrganizationMemberMutation = () => {
  const { organizationsApi } = useApi()

  return useMutation<void, unknown, DeleteOrganizationMemberMutationVariables>({
    mutationKey: mutationKeys.organization.members.remove(),
    mutationFn: async ({ organizationId, userId }) => {
      if (!organizationId) {
        throw new Error('No organization selected')
      }

      await organizationsApi.deleteOrganizationMember(organizationId, userId)
    },
  })
}
