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
  Request,
  Logger,
  UseGuards,
  HttpCode,
  UseInterceptors,
  RawBodyRequest,
  Res,
  Next,
} from '@nestjs/common'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import {
  ApiOAuth2,
  ApiResponse,
  ApiQuery,
  ApiOperation,
  ApiConsumes,
  ApiBody,
  ApiTags,
  ApiParam,
  ApiHeader,
  ApiBearerAuth,
} from '@nestjs/swagger'
import {
  FileInfoDto,
  MatchDto,
  SearchFilesResponseDto,
  ReplaceRequestDto,
  ReplaceResultDto,
  GitAddRequestDto,
  GitBranchRequestDto,
  GitDeleteBranchRequestDto,
  GitCloneRequestDto,
  GitCommitRequestDto,
  GitCommitResponseDto,
  GitRepoRequestDto,
  GitStatusDto,
  ListBranchResponseDto,
  GitCommitInfoDto,
  GitCheckoutRequestDto,
  ExecuteRequestDto,
  ExecuteResponseDto,
  ProjectDirResponseDto,
  CreateSessionRequestDto,
  SessionExecuteRequestDto,
  SessionExecuteResponseDto,
  SessionDto,
  CommandDto,
  MousePositionDto,
  MouseMoveRequestDto,
  MouseMoveResponseDto,
  MouseClickRequestDto,
  MouseClickResponseDto,
  MouseDragRequestDto,
  MouseDragResponseDto,
  MouseScrollRequestDto,
  MouseScrollResponseDto,
  KeyboardTypeRequestDto,
  KeyboardPressRequestDto,
  KeyboardHotkeyRequestDto,
  ScreenshotResponseDto,
  RegionScreenshotResponseDto,
  CompressedScreenshotResponseDto,
  DisplayInfoResponseDto,
  WindowsResponseDto,
  ComputerUseStartResponseDto,
  ComputerUseStopResponseDto,
  ComputerUseStatusResponseDto,
  ProcessStatusResponseDto,
  ProcessRestartResponseDto,
  ProcessLogsResponseDto,
  ProcessErrorsResponseDto,
} from '../dto/toolbox.dto'
import { ToolboxService } from '../services/toolbox.service'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import {
  CompletionListDto,
  LspCompletionParamsDto,
  LspDocumentRequestDto,
  LspSymbolDto,
  LspServerRequestDto,
} from '../dto/lsp.dto'
import { createProxyMiddleware, RequestHandler, fixRequestBody, Options } from 'http-proxy-middleware'
import { IncomingMessage } from 'http'
import { NextFunction } from 'express'
import { ServerResponse } from 'http'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import followRedirects from 'follow-redirects'
import { UploadFileDto } from '../dto/upload-file.dto'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

followRedirects.maxRedirects = 10
followRedirects.maxBodyLength = 50 * 1024 * 1024

@ApiTags('toolbox')
@Controller('toolbox')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard, SandboxAccessGuard)
@RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class ToolboxController {
  private readonly logger = new Logger(ToolboxController.name)
  private readonly toolboxProxy: RequestHandler<
    RawBodyRequest<IncomingMessage>,
    ServerResponse<IncomingMessage>,
    NextFunction
  >
  private readonly toolboxStreamProxy: RequestHandler<
    RawBodyRequest<IncomingMessage>,
    ServerResponse<IncomingMessage>,
    NextFunction
  >

  constructor(private readonly toolboxService: ToolboxService) {
    const commonProxyOptions: Options = {
      router: async (req: RawBodyRequest<IncomingMessage>) => {
        // eslint-disable-next-line no-useless-escape
        const sandboxId = req.url.match(/^\/api\/toolbox\/([^\/]+)\/toolbox/)?.[1]
        try {
          const runner = await this.toolboxService.getRunner(sandboxId)
          // @ts-expect-error - used later to set request headers
          req._runnerApiKey = runner.apiKey

          return runner.proxyUrl
        } catch (err) {
          // @ts-expect-error - used later to throw error
          req._err = err
        }

        // Must return a valid url
        return 'http://target-error'
      },
      pathRewrite: (path) => {
        // eslint-disable-next-line no-useless-escape
        const sandboxId = path.match(/^\/api\/toolbox\/([^\/]+)\/toolbox/)?.[1]
        const routePath = path.split(`/api/toolbox/${sandboxId}/toolbox`)[1]
        const newPath = `/sandboxes/${sandboxId}/toolbox${routePath}`

        // Handle files path which is served on /files/ in the daemon
        // TODO: Circle back to this after daemon versioning
        // We can then switch /files/ to /files and only perform this for older daemon versions
        const url = new URL(`http://runner${newPath}`)
        if (url.pathname.endsWith('/files')) {
          url.pathname = url.pathname + '/'
          return url.toString().replace('http://runner', '')
        }

        return newPath
      },
      changeOrigin: true,
      autoRewrite: true,
      proxyTimeout: 5 * 60 * 1000,
      on: {
        proxyReq: (proxyReq, req, res) => {
          // @ts-expect-error - set when routing
          if (req._err) {
            res.writeHead(400, { 'Content-Type': 'application/json' })
            // @ts-expect-error - set when routing
            res.end(JSON.stringify(req._err))
            return
          }

          // @ts-expect-error - set when routing
          const runnerApiKey = req._runnerApiKey

          try {
            proxyReq.setHeader('Authorization', `Bearer ${runnerApiKey}`)
          } catch {
            // Ignore error - headers are already set
            return
          }
          fixRequestBody(proxyReq, req)
        },
        proxyRes: (proxyRes, req, res) => {
          // console.log('proxyRes', proxyRes)
        },
      },
    }

    this.toolboxProxy = createProxyMiddleware({
      ...commonProxyOptions,
      followRedirects: true,
    })

    this.toolboxStreamProxy = createProxyMiddleware({
      ...commonProxyOptions,
      followRedirects: false,
    })
  }

  @Get(':sandboxId/toolbox/project-dir')
  @ApiOperation({
    summary: 'Get sandbox project dir',
    operationId: 'getProjectDir',
  })
  @ApiResponse({
    status: 200,
    description: 'Project directory retrieved successfully',
    type: ProjectDirResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getProjectDir(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/files')
  @ApiOperation({
    summary: 'List files',
    operationId: 'listFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Files listed successfully',
    type: [FileInfoDto],
  })
  @ApiQuery({ name: 'path', type: String, required: false })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async listFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Delete(':sandboxId/toolbox/files')
  @ApiOperation({
    summary: 'Delete file',
    description: 'Delete file inside sandbox',
    operationId: 'deleteFile',
  })
  @ApiResponse({
    status: 200,
    description: 'File deleted successfully',
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_DELETE_FILE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        path: req.query.path,
      }),
    },
  })
  async deleteFile(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/files/download')
  @ApiOperation({
    summary: 'Download file',
    description: 'Download file from sandbox',
    operationId: 'downloadFile',
  })
  @ApiResponse({
    status: 200,
    description: 'File downloaded successfully',
    schema: {
      type: 'string',
      format: 'binary',
    },
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_DOWNLOAD_FILE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        path: req.query.path,
      }),
    },
  })
  async downloadFile(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/files/find')
  @ApiOperation({
    summary: 'Search for text/pattern in files',
    description: 'Search for text/pattern inside sandbox files',
    operationId: 'findInFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Search completed successfully',
    type: [MatchDto],
  })
  @ApiQuery({ name: 'pattern', type: String, required: true })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async findInFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/files/folder')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create folder',
    description: 'Create folder inside sandbox',
    operationId: 'createFolder',
  })
  @ApiResponse({
    status: 200,
    description: 'Folder created successfully',
  })
  @ApiQuery({ name: 'mode', type: String, required: true })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_CREATE_FOLDER,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        path: req.query.path,
        mode: req.query.mode,
      }),
    },
  })
  async createFolder(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/files/info')
  @ApiOperation({
    summary: 'Get file info',
    description: 'Get file info inside sandbox',
    operationId: 'getFileInfo',
  })
  @ApiResponse({
    status: 200,
    description: 'File info retrieved successfully',
    type: FileInfoDto,
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getFileInfo(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/files/move')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Move file',
    description: 'Move file inside sandbox',
    operationId: 'moveFile',
  })
  @ApiResponse({
    status: 200,
    description: 'File moved successfully',
  })
  @ApiQuery({ name: 'destination', type: String, required: true })
  @ApiQuery({ name: 'source', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_MOVE_FILE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        destination: req.query.destination,
        source: req.query.source,
      }),
    },
  })
  async moveFile(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/files/permissions')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Set file permissions',
    description: 'Set file owner/group/permissions inside sandbox',
    operationId: 'setFilePermissions',
  })
  @ApiResponse({
    status: 200,
    description: 'File permissions updated successfully',
  })
  @ApiQuery({ name: 'mode', type: String, required: false })
  @ApiQuery({ name: 'group', type: String, required: false })
  @ApiQuery({ name: 'owner', type: String, required: false })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_SET_FILE_PERMISSIONS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        mode: req.query.mode,
        group: req.query.group,
        owner: req.query.owner,
        path: req.query.path,
      }),
    },
  })
  async setFilePermissions(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/files/replace')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Replace in files',
    description: 'Replace text/pattern in multiple files inside sandbox',
    operationId: 'replaceInFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Text replaced successfully',
    type: [ReplaceResultDto],
  })
  @ApiBody({
    type: ReplaceRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_REPLACE_IN_FILES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<ReplaceRequestDto>) => ({
        files: req.body?.files,
        pattern: req.body?.pattern,
        newValue: req.body?.newValue,
      }),
    },
  })
  async replaceInFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/files/search')
  @ApiOperation({
    summary: 'Search files',
    description: 'Search for files inside sandbox',
    operationId: 'searchFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Search completed successfully',
    type: SearchFilesResponseDto,
  })
  @ApiQuery({ name: 'pattern', type: String, required: true })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async searchFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @HttpCode(200)
  @Post(':sandboxId/toolbox/files/upload')
  @ApiOperation({
    summary: 'Upload file',
    description: 'Upload file inside sandbox',
    operationId: 'uploadFile',
    deprecated: true,
  })
  @ApiResponse({
    status: 200,
    description: 'File uploaded successfully',
  })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    schema: {
      type: 'object',
      properties: {
        file: {
          type: 'string',
          format: 'binary',
        },
      },
    },
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_UPLOAD_FILE,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      query: (req) => ({
        path: req.query.path,
      }),
    },
  })
  async uploadFile(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return this.toolboxProxy(req, res, next)
  }

  @HttpCode(200)
  @Post(':sandboxId/toolbox/files/bulk-upload')
  @ApiOperation({
    summary: 'Upload multiple files',
    description: 'Upload multiple files inside sandbox',
    operationId: 'uploadFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Files uploaded successfully',
  })
  @ApiConsumes('multipart/form-data')
  @ApiBody({ type: [UploadFileDto] })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_BULK_UPLOAD_FILES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async uploadFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return this.toolboxStreamProxy(req, res, next)
  }

  // Git operations
  @Post(':sandboxId/toolbox/git/add')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Add files',
    description: 'Add files to git commit',
    operationId: 'gitAddFiles',
  })
  @ApiResponse({
    status: 200,
    description: 'Files added to git successfully',
  })
  @ApiBody({
    type: GitAddRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_ADD_FILES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitAddRequestDto>) => ({
        path: req.body?.path,
        files: req.body?.files,
      }),
    },
  })
  async gitAddFiles(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/git/branches')
  @ApiOperation({
    summary: 'Get branch list',
    description: 'Get branch list from git repository',
    operationId: 'gitListBranches',
  })
  @ApiResponse({
    status: 200,
    description: 'Branch list retrieved successfully',
    type: ListBranchResponseDto,
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async gitBranchList(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/branches')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create branch',
    description: 'Create branch on git repository',
    operationId: 'gitCreateBranch',
  })
  @ApiResponse({
    status: 200,
    description: 'Branch created successfully',
  })
  @ApiBody({
    type: GitBranchRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_CREATE_BRANCH,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitBranchRequestDto>) => ({
        path: req.body?.path,
        name: req.body?.name,
      }),
    },
  })
  async gitCreateBranch(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Delete(':sandboxId/toolbox/git/branches')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Delete branch',
    description: 'Delete branch on git repository',
    operationId: 'gitDeleteBranch',
  })
  @ApiResponse({
    status: 200,
    description: 'Branch deleted successfully',
  })
  @ApiBody({
    type: GitDeleteBranchRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_DELETE_BRANCH,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitDeleteBranchRequestDto>) => ({
        path: req.body?.path,
        name: req.body?.name,
      }),
    },
  })
  async gitDeleteBranch(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/clone')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Clone repository',
    description: 'Clone git repository',
    operationId: 'gitCloneRepository',
  })
  @ApiResponse({
    status: 200,
    description: 'Repository cloned successfully',
  })
  @ApiBody({
    type: GitCloneRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_CLONE_REPOSITORY,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitCloneRequestDto>) => ({
        url: req.body?.url,
        path: req.body?.path,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
        branch: req.body?.branch,
        commit_id: req.body?.commit_id,
      }),
    },
  })
  async gitCloneRepository(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/commit')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Commit changes',
    description: 'Commit changes to git repository',
    operationId: 'gitCommitChanges',
  })
  @ApiResponse({
    status: 200,
    description: 'Changes committed successfully',
    type: GitCommitResponseDto,
  })
  @ApiBody({
    type: GitCommitRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_COMMIT_CHANGES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitCommitRequestDto>) => ({
        path: req.body?.path,
        message: req.body?.message,
        author: req.body?.author,
        email: req.body?.email,
      }),
    },
  })
  async gitCommitChanges(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/git/history')
  @ApiOperation({
    summary: 'Get commit history',
    description: 'Get commit history from git repository',
    operationId: 'gitGetHistory',
  })
  @ApiResponse({
    status: 200,
    description: 'Commit history retrieved successfully',
    type: [GitCommitInfoDto],
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async gitCommitHistory(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/pull')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Pull changes',
    description: 'Pull changes from remote',
    operationId: 'gitPullChanges',
  })
  @ApiResponse({
    status: 200,
    description: 'Changes pulled successfully',
  })
  @ApiBody({
    type: GitRepoRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_PULL_CHANGES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitRepoRequestDto>) => ({
        path: req.body?.path,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
      }),
    },
  })
  async gitPullChanges(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/push')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Push changes',
    description: 'Push changes to remote',
    operationId: 'gitPushChanges',
  })
  @ApiResponse({
    status: 200,
    description: 'Changes pushed successfully',
  })
  @ApiBody({
    type: GitRepoRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_PUSH_CHANGES,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitRepoRequestDto>) => ({
        path: req.body?.path,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
      }),
    },
  })
  async gitPushChanges(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/git/checkout')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Checkout branch',
    description: 'Checkout branch or commit in git repository',
    operationId: 'gitCheckoutBranch',
  })
  @ApiResponse({
    status: 200,
    description: 'Branch checked out successfully',
  })
  @ApiBody({
    type: GitCheckoutRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_GIT_CHECKOUT_BRANCH,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<GitCheckoutRequestDto>) => ({
        path: req.body?.path,
        branch: req.body?.branch,
      }),
    },
  })
  async gitCheckoutBranch(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/git/status')
  @ApiOperation({
    summary: 'Get git status',
    description: 'Get status from git repository',
    operationId: 'gitGetStatus',
  })
  @ApiResponse({
    status: 200,
    description: 'Git status retrieved successfully',
    type: GitStatusDto,
  })
  @ApiQuery({ name: 'path', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async gitStatus(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/process/execute')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Execute command',
    description: 'Execute command synchronously inside sandbox',
    operationId: 'executeCommand',
  })
  @ApiResponse({
    status: 200,
    description: 'Command executed successfully',
    type: ExecuteResponseDto,
  })
  @Audit({
    action: AuditAction.TOOLBOX_EXECUTE_COMMAND,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<ExecuteRequestDto>) => ({
        command: req.body?.command,
        cwd: req.body?.cwd,
        timeout: req.body?.timeout,
      }),
    },
  })
  async executeCommand(
    @Param('sandboxId') sandboxId: string,
    @Body() executeRequest: ExecuteRequestDto,
  ): Promise<ExecuteResponseDto> {
    const response = await this.toolboxService.forwardRequestToRunner(
      sandboxId,
      'POST',
      '/toolbox/process/execute',
      executeRequest,
    )

    // TODO: use new proxy - can't use it now because of this
    return {
      exitCode: response.code,
      result: response.result,
    }
  }

  // Session management endpoints
  @Get(':sandboxId/toolbox/process/session')
  @ApiOperation({
    summary: 'List sessions',
    description: 'List all active sessions in the sandbox',
    operationId: 'listSessions',
  })
  @ApiResponse({
    status: 200,
    description: 'Sessions retrieved successfully',
    type: [SessionDto],
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async listSessions(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/process/session/:sessionId')
  @ApiOperation({
    summary: 'Get session',
    description: 'Get session by ID',
    operationId: 'getSession',
  })
  @ApiResponse({
    status: 200,
    description: 'Session retrieved successfully',
    type: SessionDto,
  })
  @ApiParam({ name: 'sessionId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getSession(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/process/session')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create session',
    description: 'Create a new session in the sandbox',
    operationId: 'createSession',
  })
  @ApiResponse({
    status: 200,
  })
  @ApiBody({
    type: CreateSessionRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_CREATE_SESSION,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      body: (req: TypedRequest<CreateSessionRequestDto>) => ({
        sessionId: req.body?.sessionId,
      }),
    },
  })
  async createSession(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/process/session/:sessionId/exec')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Execute command in session',
    description: 'Execute a command in a specific session',
    operationId: 'executeSessionCommand',
  })
  @ApiResponse({
    status: 200,
    description: 'Command executed successfully',
    type: SessionExecuteResponseDto,
  })
  @ApiResponse({
    status: 202,
    description: 'Command accepted and is being processed',
    type: SessionExecuteResponseDto,
  })
  @ApiBody({
    type: SessionExecuteRequestDto,
  })
  @ApiParam({ name: 'sessionId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_SESSION_EXECUTE_COMMAND,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        sessionId: req.params.sessionId,
      }),
      body: (req: TypedRequest<SessionExecuteRequestDto>) => ({
        command: req.body?.command,
        runAsync: req.body?.runAsync,
        async: req.body?.async,
      }),
    },
  })
  async executeSessionCommand(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Delete(':sandboxId/toolbox/process/session/:sessionId')
  @ApiOperation({
    summary: 'Delete session',
    description: 'Delete a specific session',
    operationId: 'deleteSession',
  })
  @ApiResponse({
    status: 200,
    description: 'Session deleted successfully',
  })
  @ApiParam({ name: 'sessionId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_DELETE_SESSION,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        sessionId: req.params.sessionId,
      }),
    },
  })
  async deleteSession(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/process/session/:sessionId/command/:commandId')
  @ApiOperation({
    summary: 'Get session command',
    description: 'Get session command by ID',
    operationId: 'getSessionCommand',
  })
  @ApiResponse({
    status: 200,
    description: 'Session command retrieved successfully',
    type: CommandDto,
  })
  @ApiParam({ name: 'commandId', type: String, required: true })
  @ApiParam({ name: 'sessionId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getSessionCommand(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/process/session/:sessionId/command/:commandId/logs')
  @ApiOperation({
    summary: 'Get command logs',
    description: 'Get logs for a specific command in a session',
    operationId: 'getSessionCommandLogs',
  })
  // When follow is true, the response is an octet stream
  @ApiResponse({
    status: 200,
    description: 'Command log stream',
    content: {
      'text/plain': {
        schema: {
          type: 'string',
        },
      },
    },
  })
  @ApiQuery({ name: 'follow', type: Boolean, required: false })
  @ApiParam({ name: 'commandId', type: String, required: true })
  @ApiParam({ name: 'sessionId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getSessionCommandLogs(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/lsp/completions')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Get Lsp Completions',
    description:
      'The Completion request is sent from the client to the server to compute completion items at a given cursor position.',
    operationId: 'LspCompletions',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
    type: CompletionListDto,
  })
  @ApiBody({
    type: LspCompletionParamsDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getLspCompletions(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/lsp/did-close')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Call Lsp DidClose',
    description:
      'The document close notification is sent from the client to the server when the document got closed in the client.',
    operationId: 'LspDidClose',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
  })
  @ApiBody({
    type: LspDocumentRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async lspDidClose(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/lsp/did-open')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Call Lsp DidOpen',
    description:
      'The document open notification is sent from the client to the server to signal newly opened text documents.',
    operationId: 'LspDidOpen',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
  })
  @ApiBody({
    type: LspDocumentRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async lspDidOpen(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/lsp/document-symbols')
  @ApiOperation({
    summary: 'Call Lsp DocumentSymbols',
    description: 'The document symbol request is sent from the client to the server.',
    operationId: 'LspDocumentSymbols',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
    type: [LspSymbolDto],
  })
  @ApiQuery({ name: 'uri', type: String, required: true })
  @ApiQuery({ name: 'pathToProject', type: String, required: true })
  @ApiQuery({ name: 'languageId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getLspDocumentSymbols(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/lsp/start')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Start Lsp server',
    description: 'Start Lsp server process inside sandbox project',
    operationId: 'LspStart',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
  })
  @ApiBody({
    type: LspServerRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async startLspServer(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/lsp/stop')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Stop Lsp server',
    description: 'Stop Lsp server process inside sandbox project',
    operationId: 'LspStop',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
  })
  @ApiBody({
    type: LspServerRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async stopLspServer(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/lsp/workspace-symbols')
  @ApiOperation({
    summary: 'Call Lsp WorkspaceSymbols',
    description:
      'The workspace symbol request is sent from the client to the server to list project-wide symbols matching the query string.',
    operationId: 'LspWorkspaceSymbols',
  })
  @ApiResponse({
    status: 200,
    description: 'OK',
    type: [LspSymbolDto],
  })
  @ApiQuery({ name: 'query', type: String, required: true })
  @ApiQuery({ name: 'pathToProject', type: String, required: true })
  @ApiQuery({ name: 'languageId', type: String, required: true })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getLspWorkspaceSymbols(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  // Computer Use endpoints

  // Computer use management endpoints
  @Post(':sandboxId/toolbox/computeruse/start')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Start computer use processes',
    description: 'Start all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)',
    operationId: 'startComputerUse',
  })
  @ApiResponse({
    status: 200,
    description: 'Computer use processes started successfully',
    type: ComputerUseStartResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_COMPUTER_USE_START,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async startComputerUse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/stop')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Stop computer use processes',
    description: 'Stop all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)',
    operationId: 'stopComputerUse',
  })
  @ApiResponse({
    status: 200,
    description: 'Computer use processes stopped successfully',
    type: ComputerUseStopResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_COMPUTER_USE_STOP,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
  })
  async stopComputerUse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/status')
  @ApiOperation({
    summary: 'Get computer use status',
    description: 'Get status of all VNC desktop processes',
    operationId: 'getComputerUseStatus',
  })
  @ApiResponse({
    status: 200,
    description: 'Computer use status retrieved successfully',
    type: ComputerUseStatusResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getComputerUseStatus(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/process/:processName/status')
  @ApiOperation({
    summary: 'Get process status',
    description: 'Get status of a specific VNC process',
    operationId: 'getProcessStatus',
  })
  @ApiResponse({
    status: 200,
    description: 'Process status retrieved successfully',
    type: ProcessStatusResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @ApiParam({ name: 'processName', type: String, required: true })
  async getProcessStatus(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/process/:processName/restart')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Restart process',
    description: 'Restart a specific VNC process',
    operationId: 'restartProcess',
  })
  @ApiResponse({
    status: 200,
    description: 'Process restarted successfully',
    type: ProcessRestartResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @ApiParam({ name: 'processName', type: String, required: true })
  @Audit({
    action: AuditAction.TOOLBOX_COMPUTER_USE_RESTART_PROCESS,
    targetType: AuditTarget.SANDBOX,
    targetIdFromRequest: (req) => req.params.sandboxId,
    requestMetadata: {
      params: (req) => ({
        processName: req.params.processName,
      }),
    },
  })
  async restartProcess(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/process/:processName/logs')
  @ApiOperation({
    summary: 'Get process logs',
    description: 'Get logs for a specific VNC process',
    operationId: 'getProcessLogs',
  })
  @ApiResponse({
    status: 200,
    description: 'Process logs retrieved successfully',
    type: ProcessLogsResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @ApiParam({ name: 'processName', type: String, required: true })
  async getProcessLogs(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/process/:processName/errors')
  @ApiOperation({
    summary: 'Get process errors',
    description: 'Get error logs for a specific VNC process',
    operationId: 'getProcessErrors',
  })
  @ApiResponse({
    status: 200,
    description: 'Process errors retrieved successfully',
    type: ProcessErrorsResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  @ApiParam({ name: 'processName', type: String, required: true })
  async getProcessErrors(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  // Mouse endpoints
  @Get(':sandboxId/toolbox/computeruse/mouse/position')
  @ApiOperation({
    summary: 'Get mouse position',
    description: 'Get current mouse cursor position',
    operationId: 'getMousePosition',
  })
  @ApiResponse({
    status: 200,
    description: 'Mouse position retrieved successfully',
    type: MousePositionDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getMousePosition(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/mouse/move')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Move mouse',
    description: 'Move mouse cursor to specified coordinates',
    operationId: 'moveMouse',
  })
  @ApiResponse({
    status: 200,
    description: 'Mouse moved successfully',
    type: MouseMoveResponseDto,
  })
  @ApiBody({
    type: MouseMoveRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async moveMouse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/mouse/click')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Click mouse',
    description: 'Click mouse at specified coordinates',
    operationId: 'clickMouse',
  })
  @ApiResponse({
    status: 200,
    description: 'Mouse clicked successfully',
    type: MouseClickResponseDto,
  })
  @ApiBody({
    type: MouseClickRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async clickMouse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/mouse/drag')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Drag mouse',
    description: 'Drag mouse from start to end coordinates',
    operationId: 'dragMouse',
  })
  @ApiResponse({
    status: 200,
    description: 'Mouse dragged successfully',
    type: MouseDragResponseDto,
  })
  @ApiBody({
    type: MouseDragRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async dragMouse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/mouse/scroll')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Scroll mouse',
    description: 'Scroll mouse at specified coordinates',
    operationId: 'scrollMouse',
  })
  @ApiResponse({
    status: 200,
    description: 'Mouse scrolled successfully',
    type: MouseScrollResponseDto,
  })
  @ApiBody({
    type: MouseScrollRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async scrollMouse(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  // Keyboard endpoints
  @Post(':sandboxId/toolbox/computeruse/keyboard/type')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Type text',
    description: 'Type text using keyboard',
    operationId: 'typeText',
  })
  @ApiResponse({
    status: 200,
    description: 'Text typed successfully',
  })
  @ApiBody({
    type: KeyboardTypeRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async typeText(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/keyboard/key')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Press key',
    description: 'Press a key with optional modifiers',
    operationId: 'pressKey',
  })
  @ApiResponse({
    status: 200,
    description: 'Key pressed successfully',
  })
  @ApiBody({
    type: KeyboardPressRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async pressKey(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Post(':sandboxId/toolbox/computeruse/keyboard/hotkey')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Press hotkey',
    description: 'Press a hotkey combination',
    operationId: 'pressHotkey',
  })
  @ApiResponse({
    status: 200,
    description: 'Hotkey pressed successfully',
  })
  @ApiBody({
    type: KeyboardHotkeyRequestDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async pressHotkey(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  // Screenshot endpoints
  @Get(':sandboxId/toolbox/computeruse/screenshot')
  @ApiOperation({
    summary: 'Take screenshot',
    description: 'Take a screenshot of the entire screen',
    operationId: 'takeScreenshot',
  })
  @ApiResponse({
    status: 200,
    description: 'Screenshot taken successfully',
    type: ScreenshotResponseDto,
  })
  @ApiQuery({ name: 'show_cursor', type: Boolean, required: false })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async takeScreenshot(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/screenshot/region')
  @ApiOperation({
    summary: 'Take region screenshot',
    description: 'Take a screenshot of a specific region',
    operationId: 'takeRegionScreenshot',
  })
  @ApiResponse({
    status: 200,
    description: 'Region screenshot taken successfully',
    type: RegionScreenshotResponseDto,
  })
  @ApiQuery({ name: 'x', type: Number, required: true })
  @ApiQuery({ name: 'y', type: Number, required: true })
  @ApiQuery({ name: 'width', type: Number, required: true })
  @ApiQuery({ name: 'height', type: Number, required: true })
  @ApiQuery({ name: 'show_cursor', type: Boolean, required: false })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async takeRegionScreenshot(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/screenshot/compressed')
  @ApiOperation({
    summary: 'Take compressed screenshot',
    description: 'Take a compressed screenshot with format, quality, and scale options',
    operationId: 'takeCompressedScreenshot',
  })
  @ApiResponse({
    status: 200,
    description: 'Compressed screenshot taken successfully',
    type: CompressedScreenshotResponseDto,
  })
  @ApiQuery({ name: 'show_cursor', type: Boolean, required: false })
  @ApiQuery({ name: 'format', type: String, required: false })
  @ApiQuery({ name: 'quality', type: Number, required: false })
  @ApiQuery({ name: 'scale', type: Number, required: false })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async takeCompressedScreenshot(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/screenshot/region/compressed')
  @ApiOperation({
    summary: 'Take compressed region screenshot',
    description: 'Take a compressed screenshot of a specific region',
    operationId: 'takeCompressedRegionScreenshot',
  })
  @ApiResponse({
    status: 200,
    description: 'Compressed region screenshot taken successfully',
    type: CompressedScreenshotResponseDto,
  })
  @ApiQuery({ name: 'x', type: Number, required: true })
  @ApiQuery({ name: 'y', type: Number, required: true })
  @ApiQuery({ name: 'width', type: Number, required: true })
  @ApiQuery({ name: 'height', type: Number, required: true })
  @ApiQuery({ name: 'show_cursor', type: Boolean, required: false })
  @ApiQuery({ name: 'format', type: String, required: false })
  @ApiQuery({ name: 'quality', type: Number, required: false })
  @ApiQuery({ name: 'scale', type: Number, required: false })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async takeCompressedRegionScreenshot(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  // Display endpoints
  @Get(':sandboxId/toolbox/computeruse/display/info')
  @ApiOperation({
    summary: 'Get display info',
    description: 'Get information about displays',
    operationId: 'getDisplayInfo',
  })
  @ApiResponse({
    status: 200,
    description: 'Display info retrieved successfully',
    type: DisplayInfoResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getDisplayInfo(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }

  @Get(':sandboxId/toolbox/computeruse/display/windows')
  @ApiOperation({
    summary: 'Get windows',
    description: 'Get list of open windows',
    operationId: 'getWindows',
  })
  @ApiResponse({
    status: 200,
    description: 'Windows list retrieved successfully',
    type: WindowsResponseDto,
  })
  @ApiParam({ name: 'sandboxId', type: String, required: true })
  async getWindows(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
  ): Promise<void> {
    return await this.toolboxProxy(req, res, next)
  }
}
