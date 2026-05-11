/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Delete, Get, HttpCode, Param, ParseUUIDPipe, Post, Query, UseGuards } from '@nestjs/common'
import { ApiBearerAuth, ApiHeader, ApiOAuth2, ApiOperation, ApiQuery, ApiResponse, ApiTags } from '@nestjs/swagger'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { SessionService } from '../services/session.service'
import { SessionCodeRunRequestDto, SessionCodeRunResponseDto } from '../dto/code-run.dto'
import { SessionConnectRequestDto, SessionConnectResponseDto } from '../dto/connect.dto'
import { SessionAccessDto, SessionDto } from '../dto/session.dto'
import { SessionPackageDto } from '../dto/session-package.dto'
import { SessionTemplateDto } from '../dto/session-template.dto'
import { CreateSessionDto } from '../dto/create-session.dto'
import { CreateSessionTransientDto } from '../dto/create-session-transient.dto'

@Controller('sessions')
@ApiTags('sessions')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class SessionController {
  constructor(private readonly session: SessionService) {}

  @Post('code-run')
  @HttpCode(200)
  @ApiOperation({ summary: 'Run code synchronously and return aggregated stdout/stderr/displays.' })
  @ApiResponse({ status: 200, type: SessionCodeRunResponseDto })
  async codeRun(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Body() body: SessionCodeRunRequestDto,
  ): Promise<SessionCodeRunResponseDto> {
    return this.session.codeRun(ctx.organizationId, ctx.organization, body)
  }

  @Post('connect')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Mint a signed WebSocket URL for streaming code execution against a session.',
  })
  @ApiResponse({ status: 200, type: SessionConnectResponseDto })
  async connect(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Body() body: SessionConnectRequestDto,
  ): Promise<SessionConnectResponseDto> {
    return this.session.connect(ctx.organizationId, ctx.organization, body)
  }

  @Post('transients')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Mint a deterministic one-shot session handle with a direct-to-sandbox access bundle.',
    description:
      'Returns the SDK a stable (template, language) transient session bound to a warm-pool sandbox so subsequent run() calls bypass the API on the hot path. Idempotent — the same (template, language) pair returns the same daemon-side session until the warm instance is recycled.',
  })
  @ApiResponse({ status: 200, type: SessionDto })
  async createTransient(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Body() body: CreateSessionTransientDto,
  ): Promise<SessionDto> {
    return this.session.createTransientSession(ctx.organizationId, ctx.organization, body)
  }

  @Get('templates')
  @ApiOperation({ summary: 'List available session templates for the current organization.' })
  @ApiResponse({ status: 200, type: [SessionTemplateDto] })
  async listTemplates(@IsOrganizationAuthContext() ctx: OrganizationAuthContext): Promise<SessionTemplateDto[]> {
    return this.session.listTemplates(ctx.organizationId)
  }

  @Get('templates/:name/packages')
  @ApiOperation({ summary: 'List preinstalled packages for a template.' })
  @ApiQuery({ name: 'language', required: true })
  @ApiResponse({ status: 200, type: [SessionPackageDto] })
  async listPackages(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Param('name') name: string,
    @Query('language') language: string,
  ): Promise<SessionPackageDto[]> {
    return this.session.listPackages(ctx.organizationId, ctx.organization, name, language)
  }

  @Post()
  @HttpCode(200)
  @ApiOperation({ summary: 'Create a persistent execution session.' })
  @ApiResponse({ status: 200, type: SessionDto })
  async createSession(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Body() body: CreateSessionDto,
  ): Promise<SessionDto> {
    return this.session.createSession(ctx.organizationId, ctx.organization, body)
  }

  @Get()
  @ApiOperation({ summary: 'List active sessions in the current organization.' })
  @ApiQuery({ name: 'template', required: false })
  @ApiResponse({ status: 200, type: [SessionDto] })
  async listSessions(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Query('template') template?: string,
  ): Promise<SessionDto[]> {
    return this.session.listSessions(ctx.organizationId, template)
  }

  @Get(':id/access')
  @ApiOperation({
    summary: 'Mint or refresh the SDK direct-to-sandbox access bundle for a session.',
    description:
      "Used by the SDK to refresh the short-lived signed proxy URL before it expires. Also bumps the session's lastUsedAt as a keep-alive.",
  })
  @ApiResponse({ status: 200, type: SessionAccessDto })
  async getSessionAccess(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    // Plain string (no ParseUUIDPipe): transient/invalid ids are valid handles here and must
    // reach the service to surface a proper 404/410, not be rejected up front with a 400.
    @Param('id') id: string,
  ): Promise<SessionAccessDto> {
    return this.session.getSessionAccess(ctx.organizationId, id)
  }

  @Delete(':id')
  @HttpCode(204)
  @ApiOperation({ summary: 'Delete a session.' })
  @ApiResponse({ status: 204 })
  async deleteSession(
    @IsOrganizationAuthContext() ctx: OrganizationAuthContext,
    @Param('id', ParseUUIDPipe) id: string,
  ): Promise<void> {
    await this.session.deleteSession(ctx.organizationId, id)
  }
}
