/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException } from '@nestjs/common'
import { SnapshotService } from '../services/snapshot.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { Snapshot } from '../entities/snapshot.entity'

@Injectable()
export class SnapshotAccessGuard implements CanActivate {
  constructor(private readonly snapshotService: SnapshotService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const snapshotId: string = request.params.snapshotId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    let snapshot: Snapshot

    try {
      snapshot = await this.snapshotService.getSnapshot(snapshotId)
    } catch (error) {
      // If not found by ID, try by name
      snapshot = await this.snapshotService.getSnapshotByName(snapshotId, authContext.organizationId)
    }

    if (authContext.role !== SystemRole.ADMIN && snapshot.organizationId !== authContext.organizationId) {
      throw new ForbiddenException('Request organization ID does not match resource organization ID')
    }
    request.snapshot = snapshot
    return true
  }
}
