/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface OtelProxyContext extends BaseAuthContext {
  role: 'otel-proxy'
}

export function isOtelProxyContext(user: BaseAuthContext): user is OtelProxyContext {
  return 'role' in user && user.role === 'otel-proxy'
}
