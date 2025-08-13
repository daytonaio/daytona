/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { DockerRegistryService } from '../services/docker-registry.service'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RequestWithOrganizationContext } from '../../common/types/request.types'

@Injectable()
export class DockerRegistryAccessGuard implements CanActivate {
  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<RequestWithOrganizationContext>()
    const dockerRegistryId: string = request.params.dockerRegistryId || request.params.registryId || request.params.id

    const authContext = request.user

    try {
      const dockerRegistry = await this.dockerRegistryService.findOneOrFail(dockerRegistryId)
      if (authContext.role !== SystemRole.ADMIN && dockerRegistry.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      request.dockerRegistry = dockerRegistry
      return true
    } catch (error) {
      throw new NotFoundException(`Docker registry with ID ${dockerRegistryId} not found`)
    }
  }
}
