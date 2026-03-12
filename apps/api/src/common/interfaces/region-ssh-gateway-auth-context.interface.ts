/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'
import { RegionAuthContext } from './region-auth-context.interface'

export interface RegionSSHGatewayAuthContext extends RegionAuthContext {
  role: 'region-ssh-gateway'
}

export function isRegionSSHGatewayAuthContext(user: BaseAuthContext): user is RegionSSHGatewayAuthContext {
  return 'role' in user && user.role === 'region-ssh-gateway'
}
