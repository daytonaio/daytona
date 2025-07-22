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
import { SandboxService as WorkspaceService } from '../services/sandbox.service'
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
import { SandboxLabelsDto as WorkspaceLabelsDto } from '../dto/sandbox.dto'
import { WorkspaceDto } from '../dto/workspace.deprecated.dto'
import { RunnerService } from '../services/runner.service'
import { SandboxState as WorkspaceState } from '../enums/sandbox-state.enum'
import { Sandbox as WorkspaceEntity } from '../entities/sandbox.entity'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { Throttle } from '@nestjs/throttler'
import { Runner } from '../entities/runner.entity'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Sandbox as Workspace } from '../decorators/sandbox.decorator'
import { SandboxAccessGuard as WorkspaceAccessGuard } from '../guards/sandbox-access.guard'
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
import { CreateWorkspaceDto } from '../dto/create-workspace.deprecated.dto'
import { TypedConfigService } from '../../config/typed-config.service'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

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
    private readonly configService: TypedConfigService,
  ) {}

  @Get()
  @ApiOperation({
    summary: '[DEPRECATED] List all workspaces',
    operationId: 'listWorkspaces_deprecated',
    deprecated: true,
  })
  @ApiResponse({
    status: 200,
    description: 'List of all workspacees',
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
  async listWorkspacees(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('verbose') verbose?: boolean,
    @Query('labels') labelsQuery?: string,
  ): Promise<WorkspaceDto[]> {
    const labels = labelsQuery ? JSON.parse(labelsQuery) : {}
    const workspacees = await this.workspaceService.findAll(authContext.organizationId, labels)
    const dtos = workspacees.map(async (workspace) => {
      const runner = await this.runnerService.findOne(workspace.runnerId)
      const dto = WorkspaceDto.fromSandbox(workspace, runner.domain)
      return dto
    })
    return await Promise.all(dtos)
  }

  @Post()
  @HttpCode(200) //  for Daytona Api compatibility
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: '[DEPRECATED] Create a new workspace',
    operationId: 'createWorkspace_deprecated',
    deprecated: true,
  })
  @ApiResponse({
    status: 200,
    description: 'The workspace has been successfully created.',
    type: WorkspaceDto,
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromResult: (result: WorkspaceDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateWorkspaceDto>) => ({
        image: req.body?.image,
        user: req.body?.user,
        env: req.body?.env
          ? Object.fromEntries(Object.keys(req.body?.env).map((key) => [key, MASKED_AUDIT_VALUE]))
          : undefined,
        labels: req.body?.labels,
        public: req.body?.public,
        class: req.body?.class,
        target: req.body?.target,
        cpu: req.body?.cpu,
        gpu: req.body?.gpu,
        memory: req.body?.memory,
        disk: req.body?.disk,
        autoStopInterval: req.body?.autoStopInterval,
        autoArchiveInterval: req.body?.autoArchiveInterval,
        volumes: req.body?.volumes,
        buildInfo: req.body?.buildInfo,
      }),
    },
  })
  async createWorkspace(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createWorkspaceDto: CreateWorkspaceDto,
  ): Promise<WorkspaceDto> {
    if (createWorkspaceDto.buildInfo) {
      throw new ForbiddenException('Build info is not supported in this deprecated API - please upgrade your client')
    }

    const organization = authContext.organization

    const workspace = WorkspaceDto.fromSandboxDto(
      await this.workspaceService.createFromSnapshot(
        {
          ...createWorkspaceDto,
          snapshot: createWorkspaceDto.image,
        },
        organization,
        true,
      ),
    )

    // Wait for the workspace to start
    const sandboxState = await this.waitForWorkspaceState(
      workspace.id,
      WorkspaceState.STARTED,
      30000, // 30 seconds timeout
    )
    workspace.state = sandboxState

    return workspace
  }

  @Get(':workspaceId')
  @ApiOperation({
    summary: '[DEPRECATED] Get workspace details',
    operationId: 'getWorkspace_deprecated',
    deprecated: true,
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

    return WorkspaceDto.fromSandbox(workspace, runner?.domain)
  }

  @Delete(':workspaceId')
  @ApiOperation({
    summary: '[DEPRECATED] Delete workspace',
    operationId: 'deleteWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
  })
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
    summary: '[DEPRECATED] Start workspace',
    operationId: 'startWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.START,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
  })
  async startWorkspace(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('workspaceId') workspaceId: string,
  ): Promise<void> {
    return this.workspaceService.start(workspaceId, authContext.organization)
  }

  @Post(':workspaceId/stop')
  @HttpCode(200) //  for Daytona Api compatibility
  @ApiOperation({
    summary: '[DEPRECATED] Stop workspace',
    operationId: 'stopWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.STOP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
  })
  async stopWorkspace(@Param('workspaceId') workspaceId: string): Promise<void> {
    return this.workspaceService.stop(workspaceId)
  }

  @Put(':workspaceId/labels')
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: '[DEPRECATED] Replace workspace labels',
    operationId: 'replaceLabelsWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.REPLACE_LABELS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
    requestMetadata: {
      body: (req: TypedRequest<WorkspaceLabelsDto>) => ({
        labels: req.body?.labels,
      }),
    },
  })
  async replaceLabels(
    @Param('workspaceId') workspaceId: string,
    @Body() labelsDto: WorkspaceLabelsDto,
  ): Promise<WorkspaceLabelsDto> {
    const labels = await this.workspaceService.replaceLabels(workspaceId, labelsDto.labels)
    return { labels }
  }

  @Post(':workspaceId/backup')
  @ApiOperation({
    summary: '[DEPRECATED] Create workspace backup',
    operationId: 'createBackupWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.CREATE_BACKUP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
  })
  async createBackup(@Param('workspaceId') workspaceId: string): Promise<void> {
    await this.workspaceService.createBackup(workspaceId)
  }

  @Post(':workspaceId/public/:isPublic')
  @ApiOperation({
    summary: '[DEPRECATED] Update public status',
    operationId: 'updatePublicStatusWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.UPDATE_PUBLIC_STATUS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
    requestMetadata: {
      params: (req) => ({
        isPublic: req.params.isPublic,
      }),
    },
  })
  async updatePublicStatus(
    @Param('workspaceId') workspaceId: string,
    @Param('isPublic') isPublic: boolean,
  ): Promise<void> {
    await this.workspaceService.updatePublicStatus(workspaceId, isPublic)
  }

  @Post(':workspaceId/autostop/:interval')
  @ApiOperation({
    summary: '[DEPRECATED] Set workspace auto-stop interval',
    operationId: 'setAutostopIntervalWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.SET_AUTO_STOP_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutostopInterval(
    @Param('workspaceId') workspaceId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.workspaceService.setAutostopInterval(workspaceId, interval)
  }

  @Post(':workspaceId/autoarchive/:interval')
  @ApiOperation({
    summary: '[DEPRECATED] Set workspace auto-archive interval',
    operationId: 'setAutoArchiveIntervalWorkspace_deprecated',
    deprecated: true,
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
  @Audit({
    action: AuditAction.SET_AUTO_ARCHIVE_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutoArchiveInterval(
    @Param('workspaceId') workspaceId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.workspaceService.setAutoArchiveInterval(workspaceId, interval)
  }

  @Post(':workspaceId/archive')
  @HttpCode(200)
  @ApiOperation({
    summary: '[DEPRECATED] Archive workspace',
    operationId: 'archiveWorkspace_deprecated',
    deprecated: true,
  })
  @ApiResponse({
    status: 200,
    description: 'Workspace has been archived',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(WorkspaceAccessGuard)
  @Audit({
    action: AuditAction.ARCHIVE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.workspaceId,
  })
  async archiveWorkspace(@Param('workspaceId') workspaceId: string): Promise<void> {
    return this.workspaceService.archive(workspaceId)
  }

  @Get(':workspaceId/ports/:port/preview-url')
  @ApiOperation({
    summary: '[DEPRECATED] Get preview URL for a workspace port',
    operationId: 'getPortPreviewUrlWorkspace_deprecated',
    deprecated: true,
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
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    const proxyDomain = this.configService.get('proxy.domain')
    const proxyProtocol = this.configService.get('proxy.protocol')
    if (proxyDomain && proxyProtocol) {
      const workspace = await this.workspaceService.findOne(workspaceId)
      if (!workspace) {
        throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
      }

      // Return new preview url only for updated workspaces/sandboxes
      if (workspace.daemonVersion) {
        return {
          url: `${proxyProtocol}://${port}-${workspaceId}.${proxyDomain}`,
          token: workspace.authToken,
        }
      }
    }

    return this.workspaceService.getPortPreviewUrl(workspaceId, port)
  }

  @Get(':workspaceId/build-logs')
  @ApiOperation({
    summary: '[DEPRECATED] Get build logs',
    operationId: 'getBuildLogsWorkspace_deprecated',
    deprecated: true,
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
      if (
        workspaceState === desiredState ||
        workspaceState === WorkspaceState.ERROR ||
        workspaceState === WorkspaceState.BUILD_FAILED
      ) {
        return workspaceState
      }
      await new Promise((resolve) => setTimeout(resolve, 100)) // Wait 100 ms before checking again
    }

    return workspaceState
  }
}
