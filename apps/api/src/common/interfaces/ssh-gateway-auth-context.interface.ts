/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'

export interface SshGatewayAuthContext extends BaseAuthContext {
  role: 'ssh-gateway'
}

export function isSshGatewayAuthContext(user: unknown): user is SshGatewayAuthContext {
  return isBaseAuthContext(user) && user.role === 'ssh-gateway'
}
