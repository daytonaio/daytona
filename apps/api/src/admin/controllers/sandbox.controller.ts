/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, HttpCode, NotFoundException, Param, Post, UseGuards } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiResponse, ApiTags } from '@nestjs/swagger'
import { Audit } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { OrganizationService } from '../../organization/services/organization.service'
import { SandboxDto } from '../../sandbox/dto/sandbox.dto'
import { SandboxService } from '../../sandbox/services/sandbox.service'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('admin')
@Controller('admin/sandbox')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@RequiredApiRole([SystemRole.ADMIN])
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminSandboxController {
  constructor(
    private readonly sandboxService: SandboxService,
    private readonly organizationService: OrganizationService,
  ) {}

  @Post(':sandboxId/recover')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Recover sandbox from error state as an admin',
    operationId: 'adminRecoverSandbox',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Recovery initiated',
    type: SandboxDto,
  })
  @Audit({
    action: AuditAction.RECOVER,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async recoverSandbox(@Param('sandboxId') sandboxId: string): Promise<SandboxDto> {
    const organization = await this.organizationService.findBySandboxId(sandboxId)
    if (!organization) {
      throw new NotFoundException('Sandbox not found')
    }
    const recoveredSandbox = await this.sandboxService.recover(sandboxId, organization)
    return SandboxDto.fromSandbox(recoveredSandbox)
  }
}
