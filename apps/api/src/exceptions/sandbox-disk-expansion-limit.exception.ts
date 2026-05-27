/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class SandboxDiskExpansionLimitError extends ForbiddenException {
  constructor(message: string) {
    super({ message, code: ApiErrorCode.SANDBOX_DISK_EXPANSION_LIMIT })
  }
}
