/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, ServiceUnavailableException } from '@nestjs/common'

export function handleAuthError(error: unknown, message: string): void {
  if (!(error instanceof HttpException) || error.getStatus() >= 500) {
    throw new ServiceUnavailableException(message, { cause: error })
  }
}
