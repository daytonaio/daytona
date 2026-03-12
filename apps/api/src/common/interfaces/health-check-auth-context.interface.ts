/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface HealthCheckAuthContext extends BaseAuthContext {
  role: 'health-check'
}

export function isHealthCheckAuthContext(user: BaseAuthContext): user is HealthCheckAuthContext {
  return 'role' in user && user.role === 'health-check'
}
