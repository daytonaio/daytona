/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Post, UseGuards, Request as Req } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiOAuth2 } from '@nestjs/swagger'
import { Request } from 'express'
import { AuditLogDto } from '../dto/audit-log.dto'
import { CreateAuditLogDto } from '../dto/create-audit-log.dto'
import { AuditOutcome } from '../enums/audit-outcome-enum'
import { AuditService } from '../services/audit.service'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { RequiredSystemRole } from '../../common/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('audit')
@Controller('audit')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AuditController {
  constructor(private readonly auditService: AuditService) {}

  @Post()
  @ApiOperation({
    summary: 'Create audit log entry',
    operationId: 'createAuditLog',
  })
  @ApiResponse({
    status: 201,
    description: 'Audit log entry created successfully',
    type: AuditLogDto,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async createLog(@Req() req: Request, @Body() createAuditLogDto: CreateAuditLogDto): Promise<AuditLogDto> {
    const auditLog = await this.auditService.createLog({
      actorId: createAuditLogDto.actorId,
      actorEmail: createAuditLogDto.actorEmail,
      organizationId: createAuditLogDto.organizationId,
      action: createAuditLogDto.action,
      targetType: createAuditLogDto.targetType,
      targetId: createAuditLogDto.targetId,
      ipAddress: req.ip,
      userAgent: req.get('user-agent'),
      source: req.get(CustomHeaders.SOURCE.name),
      outcome: AuditOutcome.UNKNOWN,
    })
    return AuditLogDto.fromAuditLog(auditLog)
  }
}
