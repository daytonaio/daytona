/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface SshGatewayContext extends BaseAuthContext {
  role: 'ssh-gateway'
}

export function isSshGatewayContext(user: BaseAuthContext): user is SshGatewayContext {
  return 'role' in user && (user.role === 'ssh-gateway' || user.role === 'admin')
}
