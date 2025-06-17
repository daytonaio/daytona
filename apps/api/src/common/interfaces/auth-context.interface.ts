/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKey } from '../../api-key/api-key.entity'
import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { SystemRole } from '../../user/enums/system-role.enum'

export interface BaseAuthContext {
  role: ApiRole
}

export type ApiRole = SystemRole | 'proxy'

export interface AuthContext extends BaseAuthContext {
  userId: string
  email: string
  apiKey?: ApiKey
  organizationId?: string
}

export function isAuthContext(user: BaseAuthContext): user is AuthContext {
  return 'userId' in user
}

export interface OrganizationAuthContext extends AuthContext {
  organizationId: string
  organization: Organization
  organizationUser?: OrganizationUser
}

export function isOrganizationAuthContext(user: BaseAuthContext): user is OrganizationAuthContext {
  return 'organizationId' in user
}
