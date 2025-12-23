/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException, NotFoundException } from '@nestjs/common'
import { SnapshotService } from '../services/snapshot.service'
import {
  BaseAuthContext,
  isOrganizationAuthContext,
  OrganizationAuthContext,
} from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { isSshGatewayContext } from '../../common/interfaces/ssh-gateway-context.interface'
import { isProxyContext } from '../../common/interfaces/proxy-context.interface'
import { isRegionProxyContext, RegionProxyContext } from '../../common/interfaces/region-proxy.interface'
import {
  isRegionSSHGatewayContext,
  RegionSSHGatewayContext,
} from '../../common/interfaces/region-ssh-gateway.interface'

@Injectable()
export class SnapshotAccessGuard implements CanActivate {
  constructor(private readonly snapshotService: SnapshotService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const snapshotId: string = request.params.snapshotId || request.params.id

    let snapshot: Snapshot

    // TODO: initialize authContext safely
    const authContext: BaseAuthContext = request.user

    try {
      snapshot = await this.snapshotService.getSnapshot(snapshotId)
    } catch {
      if (!isOrganizationAuthContext(authContext)) {
        throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
      }

      // If not found by ID, try by name
      snapshot = await this.snapshotService.getSnapshotByName(snapshotId, authContext.organizationId)
    }

    try {
      switch (true) {
        case isRegionProxyContext(authContext):
        case isRegionSSHGatewayContext(authContext): {
          // For region proxy/ssh gateway authentication, verify that the runner's region ID matches the region ID
          const regionContext = authContext as RegionProxyContext | RegionSSHGatewayContext
          const isAvailable = await this.snapshotService.isAvailableInRegion(snapshot.id, regionContext.regionId)
          if (!isAvailable) {
            throw new NotFoundException(`Snapshot is not available in region ${regionContext.regionId}`)
          }
          break
        }
        case isProxyContext(authContext):
        case isSshGatewayContext(authContext):
          break
        default: {
          // For user/organization authentication, check organization access
          const orgAuthContext = authContext as OrganizationAuthContext
          if (orgAuthContext.role !== SystemRole.ADMIN && snapshot.organizationId !== orgAuthContext.organizationId) {
            throw new ForbiddenException('Request organization ID does not match resource organization ID')
          }
        }
      }

      request.snapshot = snapshot

      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        console.error(error)
      }
      throw new NotFoundException(`Snapshot with ID or name ${snapshotId} not found`)
    }
  }
}
