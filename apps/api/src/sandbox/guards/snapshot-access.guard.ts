/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { SnapshotService } from '../services/snapshot.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class SnapshotAccessGuard implements CanActivate {
  constructor(private readonly snapshotService: SnapshotService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const snapshotId: string = request.params.snapshotId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const snapshot = await this.snapshotService.getSnapshot(snapshotId)
      if (authContext.role !== SystemRole.ADMIN && snapshot.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      request.snapshot = snapshot
      return true
    } catch (error) {
      throw new NotFoundException(`Image with ID ${snapshotId} not found`)
    }
  }
}
