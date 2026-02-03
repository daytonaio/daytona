/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import Redis from 'ioredis'
import { Controller, Get, Param, Logger, NotFoundException, UseGuards, Req } from '@nestjs/common'
import { SandboxService } from '../services/sandbox.service'
import { ApiResponse, ApiOperation, ApiParam, ApiTags, ApiOAuth2, ApiBearerAuth } from '@nestjs/swagger'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrganizationUserService } from '../../organization/services/organization-user.service'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'

@ApiTags('preview')
@Controller('preview')
export class PreviewController {
  private readonly logger = new Logger(PreviewController.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly sandboxService: SandboxService,
    private readonly organizationUserService: OrganizationUserService,
  ) {}

  @Get(':sandboxId/public')
  @ApiOperation({
    summary: 'Check if sandbox is public',
    operationId: 'isSandboxPublic',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Public status of the sandbox',
    type: Boolean,
  })
  async isSandboxPublic(@Param('sandboxId') sandboxId: string): Promise<boolean> {
    const cached = await this.redis.get(`preview:public:${sandboxId}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    try {
      const isPublic = await this.sandboxService.isSandboxPublic(sandboxId)
      //  for private sandboxes, throw 404 as well
      //  to prevent using the method to check if a sandbox exists
      if (!isPublic) {
        //  cache the result for 3 seconds to avoid unnecessary requests to the database
        await this.redis.setex(`preview:public:${sandboxId}`, 3, '0')

        throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
      }
      //  cache the result for 3 seconds to avoid unnecessary requests to the database
      await this.redis.setex(`preview:public:${sandboxId}`, 3, '1')
      return true
    } catch (ex) {
      if (ex instanceof NotFoundException) {
        //  cache the not found sandbox as well
        //  as it is the same case as for the private sandboxes
        await this.redis.setex(`preview:public:${sandboxId}`, 3, '0')
        throw ex
      }
      throw ex
    }
  }

  @Get(':sandboxId/validate/:authToken')
  @ApiOperation({
    summary: 'Check if sandbox auth token is valid',
    operationId: 'isValidAuthToken',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiParam({
    name: 'authToken',
    description: 'Auth token of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox auth token validation status',
    type: Boolean,
  })
  async isValidAuthToken(
    @Param('sandboxId') sandboxId: string,
    @Param('authToken') authToken: string,
  ): Promise<boolean> {
    const cached = await this.redis.get(`preview:token:${sandboxId}:${authToken}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
    const sandbox = await this.sandboxService.findOne(sandboxId)
    if (!sandbox) {
      await this.redis.setex(`preview:token:${sandboxId}:${authToken}`, 3, '0')
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
    if (sandbox.authToken === authToken) {
      await this.redis.setex(`preview:token:${sandboxId}:${authToken}`, 3, '1')
      return true
    }
    await this.redis.setex(`preview:token:${sandboxId}:${authToken}`, 3, '0')
    throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
  }

  @Get(':sandboxId/access')
  @ApiOperation({
    summary: 'Check if user has access to the sandbox',
    operationId: 'hasSandboxAccess',
  })
  @ApiResponse({
    status: 200,
    description: 'User access status to the sandbox',
    type: Boolean,
  })
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard)
  @ApiOAuth2(['openid', 'profile', 'email'])
  @ApiBearerAuth()
  async hasSandboxAccess(@Req() req: Request, @Param('sandboxId') sandboxId: string): Promise<boolean> {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const userId = req.user?.userId

    const cached = await this.redis.get(`preview:access:${sandboxId}:${userId}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    const sandbox = await this.sandboxService.findOne(sandboxId)
    const hasAccess = await this.organizationUserService.exists(sandbox.organizationId, userId)
    if (!hasAccess) {
      await this.redis.setex(`preview:token:${sandboxId}:${userId}`, 3, '0')
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
    //  if user has access, keep it in cache longer
    await this.redis.setex(`preview:access:${sandboxId}:${userId}`, 30, '1')
    return true
  }

  @Get(':signedPreviewToken/:port/sandbox-id')
  @ApiOperation({
    summary: 'Get sandbox ID from signed preview URL token',
    operationId: 'getSandboxIdFromSignedPreviewUrlToken',
  })
  @ApiParam({
    name: 'signedPreviewToken',
    description: 'Signed preview URL token',
    type: 'string',
  })
  @ApiParam({
    name: 'port',
    description: 'Port number to get sandbox ID from signed preview URL token',
    type: 'number',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox ID from signed preview URL token',
    type: String,
  })
  async getSandboxIdFromSignedPreviewUrlToken(
    @Param('signedPreviewToken') signedPreviewToken: string,
    @Param('port') port: number,
  ): Promise<string> {
    return this.sandboxService.getSandboxIdFromSignedPreviewUrlToken(signedPreviewToken, port)
  }
}
