/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface SshGatewayAuthContext extends BaseAuthContext {
  role: 'ssh-gateway'
}

export function isSshGatewayAuthContext(user: BaseAuthContext): user is SshGatewayAuthContext {
  return 'role' in user && user.role === 'ssh-gateway'
}
