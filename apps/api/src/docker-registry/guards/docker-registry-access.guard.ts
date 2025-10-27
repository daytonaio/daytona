/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  NotFoundException,
  ForbiddenException,
  Logger,
} from '@nestjs/common'
import { DockerRegistryService } from '../services/docker-registry.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class DockerRegistryAccessGuard implements CanActivate {
  private readonly logger = new Logger(DockerRegistryAccessGuard.name)

  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const registryId: string = request.params.dockerRegistryId || request.params.registryId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const registryOrganizationId = await this.dockerRegistryService.getOrganizationId(registryId)
      if (authContext.role !== SystemRole.ADMIN && registryOrganizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }
  }
}
