/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException } from '@nestjs/common'
import { ApiErrorCode } from '../common/errors/api-error-code.enum'

export class OrganizationSuspendedError extends ForbiddenException {
  constructor(message = 'Organization is suspended') {
    super({ message, code: ApiErrorCode.ORGANIZATION_SUSPENDED })
  }
}
