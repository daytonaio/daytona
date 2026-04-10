/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, NotFoundException, ForbiddenException, Logger } from '@nestjs/common'
import { ResourceAccessGuard } from '../../common/guards/resource-access.guard'
import { DockerRegistryService } from '../services/docker-registry.service'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { RegistryType } from '../enums/registry-type.enum'

@Injectable()
export class DockerRegistryAccessGuard extends ResourceAccessGuard {
  private readonly logger = new Logger(DockerRegistryAccessGuard.name)

  constructor(private readonly dockerRegistryService: DockerRegistryService) {
    super()
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const dockerRegistryId: string = request.params.dockerRegistryId || request.params.registryId || request.params.id

    const authContext = getAuthContext(context, isOrganizationAuthContext)

    try {
      const dockerRegistry = await this.dockerRegistryService.findOneOrFail(dockerRegistryId)
      if (dockerRegistry.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      if (dockerRegistry.registryType !== RegistryType.ORGANIZATION) {
        // Allow access only to registries manually created by the organization
        throw new ForbiddenException(`Requested registry is not type "${RegistryType.ORGANIZATION}"`)
      }
      request.dockerRegistry = dockerRegistry
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException(`Docker registry with ID ${dockerRegistryId} not found`)
    }
  }
}
