/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { UnauthorizedException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class ApiKeyExpiredError extends UnauthorizedException {
  constructor(message = 'This API key has expired') {
    super({ message, code: ApiErrorCode.API_KEY_EXPIRED })
  }
}
