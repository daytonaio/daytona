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
import { JobService } from '../services/job.service'
import { isRunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'

@Injectable()
export class JobAccessGuard implements CanActivate {
  private readonly logger = new Logger(JobAccessGuard.name)

  constructor(private readonly jobService: JobService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const jobId: string = request.params.jobId || request.params.id

    const authContext = getAuthContext(context, isRunnerAuthContext)

    try {
      const job = await this.jobService.findOne(jobId)
      if (!job) {
        throw new NotFoundException('Job not found')
      }

      if (authContext.runnerId !== job.runnerId) {
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
