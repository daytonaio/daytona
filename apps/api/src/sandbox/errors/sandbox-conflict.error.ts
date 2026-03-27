/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'

export class SandboxConflictError extends ConflictException {
  constructor() {
    super('Sandbox was modified by another operation')
  }
}
