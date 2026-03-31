/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'

export interface ProxyAuthContext extends BaseAuthContext {
  role: 'proxy'
}

export function isProxyAuthContext(user: unknown): user is ProxyAuthContext {
  return isBaseAuthContext(user) && user.role === 'proxy'
}
