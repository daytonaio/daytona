/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, UseGuards, HttpCode } from '@nestjs/common'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { ObjectStorageService } from '../services/object-storage.service'
import { StorageAccessDto } from '../../sandbox/dto/storage-access-dto'
import { CustomHeaders } from '../../common/constants/header.constants'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('object-storage')
@ApiTags('object-storage')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class ObjectStorageController {
  constructor(private readonly objectStorageService: ObjectStorageService) {}

  @Get('push-access')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get temporary storage access for pushing objects',
    operationId: 'getPushAccess',
  })
  @ApiResponse({
    status: 200,
    description: 'Temporary storage access has been generated',
    type: StorageAccessDto,
  })
  async getPushAccess(@IsOrganizationAuthContext() authContext: OrganizationAuthContext): Promise<StorageAccessDto> {
    return this.objectStorageService.getPushAccess(authContext.organizationId)
  }
}
