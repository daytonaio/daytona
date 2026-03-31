/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegionAuthContext, isRegionAuthContext } from './region-auth-context.interface'

export interface RegionSSHGatewayAuthContext extends RegionAuthContext {
  role: 'region-ssh-gateway'
}

export function isRegionSSHGatewayAuthContext(user: unknown): user is RegionSSHGatewayAuthContext {
  return isRegionAuthContext(user) && user.role === 'region-ssh-gateway'
}
