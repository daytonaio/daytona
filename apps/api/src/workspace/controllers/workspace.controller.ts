/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Get,
  Post,
  Delete,
  Body,
  Param,
  Query,
  Logger,
  UseGuards,
  HttpCode,
  UseInterceptors,
  Put,
  NotFoundException,
  ForbiddenException,
  Res,
  Request,
  RawBodyRequest,
  Next,
  ParseBoolPipe,
} from '@nestjs/common'
import Redis from 'ioredis'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { WorkspaceService } from '../services/workspace.service'
import { CreateWorkspaceDto } from '../dto/create-workspace.dto'
import {
  ApiOAuth2,
  ApiResponse,
  ApiQuery,
  ApiOperation,
  ApiParam,
  ApiTags,
  ApiHeader,
  ApiBearerAuth,
} from '@nestjs/swagger'
import { WorkspaceDto, WorkspaceLabelsDto } from '../dto/workspace.dto'
import { RunnerService } from '../services/runner.service'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { Workspace as WorkspaceEntity } from '../entities/workspace.entity'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { Throttle } from '@nestjs/throttler'
import { Runner } from '../entities/runner.entity'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Workspace } from '../decorators/workspace.decorator'
import { WorkspaceAccessGuard } from '../guards/workspace-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { PortPreviewUrlDto } from '../dto/port-preview-url.dto'
import { IncomingMessage, ServerResponse } from 'http'
import { NextFunction } from 'http-proxy-middleware/dist/types'
import { LogProxy } from '../proxy/log-proxy'

@ApiTags('workspace')
@Controller('workspace')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class WorkspaceController {
  private readonly logger = new Logger(WorkspaceController.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly runnerService: RunnerService,
    private readonly workspaceService: WorkspaceService,
  ) {}

  @Get()
  @ApiOperation({
    summary: 'List all workspaces',
    operationId: 'listWorkspaces',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all workspaces',
    type: [WorkspaceDto],
  })
  @ApiQuery({
    name: 'verbose',
    required: false,
    type: Boolean,
    description: 'Include verbose output',
  })
  @ApiQuery({
    name: 'labels',
    type: String,
    required: false,
    example: '{"label1": "value1", "label2": "value2"}',
    description: 'JSON encoded labels to filter by',
  })
  async listWorkspaces(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('verbose') verbose?: boolean,
    @Query('labels') labelsQuery?: string,
  ): Promise<WorkspaceDto[]> {
    const labels = labelsQuery ? JSON.parse(labelsQuery) : {}
    const workspaces = await this.workspaceService.findAll(authContext.organizationId, labels)
    const dtos = workspaces.map(async (workspace) => {
      const runner = await this.runnerService.findOne(workspace.runnerId)
      const dto = WorkspaceDto.fromWorkspace(workspace, runner.domain)
      return dto
    })
    return await Promise.all(dtos)
  }

  @Post()
  @HttpCode(200) //  for Daytona Api compatibility
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create a new workspace',
    operationId: 'createWorkspace',
  })
  @ApiResponse({
    status: 200,
    description: 'The workspace has been successfully created.',
    type: WorkspaceDto,
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  async createWorkspace(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createWorkspaceDto: CreateWorkspaceDto,
  ): Promise<WorkspaceDto> {
    const organization = authContext.organization

    const workspace = await this.workspaceService.create(organization.id, createWorkspaceDto, organization)

    // If the workspace has no runner, it means it is still building - return the ID to the client so they can fetch logs
    if (workspace.runnerId) {
      // Wait for workspace to be started
      const workspaceState = await this.waitForWorkspaceState(
        workspace.id,
        WorkspaceState.STARTED,
        30000, // 30 seconds timeout
      )

      workspace.state = workspaceState
    }

    const runner = await this.runnerService.findOne(workspace.runnerId)
    const dto = WorkspaceDto.fromWorkspace(workspace, runner.domain)
    return dto
  }

  @Get(':workspaceId')
  @ApiOperation({
    summary: 'Get workspace details',
    operationId: 'getWorkspace',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiQuery({
    name: 'verbose',
    required: false,
    type: Boolean,
    description: 'Include verbose output',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace details',
    type: WorkspaceDto,
  })
  @UseGuards(WorkspaceAccessGuard)
  async getWorkspace(
    @Workspace() workspace: WorkspaceEntity,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    @Query('verbose') verbose?: boolean,
  ): Promise<WorkspaceDto> {
    let runner: Runner
    if (workspace.runnerId) {
      runner = await this.runnerService.findOne(workspace.runnerId)
    }

    return WorkspaceDto.fromWorkspace(workspace, runner?.domain)
  }

  @Delete(':workspaceId')
  @ApiOperation({
    summary: 'Delete workspace',
    operationId: 'deleteWorkspace',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace has been deleted',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async removeWorkspace(
    @Param('workspaceId') workspaceId: string,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    @Query('force') force?: boolean,
  ): Promise<void> {
    return this.workspaceService.destroy(workspaceId)
  }

  @Post(':workspaceId/start')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Start workspace',
    operationId: 'startWorkspace',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace has been started',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async startWorkspace(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('workspaceId') workspaceId: string,
  ): Promise<void> {
    return this.workspaceService.start(workspaceId, authContext.organization)
  }

  @Post(':workspaceId/stop')
  @HttpCode(200) //  for Daytona Api compatibility
  @ApiOperation({
    summary: 'Stop workspace',
    operationId: 'stopWorkspace',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace has been stopped',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async stopWorkspace(@Param('workspaceId') workspaceId: string): Promise<void> {
    return this.workspaceService.stop(workspaceId)
  }

  @Put(':workspaceId/labels')
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Replace workspace labels',
    operationId: 'replaceLabels',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Labels have been successfully replaced',
    type: WorkspaceLabelsDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async replaceLabels(
    @Param('workspaceId') workspaceId: string,
    @Body() labelsDto: WorkspaceLabelsDto,
  ): Promise<WorkspaceLabelsDto> {
    const labels = await this.workspaceService.replaceLabels(workspaceId, labelsDto.labels)
    return { labels }
  }

  @Post(':workspaceId/backup')
  @ApiOperation({
    summary: 'Create workspace backup',
    operationId: 'createBackup',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace backup has been initiated',
    type: WorkspaceDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async createBackup(@Param('workspaceId') workspaceId: string): Promise<void> {
    await this.workspaceService.createBackup(workspaceId)
  }

  @Post(':workspaceId/public/:isPublic')
  @ApiOperation({
    summary: 'Update public status',
    operationId: 'updatePublicStatus',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiParam({
    name: 'isPublic',
    description: 'Public status to set',
    type: 'boolean',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async updatePublicStatus(
    @Param('workspaceId') workspaceId: string,
    @Param('isPublic') isPublic: boolean,
  ): Promise<void> {
    await this.workspaceService.updatePublicStatus(workspaceId, isPublic)
  }

  @Post(':workspaceId/autostop/:interval')
  @ApiOperation({
    summary: 'Set workspace auto-stop interval',
    operationId: 'setAutostopInterval',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiParam({
    name: 'interval',
    description: 'Auto-stop interval in minutes (0 to disable)',
    type: 'number',
  })
  @ApiResponse({
    status: 200,
    description: 'Auto-stop interval has been set',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async setAutostopInterval(
    @Param('workspaceId') workspaceId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.workspaceService.setAutostopInterval(workspaceId, interval)
  }

  @Post(':workspaceId/autoarchive/:interval')
  @ApiOperation({
    summary: 'Set workspace auto-archive interval',
    operationId: 'setAutoArchiveInterval',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiParam({
    name: 'interval',
    description: 'Auto-archive interval in minutes (0 means the maximum interval will be used)',
    type: 'number',
  })
  @ApiResponse({
    status: 200,
    description: 'Auto-archive interval has been set',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async setAutoArchiveInterval(
    @Param('workspaceId') workspaceId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.workspaceService.setAutoArchiveInterval(workspaceId, interval)
  }

  @Post(':workspaceId/archive')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Archive workspace',
    operationId: 'archiveWorkspace',
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace has been archived',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  async archiveWorkspace(@Param('workspaceId') workspaceId: string): Promise<void> {
    return this.workspaceService.archive(workspaceId)
  }

  @Get(':workspaceId/ports/:port/preview-url')
  @ApiOperation({
    summary: 'Get preview URL for a workspace port',
    operationId: 'getPortPreviewUrl',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiParam({
    name: 'port',
    description: 'Port number to get preview URL for',
    type: 'number',
  })
  @ApiResponse({
    status: 200,
    description: 'Preview URL for the specified port',
    type: PortPreviewUrlDto,
  })
  @UseGuards(WorkspaceAccessGuard)
  async getPortPreviewUrl(
    @Param('workspaceId') workspaceId: string,
    @Param('port') port: number,
  ): Promise<PortPreviewUrlDto> {
    return this.workspaceService.getPortPreviewUrl(workspaceId, port)
  }

  @Get(':workspaceId/build-logs')
  @ApiOperation({
    summary: 'Get build logs',
    operationId: 'getBuildLogs',
  })
  @ApiParam({
    name: 'workspaceId',
    description: 'ID of the workspace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Build logs stream',
  })
  @ApiQuery({
    name: 'follow',
    required: false,
    type: Boolean,
    description: 'Whether to follow the logs stream',
  })
  @UseGuards(WorkspaceAccessGuard)
  async getBuildLogs(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
    @Param('workspaceId') workspaceId: string,
    @Query('follow', new ParseBoolPipe({ optional: true })) follow?: boolean,
  ): Promise<void> {
    const workspace = await this.workspaceService.findOne(workspaceId)
    if (!workspace || !workspace.runnerId) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found or has no runner assigned`)
    }

    if (!workspace.buildInfo) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} has no build info`)
    }

    const runner = await this.runnerService.findOne(workspace.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner for workspace ${workspaceId} not found`)
    }

    const logProxy = new LogProxy(
      runner.apiUrl,
      workspace.buildInfo.snapshotRef.split(':')[0],
      runner.apiKey,
      follow === true,
      req,
      res,
      next,
    )
    return logProxy.create()
  }

  private async waitForWorkspaceState(
    workspaceId: string,
    desiredState: WorkspaceState,
    timeout: number,
  ): Promise<WorkspaceState> {
    const startTime = Date.now()

    let workspaceState: WorkspaceState
    while (Date.now() - startTime < timeout) {
      const workspace = await this.workspaceService.findOne(workspaceId)
      workspaceState = workspace.state
      if (workspaceState === desiredState || workspaceState === WorkspaceState.ERROR) {
        return workspaceState
      }
      await new Promise((resolve) => setTimeout(resolve, 100)) // Wait 100 ms before checking again
    }

    return workspaceState
  }
}
