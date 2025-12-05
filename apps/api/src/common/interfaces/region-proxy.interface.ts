/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface RegionProxyContext extends BaseAuthContext {
  role: 'region-proxy'
  regionId: string
}

export function isRegionProxyContext(user: BaseAuthContext): user is RegionProxyContext {
  return 'role' in user && user.role === 'region-proxy'
}
