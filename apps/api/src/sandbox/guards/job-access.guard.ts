/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  NotFoundException,
  ForbiddenException,
  Logger,
} from '@nestjs/common'
import { BaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { JobService } from '../services/job.service'
import { isRunnerContext, RunnerContext } from '../../common/interfaces/runner-context.interface'

@Injectable()
export class JobAccessGuard implements CanActivate {
  private readonly logger = new Logger(JobAccessGuard.name)

  constructor(private readonly jobService: JobService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const jobId: string = request.params.jobId || request.params.id

    // TODO: initialize authContext safely
    const authContext: BaseAuthContext = request.user

    try {
      const job = await this.jobService.findOne(jobId)
      if (!job) {
        throw new NotFoundException('Job not found')
      }

      if (!isRunnerContext(authContext)) {
        throw new ForbiddenException('User is not a runner')
      }

      const runnerContext = authContext as RunnerContext

      if (runnerContext.runnerId !== job.runnerId) {
        throw new ForbiddenException('Runner ID does not match job runner ID')
      }

      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException('Job not found')
    }
  }
}
