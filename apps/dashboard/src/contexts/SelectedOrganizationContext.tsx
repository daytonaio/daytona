/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Organization, OrganizationRolePermissionsEnum, OrganizationUser } from '@daytonaio/api-client'

import { createContext } from 'react'

export interface ISelectedOrganizationContext {
  selectedOrganization: Organization | null
  organizationMembers: OrganizationUser[]
  refreshOrganizationMembers: () => Promise<OrganizationUser[]>
  authenticatedUserOrganizationMember: OrganizationUser | null
  authenticatedUserHasPermission: (permission: OrganizationRolePermissionsEnum) => boolean
  onSelectOrganization: (organizationId: string) => Promise<boolean>
}

export const SelectedOrganizationContext = createContext<ISelectedOrganizationContext | undefined>(undefined)
