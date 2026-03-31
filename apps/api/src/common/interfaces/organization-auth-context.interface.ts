/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { UserAuthContext, isUserAuthContext } from './user-auth-context.interface'

export interface OrganizationAuthContext extends UserAuthContext {
  organizationId: string
  organization: Organization
  organizationUser: OrganizationUser
}

export function isOrganizationAuthContext(user: unknown): user is OrganizationAuthContext {
  return isUserAuthContext(user) && 'organizationId' in user && 'organization' in user && 'organizationUser' in user
}
