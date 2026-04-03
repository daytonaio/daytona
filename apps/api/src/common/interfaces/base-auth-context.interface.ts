/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SystemRole } from '../../user/enums/system-role.enum'
export interface BaseAuthContext {
  role:
    | SystemRole
    | 'proxy'
    | 'runner'
    | 'ssh-gateway'
    | 'region-proxy'
    | 'region-ssh-gateway'
    | 'otel-collector'
    | 'health-check'
}

export function isBaseAuthContext(user: unknown): user is BaseAuthContext {
  return typeof user === 'object' && user !== null && 'role' in user
}
