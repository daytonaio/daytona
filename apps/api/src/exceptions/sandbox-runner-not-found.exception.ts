/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

// Thrown when a sandbox exists but has no runner assigned to it (the sandbox
// is in a state that doesn't currently occupy a runner: never started,
// archived, pending assignment, etc.).
export class SandboxRunnerNotFoundError extends NotFoundException {
  constructor(message: string) {
    super({ message, code: ApiErrorCode.SANDBOX_RUNNER_NOT_FOUND })
  }
}
