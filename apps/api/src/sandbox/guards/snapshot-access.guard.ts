/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException, NotFoundException } from '@nestjs/common'
import { SnapshotService } from '../services/snapshot.service'
import { isBaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { SystemRole } from '../../user/enums/system-role.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isRegionProxyAuthContext } from '../../common/interfaces/region-proxy-auth-context.interface'
import { isRegionSSHGatewayAuthContext } from '../../common/interfaces/region-ssh-gateway-auth-context.interface'

@Injectable()
export class SnapshotAccessGuard implements CanActivate {
  constructor(private readonly snapshotService: SnapshotService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const snapshotId: string = request.params.snapshotId || request.params.id

    let snapshot: Snapshot

    const authContext = getAuthContext(context, isBaseAuthContext)

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
        case isRegionProxyAuthContext(authContext): {
          const isAvailable = await this.snapshotService.isAvailableInRegion(snapshot.id, authContext.regionId)
          if (!isAvailable) {
            throw new NotFoundException(`Snapshot is not available in region ${authContext.regionId}`)
          }
          break
        }
        case isRegionSSHGatewayAuthContext(authContext): {
          const isAvailable = await this.snapshotService.isAvailableInRegion(snapshot.id, authContext.regionId)
          if (!isAvailable) {
            throw new NotFoundException(`Snapshot is not available in region ${authContext.regionId}`)
          }
          break
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          break
        case isOrganizationAuthContext(authContext): {
          if (authContext.role !== SystemRole.ADMIN && snapshot.organizationId !== authContext.organizationId) {
            throw new ForbiddenException('Request organization ID does not match resource organization ID')
          }
          break
        }
        default:
          return false
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
