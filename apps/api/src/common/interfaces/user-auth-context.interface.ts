/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKey } from '../../api-key/api-key.entity'
import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'

export interface UserAuthContext extends BaseAuthContext {
  userId: string
  email: string
  apiKey?: ApiKey
  organizationId?: string
}

export function isUserAuthContext(user: unknown): user is UserAuthContext {
  return isBaseAuthContext(user) && 'userId' in user
}
