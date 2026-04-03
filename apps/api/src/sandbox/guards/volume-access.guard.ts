/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import {
  Injectable,
  CanActivate,
  ExecutionContext,
  ForbiddenException,
  Logger,
  NotFoundException,
} from '@nestjs/common'
import { isBaseAuthContext } from '../../common/interfaces/base-auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { VolumeService } from '../services/volume.service'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'

@Injectable()
export class VolumeAccessGuard implements CanActivate {
  private readonly logger = new Logger(VolumeAccessGuard.name)

  constructor(private readonly volumeService: VolumeService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()

    const volumeId = request.params.volumeId || request.params.id
    const volumeName = request.params.name

    if (!volumeId && !volumeName) {
      throw new NotFoundException(`Volume not found`)
    }

    const authContext = getAuthContext(context, isBaseAuthContext)

    try {
      switch (true) {
        case isOrganizationAuthContext(authContext): {
          const params = volumeId ? { id: volumeId } : { name: volumeName, organizationId: authContext.organizationId }
          const volumeOrganizationId = await this.volumeService.getOrganizationId(params)
          if (volumeOrganizationId !== authContext.organizationId) {
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
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException(`Volume with ${volumeId ? 'ID' : 'name'} ${volumeId || volumeName} not found`)
    }
  }
}
