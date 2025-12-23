/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { DockerRegistryService } from '../services/docker-registry.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RegistryType } from '../enums/registry-type.enum'

@Injectable()
export class DockerRegistryAccessGuard implements CanActivate {
  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const dockerRegistryId: string = request.params.dockerRegistryId || request.params.registryId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const dockerRegistry = await this.dockerRegistryService.findOneOrFail(dockerRegistryId)
      if (authContext.role !== SystemRole.ADMIN && dockerRegistry.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      if (authContext.role !== SystemRole.ADMIN && dockerRegistry.registryType !== RegistryType.ORGANIZATION) {
        // allow access only to registries manually created by the organization
        throw new ForbiddenException(`Requested registry in not type "${RegistryType.ORGANIZATION}"`)
      }
      request.dockerRegistry = dockerRegistry
      return true
    } catch (error) {
      throw new NotFoundException(`Docker registry with ID ${dockerRegistryId} not found`)
    }
  }
}
