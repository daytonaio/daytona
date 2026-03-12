/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { BaseAuthContext } from './auth-context.interface'
import { UserAuthContext } from './user-auth-context.interface'

export interface OrganizationAuthContext extends UserAuthContext {
  organizationId: string
  organization: Organization
  organizationUser?: OrganizationUser
}

export function isOrganizationAuthContext(user: BaseAuthContext): user is OrganizationAuthContext {
  return 'organizationId' in user
}
