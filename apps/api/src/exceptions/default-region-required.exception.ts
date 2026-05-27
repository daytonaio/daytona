/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class DefaultRegionRequiredError extends BadRequestException {
  constructor(
    message = 'This organization does not have a default region. Please open the Daytona Dashboard to set a default region.',
  ) {
    super({ message, code: ApiErrorCode.DEFAULT_REGION_REQUIRED })
  }
}
