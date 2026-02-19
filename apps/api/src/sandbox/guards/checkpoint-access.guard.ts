/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException, NotFoundException } from '@nestjs/common'
import { CheckpointService } from '../services/checkpoint.service'
import {
  BaseAuthContext,
  isOrganizationAuthContext,
  OrganizationAuthContext,
} from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class CheckpointAccessGuard implements CanActivate {
  constructor(private readonly checkpointService: CheckpointService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const checkpointId: string = request.params.checkpointId || request.params.id

    if (!checkpointId) {
      return true
    }

    const authContext: BaseAuthContext = request.user

    if (!isOrganizationAuthContext(authContext)) {
      throw new ForbiddenException('Organization context is required')
    }

    const orgAuthContext = authContext as OrganizationAuthContext

    try {
      const checkpoint = await this.checkpointService.getCheckpointById(checkpointId)

      if (orgAuthContext.role !== SystemRole.ADMIN && checkpoint.organizationId !== orgAuthContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }

      request.checkpoint = checkpoint

      return true
    } catch (error) {
      if (error instanceof ForbiddenException) {
        throw error
      }
      throw new NotFoundException(`Checkpoint ${checkpointId} not found`)
    }
  }
}
