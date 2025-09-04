/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createParamDecorator, ExecutionContext } from '@nestjs/common'
import { RunnerContext } from '../interfaces/runner-context.interface'

export const RunnerContextDecorator = createParamDecorator((data: unknown, ctx: ExecutionContext): RunnerContext => {
  const request = ctx.switchToHttp().getRequest()
  return request.user as RunnerContext
})
