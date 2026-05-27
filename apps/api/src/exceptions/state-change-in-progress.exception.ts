/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class StateChangeInProgressError extends ConflictException {
  constructor(message = 'Sandbox state change in progress') {
    super({ message, code: ApiErrorCode.SANDBOX_STATE_CHANGE_IN_PROGRESS })
  }
}
