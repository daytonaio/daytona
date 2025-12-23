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
import { UpdateSandboxStateDto } from '../dto/update-sandbox-state.dto'
import { PaginatedSandboxesDto } from '../dto/paginated-sandboxes.dto'
import { RunnerService } from '../services/runner.service'
import { RunnerAuthGuard } from '../../auth/runner-auth.guard'
import { RunnerContextDecorator } from '../../common/decorators/runner-context.decorator'
import { RunnerContext } from '../../common/interfaces/runner-context.interface'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
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
// import { UpdateSandboxNetworkSettingsDto } from '../dto/update-sandbox-network-settings.dto'
import { SshAccessDto, SshAccessValidationDto } from '../dto/ssh-access.dto'
import { ListSandboxesQueryDto } from '../dto/list-sandboxes-query.dto'
import { ProxyGuard } from '../../auth/proxy.guard'
import { OrGuard } from '../../auth/or.guard'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { SkipThrottle } from '@nestjs/throttler'
import { ThrottlerScope } from '../../common/decorators/throttler-scope.decorator'
import { SshGatewayGuard } from '../../auth/ssh-gateway.guard'
import { ToolboxProxyUrlDto } from '../dto/toolbox-proxy-url.dto'
import { UrlDto } from '../../common/dto/url.dto'

@ApiTags('sandbox')
@Controller('sandbox')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard, AuthenticatedRateLimitGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class SandboxController {
  private readonly logger = new Logger(SandboxController.name)

  constructor(
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
    const labels = labelsQuery ? JSON.parse(labelsQuery) : undefined
    const sandboxes = await this.sandboxService.findAllDeprecated(
      authContext.organizationId,
      labels,
      includeErroredDeleted,
    )

    return sandboxes.map((sandbox) => {
      return SandboxDto.fromSandbox(sandbox)
    })
  }

  @Get('paginated')
  @ApiOperation({
    summary: 'List all sandboxes paginated',
    operationId: 'listSandboxesPaginated',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of all sandboxes',
    type: PaginatedSandboxesDto,
  })
  async listSandboxesPaginated(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query() queryParams: ListSandboxesQueryDto,
  ): Promise<PaginatedSandboxesDto> {
    const {
      page,
      limit,
      id,
      name,
      labels,
      includeErroredDeleted: includeErroredDestroyed,
      states,
      snapshots,
      regions,
      minCpu,
      maxCpu,
      minMemoryGiB,
      maxMemoryGiB,
      minDiskGiB,
      maxDiskGiB,
      lastEventAfter,
      lastEventBefore,
      sort: sortField,
      order: sortDirection,
    } = queryParams

    const result = await this.sandboxService.findAll(
      authContext.organizationId,
      page,
      limit,
      {
        id,
        name,
        labels: labels ? JSON.parse(labels) : undefined,
        includeErroredDestroyed,
        states,
        snapshots,
        regionIds: regions,
        minCpu,
        maxCpu,
        minMemoryGiB,
        maxMemoryGiB,
        minDiskGiB,
        maxDiskGiB,
        lastEventAfter,
        lastEventBefore,
      },
      {
        field: sortField,
        direction: sortDirection,
      },
    )

    return {
      items: result.items.map((sandbox) => {
        return SandboxDto.fromSandbox(sandbox)
      }),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Post()
  @HttpCode(200) //  for Daytona Api compatibility
  @UseInterceptors(ContentTypeInterceptor)
  @SkipThrottle({ authenticated: true })
  @ThrottlerScope('sandbox-create')
  @ApiOperation({
    summary: 'Create a new sandbox',
    operationId: 'createSandbox',
  })
  @ApiResponse({
    status: 200,
    description: 'The sandbox has been successfully created.',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateSandboxDto>) => ({
        name: req.body?.name,
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
        networkBlockAll: req.body?.networkBlockAll,
        networkAllowList: req.body?.networkAllowList,
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

      await this.waitForSandboxStarted(sandbox, 30)
    }

    return sandbox
  }

  @Get('for-runner')
  @UseGuards(RunnerAuthGuard)
  @ApiOperation({
    summary: 'Get sandboxes for the authenticated runner',
    operationId: 'getSandboxesForRunner',
  })
  @ApiQuery({
    name: 'states',
    required: false,
    type: String,
    description: 'Comma-separated list of sandbox states to filter by',
  })
  @ApiQuery({
    name: 'skipReconcilingSandboxes',
    required: false,
    type: Boolean,
    description: 'Skip sandboxes where state differs from desired state',
  })
  @ApiResponse({
    status: 200,
    description: 'List of sandboxes for the authenticated runner',
    type: [SandboxDto],
  })
  async getSandboxesForRunner(
    @RunnerContextDecorator() runnerContext: RunnerContext,
    @Query('states') states?: string,
    @Query('skipReconcilingSandboxes') skipReconcilingSandboxes?: string,
  ): Promise<SandboxDto[]> {
    const stateArray = states
      ? states.split(',').map((s) => {
          if (!Object.values(SandboxState).includes(s as SandboxState)) {
            throw new BadRequestError(`Invalid sandbox state: ${s}`)
          }
          return s as SandboxState
        })
      : undefined

    const skip = skipReconcilingSandboxes === 'true'
    const sandboxes = await this.sandboxService.findByRunnerId(runnerContext.runnerId, stateArray, skip)

    return sandboxes.map((sandbox) => SandboxDto.fromSandbox(sandbox))
  }

  @Get(':sandboxIdOrName')
  @ApiOperation({
    summary: 'Get sandbox details',
    operationId: 'getSandbox',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    @Query('verbose') verbose?: boolean,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.findOneByIdOrName(sandboxIdOrName, authContext.organizationId)

    return SandboxDto.fromSandbox(sandbox)
  }

  @Delete(':sandboxIdOrName')
  @SkipThrottle({ authenticated: true })
  @ThrottlerScope('sandbox-lifecycle')
  @ApiOperation({
    summary: 'Delete sandbox',
    operationId: 'deleteSandbox',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been deleted',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async deleteSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.destroy(sandboxIdOrName, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Post(':sandboxIdOrName/start')
  @HttpCode(200)
  @SkipThrottle({ authenticated: true })
  @ThrottlerScope('sandbox-lifecycle')
  @ApiOperation({
    summary: 'Start sandbox',
    operationId: 'startSandbox',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been started or is being restored from archived state',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.START,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async startSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
  ): Promise<SandboxDto> {
    const sbx = await this.sandboxService.start(sandboxIdOrName, authContext.organization)
    let sandbox = SandboxDto.fromSandbox(sbx)

    if (![SandboxState.ARCHIVED, SandboxState.RESTORING, SandboxState.STARTED].includes(sandbox.state)) {
      sandbox = await this.waitForSandboxStarted(sandbox, 30)
    }

    return sandbox
  }

  @Post(':sandboxIdOrName/stop')
  @HttpCode(200) //  for Daytona Api compatibility
  @SkipThrottle({ authenticated: true })
  @ThrottlerScope('sandbox-lifecycle')
  @ApiOperation({
    summary: 'Stop sandbox',
    operationId: 'stopSandbox',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been stopped',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.STOP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async stopSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.stop(sandboxIdOrName, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Put(':sandboxIdOrName/labels')
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Replace sandbox labels',
    operationId: 'replaceLabels',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<SandboxLabelsDto>) => ({
        labels: req.body?.labels,
      }),
    },
  })
  async replaceLabels(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Body() labelsDto: SandboxLabelsDto,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.replaceLabels(
      sandboxIdOrName,
      labelsDto.labels,
      authContext.organizationId,
    )
    return SandboxDto.fromSandbox(sandbox)
  }

  @Put(':sandboxId/state')
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Update sandbox state',
    operationId: 'updateSandboxState',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox state has been successfully updated',
  })
  @UseGuards(RunnerAuthGuard)
  @UseGuards(SandboxAccessGuard)
  async updateSandboxState(
    @Param('sandboxId') sandboxId: string,
    @Body() updateStateDto: UpdateSandboxStateDto,
  ): Promise<void> {
    await this.sandboxService.updateState(sandboxId, updateStateDto.state, updateStateDto.errorReason)
  }

  @Post(':sandboxIdOrName/backup')
  @ApiOperation({
    summary: 'Create sandbox backup',
    operationId: 'createBackup',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async createBackup(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.createBackup(sandboxIdOrName, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Post(':sandboxIdOrName/public/:isPublic')
  @ApiOperation({
    summary: 'Update public status',
    operationId: 'updatePublicStatus',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiParam({
    name: 'isPublic',
    description: 'Public status to set',
    type: 'boolean',
  })
  @ApiResponse({
    status: 200,
    description: 'Public status has been successfully updated',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.UPDATE_PUBLIC_STATUS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      params: (req) => ({
        isPublic: req.params.isPublic,
      }),
    },
  })
  async updatePublicStatus(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Param('isPublic') isPublic: boolean,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.updatePublicStatus(sandboxIdOrName, isPublic, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Post(':sandboxId/last-activity')
  @ApiOperation({
    summary: 'Update sandbox last activity',
    operationId: 'updateLastActivity',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 201,
    description: 'Last activity has been updated',
  })
  @UseGuards(OrGuard([SandboxAccessGuard, ProxyGuard, SshGatewayGuard]))
  async updateLastActivity(@Param('sandboxId') sandboxId: string): Promise<void> {
    await this.sandboxService.updateLastActivityAt(sandboxId, new Date())
  }

  @Post(':sandboxIdOrName/autostop/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-stop interval',
    operationId: 'setAutostopInterval',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_STOP_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutostopInterval(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Param('interval') interval: number,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.setAutostopInterval(sandboxIdOrName, interval, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Post(':sandboxIdOrName/autoarchive/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-archive interval',
    operationId: 'setAutoArchiveInterval',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_ARCHIVE_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutoArchiveInterval(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Param('interval') interval: number,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.setAutoArchiveInterval(
      sandboxIdOrName,
      interval,
      authContext.organizationId,
    )
    return SandboxDto.fromSandbox(sandbox)
  }

  @Post(':sandboxIdOrName/autodelete/:interval')
  @ApiOperation({
    summary: 'Set sandbox auto-delete interval',
    operationId: 'setAutoDeleteInterval',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.SET_AUTO_DELETE_INTERVAL,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      params: (req) => ({
        interval: req.params.interval,
      }),
    },
  })
  async setAutoDeleteInterval(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Param('interval') interval: number,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.setAutoDeleteInterval(
      sandboxIdOrName,
      interval,
      authContext.organizationId,
    )
    return SandboxDto.fromSandbox(sandbox)
  }

  // TODO: Network settings endpoint will not be enabled for now
  // @Post(':sandboxIdOrName/network-settings')
  // @ApiOperation({
  //   summary: 'Update sandbox network settings',
  //   operationId: 'updateNetworkSettings',
  // })
  // @ApiParam({
  //   name: 'sandboxIdOrName',
  //   description: 'ID or name of the sandbox',
  //   type: 'string',
  // })
  // @ApiResponse({
  //   status: 200,
  //   description: 'Network settings have been updated',
  //   type: SandboxDto,
  // })
  // @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  // @UseGuards(SandboxAccessGuard)
  // @Audit({
  //   action: AuditAction.UPDATE_NETWORK_SETTINGS,
  //   targetType: AuditTarget.SANDBOX,
  //   targetIdFromRequest: (req) => req.params.sandboxIdOrName,
  //   targetIdFromResult: (result: SandboxDto) => result?.id,
  //   requestMetadata: {
  //     body: (req: TypedRequest<UpdateSandboxNetworkSettingsDto>) => ({
  //       networkBlockAll: req.body?.networkBlockAll,
  //       networkAllowList: req.body?.networkAllowList,
  //     }),
  //   },
  // })
  // async updateNetworkSettings(
  //   @AuthContext() authContext: OrganizationAuthContext,
  //   @Param('sandboxIdOrName') sandboxIdOrName: string,
  //   @Body() networkSettings: UpdateSandboxNetworkSettingsDto,
  // ): Promise<SandboxDto> {
  //   const sandbox = await this.sandboxService.updateNetworkSettings(
  //     sandboxIdOrName,
  //     networkSettings.networkBlockAll,
  //     networkSettings.networkAllowList,
  //     authContext.organizationId,
  //   )
  //   return SandboxDto.fromSandbox(sandbox, '')
  // }

  @Post(':sandboxIdOrName/archive')
  @HttpCode(200)
  @SkipThrottle({ authenticated: true })
  @ThrottlerScope('sandbox-lifecycle')
  @ApiOperation({
    summary: 'Archive sandbox',
    operationId: 'archiveSandbox',
  })
  @ApiResponse({
    status: 200,
    description: 'Sandbox has been archived',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.ARCHIVE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
  })
  async archiveSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.archive(sandboxIdOrName, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Get(':sandboxIdOrName/ports/:port/preview-url')
  @ApiOperation({
    summary: 'Get preview URL for a sandbox port',
    operationId: 'getPortPreviewUrl',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
  async getPortPreviewUrl(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Param('port') port: number,
  ): Promise<PortPreviewUrlDto> {
    return this.sandboxService.getPortPreviewUrl(sandboxIdOrName, authContext.organizationId, port)
  }

  @Get(':sandboxIdOrName/build-logs')
  @ApiOperation({
    summary: 'Get build logs',
    operationId: 'getBuildLogs',
    deprecated: true,
    description: 'This endpoint is deprecated. Use `getBuildLogsUrl` instead.',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
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
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Query('follow', new ParseBoolPipe({ optional: true })) follow?: boolean,
  ): Promise<void> {
    const sandbox = await this.sandboxService.findOneByIdOrName(sandboxIdOrName, authContext.organizationId)
    if (!sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} has no runner assigned`)
    }

    if (!sandbox.buildInfo) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} has no build info`)
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner for sandbox ${sandboxIdOrName} not found`)
    }

    if (!runner.apiUrl) {
      throw new NotFoundException(`Runner for sandbox ${sandboxIdOrName} has no API URL`)
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

  @Get(':sandboxIdOrName/build-logs-url')
  @ApiOperation({
    summary: 'Get build logs URL',
    operationId: 'getBuildLogsUrl',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Build logs URL',
    type: UrlDto,
  })
  @UseGuards(SandboxAccessGuard)
  async getBuildLogsUrl(@Param('sandboxIdOrName') sandboxIdOrName: string): Promise<UrlDto> {
    const buildLogsUrl = await this.sandboxService.getBuildLogsUrl(sandboxIdOrName)

    return new UrlDto(buildLogsUrl)
  }

  @Post(':sandboxIdOrName/ssh-access')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Create SSH access for sandbox',
    operationId: 'createSshAccess',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiQuery({
    name: 'expiresInMinutes',
    required: false,
    type: Number,
    description: 'Expiration time in minutes (default: 60)',
  })
  @ApiResponse({
    status: 200,
    description: 'SSH access has been created',
    type: SshAccessDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.CREATE_SSH_ACCESS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SshAccessDto) => result?.sandboxId,
    requestMetadata: {
      query: (req) => ({
        expiresInMinutes: req.query.expiresInMinutes,
      }),
    },
  })
  async createSshAccess(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Query('expiresInMinutes') expiresInMinutes?: number,
  ): Promise<SshAccessDto> {
    return await this.sandboxService.createSshAccess(sandboxIdOrName, expiresInMinutes, authContext.organizationId)
  }

  @Delete(':sandboxIdOrName/ssh-access')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Revoke SSH access for sandbox',
    operationId: 'revokeSshAccess',
  })
  @ApiParam({
    name: 'sandboxIdOrName',
    description: 'ID or name of the sandbox',
    type: 'string',
  })
  @ApiQuery({
    name: 'token',
    required: false,
    type: String,
    description: 'SSH access token to revoke. If not provided, all SSH access for the sandbox will be revoked.',
  })
  @ApiResponse({
    status: 200,
    description: 'SSH access has been revoked',
    type: SandboxDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  @Audit({
    action: AuditAction.REVOKE_SSH_ACCESS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxIdOrName,
    targetIdFromResult: (result: SandboxDto) => result?.id,
    requestMetadata: {
      query: (req) => ({
        token: req.query.token,
      }),
    },
  })
  async revokeSshAccess(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxIdOrName') sandboxIdOrName: string,
    @Query('token') token?: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.sandboxService.revokeSshAccess(sandboxIdOrName, token, authContext.organizationId)
    return SandboxDto.fromSandbox(sandbox)
  }

  @Get('ssh-access/validate')
  @ApiOperation({
    summary: 'Validate SSH access for sandbox',
    operationId: 'validateSshAccess',
  })
  @ApiQuery({
    name: 'token',
    required: true,
    type: String,
    description: 'SSH access token to validate',
  })
  @ApiResponse({
    status: 200,
    description: 'SSH access validation result',
    type: SshAccessValidationDto,
  })
  async validateSshAccess(@Query('token') token: string): Promise<SshAccessValidationDto> {
    const result = await this.sandboxService.validateSshAccess(token)
    return SshAccessValidationDto.fromValidationResult(result.valid, result.sandboxId)
  }

  @Get(':sandboxId/toolbox-proxy-url')
  @ApiOperation({
    summary: 'Get toolbox proxy URL for a sandbox',
    operationId: 'getToolboxProxyUrl',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Toolbox proxy URL for the specified sandbox',
    type: ToolboxProxyUrlDto,
  })
  @UseGuards(SandboxAccessGuard)
  async getToolboxProxyUrl(@Param('sandboxId') sandboxId: string): Promise<ToolboxProxyUrlDto> {
    const url = await this.sandboxService.getToolboxProxyUrl(sandboxId)
    return new ToolboxProxyUrlDto(url)
  }

  // wait up to `timeoutSeconds` for the sandbox to start; if it doesn’t, return current sandbox
  private async waitForSandboxStarted(sandbox: SandboxDto, timeoutSeconds: number): Promise<SandboxDto> {
    let latestSandbox: Sandbox
    const waitForStarted = new Promise<SandboxDto>((resolve, reject) => {
      // eslint-disable-next-line
      let timeout: NodeJS.Timeout
      const handleStateUpdated = (event: SandboxStateUpdatedEvent) => {
        if (event.sandbox.id !== sandbox.id) {
          return
        }
        latestSandbox = event.sandbox
        if (event.sandbox.state === SandboxState.STARTED) {
          this.eventEmitter.off(SandboxEvents.STATE_UPDATED, handleStateUpdated)
          clearTimeout(timeout)
          resolve(SandboxDto.fromSandbox(event.sandbox))
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
        if (latestSandbox) {
          resolve(SandboxDto.fromSandbox(latestSandbox))
        } else {
          resolve(sandbox)
        }
      }, timeoutSeconds * 1000)
    })

    return waitForStarted
  }
}
