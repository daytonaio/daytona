/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { useOrganizationBillingPortalUrlQuery } from './useOrganizationBillingPortalUrlQuery'
import { useOrganizationEmailsQuery } from './useOrganizationEmailsQuery'
import { useOrganizationTierQuery } from './useOrganizationTierQuery'
import { useOrganizationWalletQuery } from './useOrganizationWalletQuery'

function useSelectedOrgBillingScope() {
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const isOwner = authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER

  return {
    organizationId: selectedOrganization?.id ?? '',
    enabled: Boolean(selectedOrganization && isOwner),
  }
}

export function useOwnerWalletQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationWalletQuery(scope)
}

export function useOwnerTierQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationTierQuery(scope)
}

export function useOwnerOrganizationEmailsQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationEmailsQuery(scope)
}

export function useOwnerBillingPortalUrlQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationBillingPortalUrlQuery(scope)
}
