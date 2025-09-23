/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface RunnerContext extends BaseAuthContext {
  role: 'runner'
  runnerId: string
}

export function isRunnerContext(user: BaseAuthContext): user is RunnerContext {
  return 'role' in user && user.role === 'runner'
}
