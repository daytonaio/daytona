/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class NoAvailableRunnersError extends BadRequestException {
  constructor(message = 'No available runners') {
    super({ message, code: ApiErrorCode.NO_AVAILABLE_RUNNERS })
  }
}
