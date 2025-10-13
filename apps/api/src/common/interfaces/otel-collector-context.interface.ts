/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface OtelCollectorContext extends BaseAuthContext {
  role: 'otel-collector'
}

export function isOtelCollectorContext(user: BaseAuthContext): user is OtelCollectorContext {
  return 'role' in user && user.role === 'otel-collector'
}
