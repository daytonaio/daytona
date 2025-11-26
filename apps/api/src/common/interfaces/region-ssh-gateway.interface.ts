/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface RegionSSHGatewayContext extends BaseAuthContext {
  role: 'region-ssh-gateway'
  regionId: string
}

export function isRegionSSHGatewayContext(user: BaseAuthContext): user is RegionSSHGatewayContext {
  return 'role' in user && user.role === 'region-ssh-gateway'
}
