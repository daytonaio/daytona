/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { isRegionAuthContext, RegionAuthContext } from './region-auth-context.interface'

export interface RegionProxyAuthContext extends RegionAuthContext {
  role: 'region-proxy'
}

export function isRegionProxyAuthContext(user: unknown): user is RegionProxyAuthContext {
  return isRegionAuthContext(user) && user.role === 'region-proxy'
}
