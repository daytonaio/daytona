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
  Res,
  Request,
  RawBodyRequest,
  Next,
  ParseBoolPipe,
} from '@nestjs/common'
import Redis from 'ioredis'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SandboxService } from '../services/sandbox.service'
import { CreateSandboxDto } from '../dto/create-sandbox.dto'
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
import { SandboxDto, SandboxLabelsDto } from '../dto/sandbox.dto'
import { RunnerService } from '../services/runner.service'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Sandbox as SandboxEntity } from '../entities/sandbox.entity'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { Throttle } from '@nestjs/throttler'
import { Runner } from '../entities/runner.entity'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Sandbox } from '../decorators/sandbox.decorator'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
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
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { TypedConfigService } from '../../config/typed-config.service'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('sandbox')
@Controller('sandbox')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class SandboxController {
  private readonly logger = new Logger(SandboxController.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly runnerService: RunnerService,
    private readonly sandboxService: SandboxService,
    private readonly configService: TypedConfigService,
    private readonly eventEmitter: EventEmitter2,
  ) {}

  @Get()
  @ApiOperation({
    summary: 'List all sandboxes',
    operationId: 'listSandboxes',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all sandboxes',
    type: [SandboxDto],
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
  @ApiQuery({
    name: 'includeErroredDeleted',
    required: false,
    type: Boolean,
    description: 'Include errored and deleted sandboxes',
  })
  async listSandboxes(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('verbose') verbose?: boolean,
    @Query('labels') labelsQuery?: string,
    @Query('includeErroredDeleted') includeErroredDeleted?: boolean,
  ): Promise<SandboxDto[]> {
    const labels = labelsQuery ? JSON.parse(labelsQuery) : {}
    const sandboxes = await this.sandboxService.findAll(authContext.organizationId, labels, includeErroredDeleted)

    const runnerIds = new Set(sandboxes.map((s) => s.runnerId))
    const runners = await this.runnerService.findByIds(Array.from(runnerIds))
    const runnerMap = new Map(runners.map((runner) => [runner.id, runner]))

    return sandboxes.map((sandbox) => {
      const runner = runnerMap.get(sandbox.runnerId)
      return SandboxDto.fromSandbox(sandbox, runner?.domain)
    })
  }

  @Post()
  @HttpCode(200) //  for Daytona Api compatibility
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create a new sandbox',
    operationId: 'createSandbox',
  })
  @ApiResponse({
    status: 200,
    description: 'The sandbox has been successfully created.',
    type: SandboxDto,
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateSandboxDto>) => ({
        snapshot: req.body?.snapshot,
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
        autoDeleteInterval: req.body?.autoDeleteInterval,
        volumes: req.body?.volumes,
        buildInfo: req.body?.buildInfo,
      }),
    },
  })
  async createSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createSandboxDto: CreateSandboxDto,
  ): Promise<SandboxDto> {
    const organization = authContext.organization
    let sandbox: SandboxDto

    if (createSandboxDto.buildInfo) {
      if (createSandboxDto.snapshot) {
        throw new BadRequestError('Cannot specify a snapshot when using a build info entry')
      }
      sandbox = await this.sandboxService.createFromBuildInfo(createSandboxDto, organization)
    } else {
      if (createSandboxDto.cpu || createSandboxDto.gpu || createSandboxDto.memory || createSandboxDto.disk) {
        throw new BadRequestError('Cannot specify Sandbox resources when using a snapshot')
      }
      sandbox = await this.sandboxService.createFromSnapshot(createSandboxDto, organization)
      if (sandbox.state === SandboxState.STARTED) {
        return sandbox
      }

      await this.waitForSandboxStarted(sandbox.id, 30)
      sandbox.state = SandboxState.STARTED
    }

    return sandbox
  }

  @Get(':sandboxId')
  @ApiOperation({
    summary: 'Get sandbox details',
    operationId: 'getSandbox',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
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
    description: 'Sandbox details',
    type: SandboxDto,
  })
  @UseGuards(SandboxAccessGuard)
  async getSandbox(
    @Sandbox() sandbox: SandboxEntity,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    @Query('verbose') verbose?: boolean,
  ): Promise<SandboxDto> {
    let runner: Runner
    if (sandbox.runnerId) {
      runner = await this.runnerService.findOne(sandbox.runnerId)
    }

    return SandboxDto.fromSandbox(sandbox, runner?.domain)
  }

  @Delete(':sandboxId')
  @ApiOperation({
    summary: 'Delete sandbox',
    operationId: 'deleteSandbox',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been deleted',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async deleteSandbox(
    @Param('sandboxId') sandboxId: string,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    @Query('force') force?: boolean,
  ): Promise<void> {
    return this.sandboxService.destroy(sandboxId)
  }

  @Post(':sandboxId/start')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Start sandbox',
    operationId: 'startSandbox',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been started',
    type: SandboxDto,
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.START,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async startSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxId') sandboxId: string,
  ): Promise<SandboxDto> {
    await this.sandboxService.start(sandboxId, authContext.organization)

    const sandbox = await this.waitForSandboxStarted(sandboxId, 30)

    if (!sandbox.runnerDomain) {
      const runner = await this.runnerService.findBySandboxId(sandboxId)
      if (!runner) {
        throw new NotFoundException(`Runner for sandbox ${sandboxId} not found`)
      }
      sandbox.runnerDomain = runner.domain
    }

    return sandbox
  }

  @Post(':sandboxId/stop')
  @HttpCode(200) //  for Daytona Api compatibility
  @ApiOperation({
    summary: 'Stop sandbox',
    operationId: 'stopSandbox',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been stopped',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.STOP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async stopSandbox(@Param('sandboxId') sandboxId: string): Promise<void> {
    return this.sandboxService.stop(sandboxId)
  }

  @Put(':sandboxId/labels')
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Replace sandbox labels',
    operationId: 'replaceLabels',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Labels have been successfully replaced',
    type: SandboxLabelsDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.REPLACE_LABELS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<SandboxLabelsDto>) => ({
        labels: req.body?.labels,
      }),
    },
  })
  async replaceLabels(
    @Param('sandboxId') sandboxId: string,
    @Body() labelsDto: SandboxLabelsDto,
  ): Promise<SandboxLabelsDto> {
    const labels = await this.sandboxService.replaceLabels(sandboxId, labelsDto.labels)
    return { labels }
  }

  @Post(':sandboxId/backup')
  @ApiOperation({
    summary: 'Create sandbox backup',
    operationId: 'createBackup',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox backup has been initiated',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.CREATE_BACKUP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async createBackup(@Param('sandboxId') sandboxId: string): Promise<void> {
    await this.sandboxService.createBackup(sandboxId)
  }

  @Post(':sandboxId/public/:isPublic')
  @ApiOperation({
    summary: 'Update public status',
    operationId: 'updatePublicStatus',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiParam({
    name: 'isPublic',
    description: 'Public status to set',
    type: 'boolean',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.UPDATE_PUBLIC_STATUS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        isPublic: req.params.isPublic,
      }),
    },
  })
  async updatePublicStatus(@Param('sandboxId') sandboxId: string, @Param('isPublic') isPublic: boolean): Promise<void> {
    await this.sandboxService.updatePublicStatus(sandboxId, isPublic)
  }

  @Post(':sandboxId/autostop/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-stop interval',
    operationId: 'setAutostopInterval',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
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
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_STOP_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutostopInterval(@Param('sandboxId') sandboxId: string, @Param('interval') interval: number): Promise<void> {
    await this.sandboxService.setAutostopInterval(sandboxId, interval)
  }

  @Post(':sandboxId/autoarchive/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-archive interval',
    operationId: 'setAutoArchiveInterval',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
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
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_ARCHIVE_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutoArchiveInterval(
    @Param('sandboxId') sandboxId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.sandboxService.setAutoArchiveInterval(sandboxId, interval)
  }

  @Post(':sandboxId/autodelete/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-delete interval',
    operationId: 'setAutoDeleteInterval',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiParam({
    name: 'interval',
    description:
      'Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)',
    type: 'number',
  })
  @ApiResponse({
    status: 200,
    description: 'Auto-delete interval has been set',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_DELETE_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutoDeleteInterval(
    @Param('sandboxId') sandboxId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.sandboxService.setAutoDeleteInterval(sandboxId, interval)
  }

  @Post(':sandboxId/archive')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Archive sandbox',
    operationId: 'archiveSandbox',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been archived',
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.ARCHIVE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async archiveSandbox(@Param('sandboxId') sandboxId: string): Promise<void> {
    return this.sandboxService.archive(sandboxId)
  }

  @Get(':sandboxId/ports/:port/preview-url')
  @ApiOperation({
    summary: 'Get preview URL for a sandbox port',
    operationId: 'getPortPreviewUrl',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
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
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.GET_PORT_PREVIEW_URL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        port: req.params.port,
      }),
    },
  })
  async getPortPreviewUrl(
    @Param('sandboxId') sandboxId: string,
    @Param('port') port: number,
  ): Promise<PortPreviewUrlDto> {
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    const proxyDomain = this.configService.get('proxy.domain')
    const proxyProtocol = this.configService.get('proxy.protocol')
    if (proxyDomain && proxyProtocol) {
      const sandbox = await this.sandboxService.findOne(sandboxId)
      if (!sandbox) {
        throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
      }

      // Get runner info
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (!runner) {
        throw new NotFoundException(`Runner not found for sandbox ${sandboxId}`)
      }

      // Return new preview url only for updated sandboxes
      if (sandbox.daemonVersion) {
        return {
          url: `${proxyProtocol}://${port}-${sandbox.id}.${proxyDomain}`,
          legacyProxyUrl: `https://${port}-${sandbox.id}.${runner.domain}`,
          token: sandbox.authToken,
        }
      }
    }

    return this.sandboxService.getPortPreviewUrl(sandboxId, port)
  }

  @Get(':sandboxId/build-logs')
  @ApiOperation({
    summary: 'Get build logs',
    operationId: 'getBuildLogs',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
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
  @UseGuards(SandboxAccessGuard)
  async getBuildLogs(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
    @Param('sandboxId') sandboxId: string,
    @Query('follow', new ParseBoolPipe({ optional: true })) follow?: boolean,
  ): Promise<void> {
    const sandbox = await this.sandboxService.findOne(sandboxId)
    if (!sandbox || !sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found or has no runner assigned`)
    }

    if (!sandbox.buildInfo) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} has no build info`)
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner for sandbox ${sandboxId} not found`)
    }

    const logProxy = new LogProxy(
      runner.apiUrl,
      sandbox.buildInfo.snapshotRef.split(':')[0],
      runner.apiKey,
      follow === true,
      req,
      res,
      next,
    )
    return logProxy.create()
  }

  private async waitForSandboxStarted(sandboxId: string, timeoutSeconds: number): Promise<SandboxDto> {
    const waitForStarted = new Promise<SandboxDto>((resolve, reject) => {
      // eslint-disable-next-line
      let timeout: NodeJS.Timeout
      const handleStateUpdated = (event: SandboxStateUpdatedEvent) => {
        if (event.sandbox.id !== sandboxId) {
          return
        }
        if (event.sandbox.state === SandboxState.STARTED) {
          this.eventEmitter.off(SandboxEvents.STATE_UPDATED, handleStateUpdated)
          clearTimeout(timeout)
          resolve(SandboxDto.fromSandbox(event.sandbox, ''))
        }
        if (event.sandbox.state === SandboxState.ERROR || event.sandbox.state === SandboxState.BUILD_FAILED) {
          this.eventEmitter.off(SandboxEvents.STATE_UPDATED, handleStateUpdated)
          clearTimeout(timeout)
          reject(new BadRequestError(`Sandbox failed to start: ${event.sandbox.errorReason}`))
        }
      }

      this.eventEmitter.on(SandboxEvents.STATE_UPDATED, handleStateUpdated)
      timeout = setTimeout(() => {
        this.eventEmitter.off(SandboxEvents.STATE_UPDATED, handleStateUpdated)
        reject(new BadRequestError(`Sandbox failed to start: Timeout after ${timeoutSeconds} seconds`))
      }, timeoutSeconds * 1000)
    })

    return waitForStarted
  }
}
