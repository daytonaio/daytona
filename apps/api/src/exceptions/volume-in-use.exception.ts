/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class VolumeInUseError extends ConflictException {
  constructor(message: string) {
    super({ message, code: ApiErrorCode.VOLUME_IN_USE })
  }
}
