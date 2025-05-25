/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, UseGuards, HttpCode } from '@nestjs/common'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiHeader } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { ObjectStorageService } from '../services/object-storage.service'
import { StorageAccessDto } from '../../sandbox/dto/storage-access-dto'
import { CustomHeaders } from '../../common/constants/header.constants'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { AuthContext } from '../../common/decorators/auth-context.decorator'

@ApiTags('object-storage')
@Controller('object-storage')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
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
  async getPushAccess(@AuthContext() authContext: OrganizationAuthContext): Promise<StorageAccessDto> {
    return this.objectStorageService.getPushAccess(authContext.organizationId)
  }
}
