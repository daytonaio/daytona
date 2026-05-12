/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, ForbiddenException, Logger, NotFoundException } from '@nestjs/common'
import { EntityNotFoundError } from 'typeorm'
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
export class SnapshotReadAccessGuard extends ResourceAccessGuard {
  private readonly logger = new Logger(SnapshotReadAccessGuard.name)

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
        throw new NotFoundException({ code: 'SNAPSHOT_NOT_FOUND', message: `Snapshot with ID ${snapshotId} not found` })
      }

      try {
        snapshot = await this.snapshotService.getSnapshotByName(snapshotId, authContext.organizationId)
      } catch {
        throw new NotFoundException({ code: 'SNAPSHOT_NOT_FOUND', message: `Snapshot with ID or name ${snapshotId} not found` })
      }
    }

    try {
      switch (true) {
        case isRegionAuthContext(authContext): {
          const isAvailable = await this.snapshotService.isAvailableInRegion(snapshot.id, authContext.regionId)
          if (!isAvailable) {
            throw new NotFoundException({ code: 'SNAPSHOT_NOT_FOUND', message: `Snapshot is not available in region ${authContext.regionId}` })
          }
          break
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          break
        case isOrganizationAuthContext(authContext): {
          if (snapshot.organizationId !== authContext.organizationId && !snapshot.general) {
            throw new ForbiddenException({ code: 'SNAPSHOT_ACCESS_DENIED', message: 'Request organization ID does not match resource organization ID' })
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

  // Preserve typed error bodies (code field set by guard) instead of collapsing to plain string.
  protected handleResourceAccessError(error: unknown, logger: Logger, notFoundMessage: string): never {
    if (error instanceof ForbiddenException || error instanceof NotFoundException) {
      const body = error.getResponse()
      if (typeof body === 'object' && body !== null && 'code' in body) {
        throw error // already typed — pass through as-is
      }
    }
    if (!(error instanceof NotFoundException) && !(error instanceof EntityNotFoundError)) {
      logger.error(error)
    }
    throw new NotFoundException({ code: 'SNAPSHOT_NOT_FOUND', message: notFoundMessage })
  }
}
