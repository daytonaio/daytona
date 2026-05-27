/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class SandboxBackupStateError extends BadRequestException {
  constructor(message: string) {
    super({ message, code: ApiErrorCode.SANDBOX_BACKUP_STATE_ERROR })
  }
}
