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
  async listSandboxes(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('verbose') verbose?: boolean,
    @Query('labels') labelsQuery?: string,
  ): Promise<SandboxDto[]> {
    const labels = labelsQuery ? JSON.parse(labelsQuery) : {}
    const sandboxes = await this.sandboxService.findAll(authContext.organizationId, labels)
    const dtos = sandboxes.map(async (sandbox) => {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const dto = SandboxDto.fromSandbox(sandbox, runner.domain)
      return dto
    })
    return await Promise.all(dtos)
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

      // Wait for the sandbox to start
      const sandboxState = await this.waitForSandboxState(
        sandbox.id,
        SandboxState.STARTED,
        30000, // 30 seconds timeout
      )
      sandbox.state = sandboxState
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
  async removeSandbox(
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
  })
  @Throttle({ default: { limit: 100 } })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @UseGuards(SandboxAccessGuard)
  async startSandbox(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('sandboxId') sandboxId: string,
  ): Promise<void> {
    return this.sandboxService.start(sandboxId, authContext.organization)
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
  async setAutoArchiveInterval(
    @Param('sandboxId') sandboxId: string,
    @Param('interval') interval: number,
  ): Promise<void> {
    await this.sandboxService.setAutoArchiveInterval(sandboxId, interval)
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
  async getPortPreviewUrl(
    @Param('sandboxId') sandboxId: string,
    @Param('port') port: number,
  ): Promise<PortPreviewUrlDto> {
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

  private async waitForSandboxState(
    sandboxId: string,
    desiredState: SandboxState,
    timeout: number,
  ): Promise<SandboxState> {
    const startTime = Date.now()

    let sandboxState: SandboxState
    while (Date.now() - startTime < timeout) {
      const sandbox = await this.sandboxService.findOne(sandboxId)
      sandboxState = sandbox.state
      if (
        sandboxState === desiredState ||
        sandboxState === SandboxState.ERROR ||
        sandboxState === SandboxState.BUILD_FAILED
      ) {
        return sandboxState
      }
      await new Promise((resolve) => setTimeout(resolve, 100)) // Wait 100 ms before checking again
    }

    return sandboxState
  }
}
