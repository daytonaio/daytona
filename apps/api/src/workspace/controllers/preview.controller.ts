/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import Redis from 'ioredis'
import { Controller, Get, Param, Logger, NotFoundException, UseGuards, Req } from '@nestjs/common'
import { WorkspaceService } from '../services/workspace.service'
import { ApiResponse, ApiOperation, ApiParam, ApiTags, ApiOAuth2, ApiBearerAuth } from '@nestjs/swagger'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { OrganizationUser } from '../../organization/entities/organization-user.entity'

@ApiTags('preview')
@Controller('preview')
export class PreviewController {
  private readonly logger = new Logger(PreviewController.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly workspaceService: WorkspaceService,
    @InjectRepository(OrganizationUser)
    private readonly organizationUserRepository: Repository<OrganizationUser>,
  ) {}

  @Get(':workspaceId/public')
  @ApiOperation({
    summary: 'Check if workspace is public',
    operationId: 'isWorkspacePublic',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Public status of the workspace',
    type: Boolean,
  })
  async isWorkspacePublic(@Param('workspaceId') workspaceId: string): Promise<boolean> {
    const cached = await this.redis.get(`preview:public:${workspaceId}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    try {
      const isPublic = await this.workspaceService.isWorkspacePublic(workspaceId)
      //  for private workspaces, throw 404 as well
      //  to prevent using the method to check if a workspace exists
      if (!isPublic) {
        //  cache the result for 3 seconds to avoid unnecessary requests to the database
        await this.redis.setex(`preview:public:${workspaceId}`, 3, '0')

        throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
      }
      //  cache the result for 3 seconds to avoid unnecessary requests to the database
      await this.redis.setex(`preview:public:${workspaceId}`, 3, '1')
      return true
    } catch (ex) {
      if (ex instanceof NotFoundException) {
        //  cache the not found workspace as well
        //  as it is the same case as for the private workspaces
        await this.redis.setex(`preview:public:${workspaceId}`, 3, '0')
        throw ex
      }
      throw ex
    }
  }

  @Get(':workspaceId/validate/:authToken')
  @ApiOperation({
    summary: 'Check if workspace auth token is valid',
    operationId: 'isValidAuthToken',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiParam({
    name: 'authToken',
    description: 'Auth token of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace auth token validation status',
    type: Boolean,
  })
  async isValidAuthToken(
    @Param('workspaceId') workspaceId: string,
    @Param('authToken') authToken: string,
  ): Promise<boolean> {
    const cached = await this.redis.get(`preview:token:${workspaceId}:${authToken}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }
    const workspace = await this.workspaceService.findOne(workspaceId)
    if (!workspace) {
      await this.redis.setex(`preview:token:${workspaceId}:${authToken}`, 3, '0')
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }
    if (workspace.authToken === authToken) {
      await this.redis.setex(`preview:token:${workspaceId}:${authToken}`, 3, '1')
      return true
    }
    await this.redis.setex(`preview:token:${workspaceId}:${authToken}`, 3, '0')
    throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
  }

  @Get(':workspaceId/access')
  @ApiOperation({
    summary: 'Check if user has access to the workspace',
    operationId: 'hasWorkspaceAccess',
  })
  @UseGuards(CombinedAuthGuard)
  @ApiOAuth2(['openid', 'profile', 'email'])
  @ApiBearerAuth()
  async hasWorkspaceAccess(@Req() req: Request, @Param('workspaceId') workspaceId: string): Promise<boolean> {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const userId = req.user?.userId

    const cached = await this.redis.get(`preview:access:${workspaceId}:${userId}`)
    if (cached) {
      if (cached === '1') {
        return true
      }
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    const organizationUsers = await this.organizationUserRepository.find({
      where: {
        userId,
      },
    })

    const workspace = await this.workspaceService.findOne(workspaceId)
    const hasAccess = organizationUsers.find((org) => org.organizationId === workspace.organizationId)
    if (!hasAccess) {
      await this.redis.setex(`preview:token:${workspaceId}:${userId}`, 3, '0')
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }
    //  if user has access, keep it in cache longer
    await this.redis.setex(`preview:access:${workspaceId}:${userId}`, 30, '1')
    return true
  }
}
