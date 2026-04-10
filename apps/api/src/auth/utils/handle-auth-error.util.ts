/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, ServiceUnavailableException } from '@nestjs/common'

export function handleAuthError(error: unknown, message: string): void {
  if (!(error instanceof HttpException)) {
    throw new ServiceUnavailableException(message, { cause: error })
  }
}
