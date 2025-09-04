/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, UnauthorizedException, Logger } from '@nestjs/common'
import { RunnerService } from '../sandbox/services/runner.service'
import { RunnerContext } from '../common/interfaces/runner-context.interface'

@Injectable()
export class RunnerAuthGuard implements CanActivate {
  private readonly logger = new Logger(RunnerAuthGuard.name)

  constructor(private readonly runnerService: RunnerService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()

    // Extract token from Authorization header
    const authHeader = request.headers.authorization
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new UnauthorizedException('Missing or invalid authorization header')
    }

    const token = authHeader.substring(7) // Remove 'Bearer ' prefix

    try {
      // Check if this token belongs to a runner
      const runner = await this.runnerService.findByApiKey(token)
      if (!runner) {
        throw new UnauthorizedException('Invalid runner API key')
      }

      // Set the runner context on the request
      const runnerContext: RunnerContext = {
        role: 'runner',
        runnerId: runner.id,
      }

      request.user = runnerContext
      this.logger.debug(`Runner authenticated: ${runner.id}`)

      return true
    } catch (error) {
      this.logger.debug('Runner authentication failed:', error)
      throw new UnauthorizedException('Invalid runner API key')
    }
  }
}
