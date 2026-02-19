/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { UseQueryOptions } from '@tanstack/react-query'
import { useOrganizationBillingPortalUrlQuery } from './useOrganizationBillingPortalUrlQuery'
import { useOrganizationInvoicesQuery } from './useOrganizationInvoicesQuery'
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

export function useOwnerWalletQuery(
  queryOptions?: Omit<UseQueryOptions<OrganizationWallet>, 'queryKey' | 'queryFn' | 'enabled'>,
) {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationWalletQuery({
    ...scope,
    ...queryOptions,
  })
}

export function useOwnerTierQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationTierQuery(scope)
}

export function useOwnerBillingPortalUrlQuery() {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationBillingPortalUrlQuery(scope)
}

export function useOwnerInvoicesQuery(page?: number, perPage?: number) {
  const scope = useSelectedOrgBillingScope()
  return useOrganizationInvoicesQuery({
    ...scope,
    page,
    perPage,
  })
}
