/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'
import { Runner } from '../../sandbox/entities/runner.entity'

export interface RunnerAuthContext extends BaseAuthContext {
  role: 'runner'
  runnerId: string
  runner: Runner
}

export function isRunnerAuthContext(user: unknown): user is RunnerAuthContext {
  return isBaseAuthContext(user) && user.role === 'runner'
}
