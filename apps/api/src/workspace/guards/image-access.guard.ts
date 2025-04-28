/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { ImageService } from '../services/image.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class ImageAccessGuard implements CanActivate {
  constructor(private readonly imageService: ImageService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const imageId: string = request.params.imageId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const image = await this.imageService.getImage(imageId)
      if (authContext.role !== SystemRole.ADMIN && image.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      request.image = image
      return true
    } catch (error) {
      throw new NotFoundException(`Image with ID ${imageId} not found`)
    }
  }
}
