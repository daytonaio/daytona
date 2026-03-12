/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface OtelCollectorAuthContext extends BaseAuthContext {
  role: 'otel-collector'
}

export function isOtelCollectorAuthContext(user: BaseAuthContext): user is OtelCollectorAuthContext {
  return 'role' in user && user.role === 'otel-collector'
}
