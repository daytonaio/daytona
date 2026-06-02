/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  Get,
  HttpCode,
  NotFoundException,
  Param,
  ParseEnumPipe,
  Patch,
  Post,
  UseGuards,
} from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiResponse, ApiTags } from '@nestjs/swagger'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { OrganizationService } from '../../organization/services/organization.service'
import { CreateOrganizationRegionQuotaDto } from '../../organization/dto/create-organization-region-quota.dto'
import { UpdateOrganizationRegionQuotaDto } from '../../organization/dto/update-organization-region-quota.dto'
import { RegionQuotaDto } from '../../organization/dto/region-quota.dto'

@Controller('admin/organizations')
@ApiTags('admin')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@UseGuards(AuthenticatedRateLimitGuard)
export class AdminOrganizationController {
  constructor(private readonly organizationService: OrganizationService) {}

  @Post(':organizationId/quota/:regionId')
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create organization region quota',
    operationId: 'adminCreateOrganizationRegionQuota',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'regionId',
    description: 'ID of the region the new quota applies to',
    type: 'string',
  })
  @ApiResponse({
    status: 201,
    description: 'Region quota created successfully',
    type: RegionQuotaDto,
  })
  @Audit({
    action: AuditAction.CREATE_REGION_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      params: (req) => ({
        regionId: req.params.regionId,
      }),
      body: (req: TypedRequest<CreateOrganizationRegionQuotaDto>) => ({
        sandboxClass: req.body?.sandboxClass,
        totalCpuQuota: req.body?.totalCpuQuota,
        totalMemoryQuota: req.body?.totalMemoryQuota,
        totalDiskQuota: req.body?.totalDiskQuota,
        totalGpuQuota: req.body?.totalGpuQuota,
        allowedGpuTypes: req.body?.allowedGpuTypes,
        maxCpuPerSandbox: req.body?.maxCpuPerSandbox,
        maxMemoryPerSandbox: req.body?.maxMemoryPerSandbox,
        maxDiskPerSandbox: req.body?.maxDiskPerSandbox,
        maxDiskPerNonEphemeralSandbox: req.body?.maxDiskPerNonEphemeralSandbox,
        maxCpuPerGpuSandbox: req.body?.maxCpuPerGpuSandbox,
        maxMemoryPerGpuSandbox: req.body?.maxMemoryPerGpuSandbox,
        maxDiskPerGpuSandbox: req.body?.maxDiskPerGpuSandbox,
      }),
    },
  })
  async createRegionQuota(
    @Param('organizationId') organizationId: string,
    @Param('regionId') regionId: string,
    @Body() createDto: CreateOrganizationRegionQuotaDto,
  ): Promise<RegionQuotaDto> {
    return this.organizationService.createRegionQuota(organizationId, regionId, createDto)
  }

  @Get(':organizationId/quota/:regionId/:sandboxClass')
  @ApiOperation({
    summary: 'Get organization region quota',
    operationId: 'adminGetOrganizationRegionQuota',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'regionId',
    description: 'Region ID',
    type: 'string',
  })
  @ApiParam({
    name: 'sandboxClass',
    description: 'Sandbox class',
    enum: SandboxClass,
    enumName: 'SandboxClass',
  })
  @ApiResponse({
    status: 200,
    description: 'Region quota',
    type: RegionQuotaDto,
  })
  async getRegionQuota(
    @Param('organizationId') organizationId: string,
    @Param('regionId') regionId: string,
    @Param('sandboxClass', new ParseEnumPipe(SandboxClass)) sandboxClass: SandboxClass,
  ): Promise<RegionQuotaDto> {
    const regionQuota = await this.organizationService.getRegionQuota(organizationId, regionId, sandboxClass)
    if (!regionQuota) {
      throw new NotFoundException(
        `Region quota for organization ${organizationId}, region ${regionId}, sandbox class ${sandboxClass} not found`,
      )
    }
    return regionQuota
  }

  @Patch(':organizationId/quota/:regionId')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Update organization region quota',
    operationId: 'adminUpdateOrganizationRegionQuota',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'regionId',
    description: 'Region ID',
    type: 'string',
  })
  @ApiResponse({
    status: 204,
    description: 'Region quota updated successfully',
  })
  @Audit({
    action: AuditAction.UPDATE_REGION_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      params: (req) => ({
        regionId: req.params.regionId,
      }),
      body: (req: TypedRequest<UpdateOrganizationRegionQuotaDto>) => ({
        sandboxClass: req.body?.sandboxClass,
        totalCpuQuota: req.body?.totalCpuQuota,
        totalMemoryQuota: req.body?.totalMemoryQuota,
        totalDiskQuota: req.body?.totalDiskQuota,
        totalGpuQuota: req.body?.totalGpuQuota,
        allowedGpuTypes: req.body?.allowedGpuTypes,
        maxCpuPerSandbox: req.body?.maxCpuPerSandbox,
        maxMemoryPerSandbox: req.body?.maxMemoryPerSandbox,
        maxDiskPerSandbox: req.body?.maxDiskPerSandbox,
        maxDiskPerNonEphemeralSandbox: req.body?.maxDiskPerNonEphemeralSandbox,
        maxCpuPerGpuSandbox: req.body?.maxCpuPerGpuSandbox,
        maxMemoryPerGpuSandbox: req.body?.maxMemoryPerGpuSandbox,
        maxDiskPerGpuSandbox: req.body?.maxDiskPerGpuSandbox,
      }),
    },
  })
  async updateRegionQuota(
    @Param('organizationId') organizationId: string,
    @Param('regionId') regionId: string,
    @Body() updateDto: UpdateOrganizationRegionQuotaDto,
  ): Promise<void> {
    await this.organizationService.updateRegionQuota(organizationId, regionId, updateDto)
  }

  @Delete(':organizationId/quota/:regionId/:sandboxClass')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Delete organization region quota',
    operationId: 'adminDeleteOrganizationRegionQuota',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'regionId',
    description: 'Region ID',
    type: 'string',
  })
  @ApiParam({
    name: 'sandboxClass',
    description: 'Sandbox class',
    enum: SandboxClass,
    enumName: 'SandboxClass',
  })
  @ApiResponse({
    status: 204,
    description: 'Region quota deleted successfully',
  })
  @Audit({
    action: AuditAction.DELETE_REGION_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      params: (req) => ({
        regionId: req.params.regionId,
        sandboxClass: req.params.sandboxClass,
      }),
    },
  })
  async deleteRegionQuota(
    @Param('organizationId') organizationId: string,
    @Param('regionId') regionId: string,
    @Param('sandboxClass', new ParseEnumPipe(SandboxClass)) sandboxClass: SandboxClass,
  ): Promise<void> {
    await this.organizationService.deleteRegionQuota(organizationId, regionId, sandboxClass)
  }
}
