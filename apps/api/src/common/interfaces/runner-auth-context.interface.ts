/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'
import { Runner } from '../../sandbox/entities/runner.entity'

export interface RunnerAuthContext extends BaseAuthContext {
  role: 'runner'
  runnerId: string
  runner: Runner
}

export function isRunnerAuthContext(user: BaseAuthContext): user is RunnerAuthContext {
  return 'role' in user && user.role === 'runner'
}
