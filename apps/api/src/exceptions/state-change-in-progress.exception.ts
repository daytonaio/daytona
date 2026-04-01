/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, HttpStatus } from '@nestjs/common'

export class StateChangeInProgressError extends HttpException {
  constructor(message = 'Sandbox state change in progress') {
    super(message, HttpStatus.CONFLICT)
  }
}
