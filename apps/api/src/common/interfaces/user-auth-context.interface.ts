/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKey } from '../../api-key/api-key.entity'
import { BaseAuthContext } from './auth-context.interface'

export interface UserAuthContext extends BaseAuthContext {
  userId: string
  email: string
  apiKey?: ApiKey
  organizationId?: string
}

export function isUserAuthContext(user: BaseAuthContext): user is UserAuthContext {
  return 'userId' in user
}
