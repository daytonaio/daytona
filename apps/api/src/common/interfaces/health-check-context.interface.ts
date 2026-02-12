/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface HealthCheckContext extends BaseAuthContext {
  role: 'health-check'
}

export function isHealthCheckContext(user: BaseAuthContext): user is HealthCheckContext {
  return 'role' in user && user.role === 'health-check'
}
