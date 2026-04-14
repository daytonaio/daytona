/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, ForbiddenException, Logger, NotFoundException } from '@nestjs/common'
import { ResourceAccessGuard } from '../../common/guards/resource-access.guard'
import { SnapshotService } from '../services/snapshot.service'
import { isBaseAuthContext } from '../../common/interfaces/base-auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { Snapshot } from '../entities/snapshot.entity'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isRegionAuthContext } from '../../common/interfaces/region-auth-context.interface'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'

@Injectable()
export class SnapshotAccessGuard extends ResourceAccessGuard {
  private readonly logger = new Logger(SnapshotAccessGuard.name)

  constructor(private readonly snapshotService: SnapshotService) {
    super()
  }

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
        case isRegionAuthContext(authContext): {
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
          if (snapshot.organizationId !== authContext.organizationId) {
            throw new ForbiddenException('Request organization ID does not match resource organization ID')
          }
          break
        }
        default:
          throw new InvalidAuthenticationContextException()
      }

      // Access granted
      return true
    } catch (error) {
      this.handleResourceAccessError(error, this.logger, `Snapshot with ID or name ${snapshotId} not found`)
    }
  }
}
