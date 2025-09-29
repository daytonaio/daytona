/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createParamDecorator, ExecutionContext } from '@nestjs/common'
import { DaemonContext } from '../interfaces/daemon-context.interface'

export const DaemonContextDecorator = createParamDecorator((data: unknown, ctx: ExecutionContext): DaemonContext => {
  const request = ctx.switchToHttp().getRequest()
  return request.user as DaemonContext
})
