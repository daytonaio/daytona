/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class SnapshotStateChangeInProgressError extends BadRequestException {
  constructor(message: string) {
    super({ message, code: ApiErrorCode.SNAPSHOT_STATE_CHANGE_IN_PROGRESS })
  }
}
