/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'
import { RegionAuthContext } from './region-auth-context.interface'

export interface RegionProxyAuthContext extends RegionAuthContext {
  role: 'region-proxy'
}

export function isRegionProxyAuthContext(user: BaseAuthContext): user is RegionProxyAuthContext {
  return 'role' in user && user.role === 'region-proxy'
}
