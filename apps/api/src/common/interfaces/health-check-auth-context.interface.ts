/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'

export interface HealthCheckAuthContext extends BaseAuthContext {
  role: 'health-check'
}

export function isHealthCheckAuthContext(user: unknown): user is HealthCheckAuthContext {
  return isBaseAuthContext(user) && user.role === 'health-check'
}
