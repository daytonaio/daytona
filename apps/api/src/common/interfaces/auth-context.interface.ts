/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKey } from '../../api-key/api-key.entity'
import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { SystemRole } from '../../user/enums/system-role.enum'

export interface AuthContext {
  userId: string
  email: string
  role: SystemRole
  apiKey?: ApiKey
  organizationId?: string
}

export interface OrganizationAuthContext extends AuthContext {
  organizationId: string
  organizationUser?: OrganizationUser
}
