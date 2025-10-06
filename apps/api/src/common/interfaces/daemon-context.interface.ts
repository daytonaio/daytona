/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiRole, BaseAuthContext } from './auth-context.interface'

export interface DaemonContext extends BaseAuthContext {
  role: ApiRole
  sandboxId: string
}

export function isDaemonContext(user: BaseAuthContext): user is DaemonContext {
  return 'role' in user && user.role === 'daemon'
}
