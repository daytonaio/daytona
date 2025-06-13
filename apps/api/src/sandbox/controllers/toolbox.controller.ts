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

          return runner.apiUrl
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

          proxyReq.setHeader('Authorization', `Bearer ${runnerApiKey}`)
          fixRequestBody(proxyReq, req)
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
}
