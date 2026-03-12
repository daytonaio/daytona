/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface ProxyAuthContext extends BaseAuthContext {
  role: 'proxy'
}

export function isProxyAuthContext(user: BaseAuthContext): user is ProxyAuthContext {
  return 'role' in user && user.role === 'proxy'
}
