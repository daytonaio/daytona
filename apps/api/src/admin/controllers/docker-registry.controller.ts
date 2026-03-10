/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Param, Post } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiResponse, ApiTags } from '@nestjs/swagger'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { DockerRegistryDto } from '../../docker-registry/dto/docker-registry.dto'
import { Audit } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@ApiTags('admin')
@Controller('admin/docker-registry')
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminDockerRegistryController {
  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  @Post(':id/set-default')
  @ApiOperation({
    summary: 'Set default registry',
    operationId: 'adminSetDefaultRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'ID of the docker registry',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'The docker registry has been set as default.',
    type: DockerRegistryDto,
  })
  @Audit({
    action: AuditAction.SET_DEFAULT,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
  })
  async setDefault(@Param('id') registryId: string): Promise<DockerRegistryDto> {
    const dockerRegistry = await this.dockerRegistryService.setDefault(registryId)
    return DockerRegistryDto.fromDockerRegistry(dockerRegistry)
  }
}
