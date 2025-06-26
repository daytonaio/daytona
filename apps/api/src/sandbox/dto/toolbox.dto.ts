/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsString, IsBoolean, IsOptional, IsArray } from 'class-validator'

@ApiSchema({ name: 'FileInfo' })
export class FileInfoDto {
  @ApiProperty()
  name: string

  @ApiProperty()
  isDir: boolean

  @ApiProperty()
  size: number

  @ApiProperty()
  modTime: string

  @ApiProperty()
  mode: string

  @ApiProperty()
  permissions: string

  @ApiProperty()
  owner: string

  @ApiProperty()
  group: string
}

@ApiSchema({ name: 'Match' })
export class MatchDto {
  @ApiProperty()
  file: string

  @ApiProperty()
  line: number

  @ApiProperty()
  content: string
}

@ApiSchema({ name: 'SearchRequest' })
export class SearchRequestDto {
  @ApiProperty({ description: 'Search query/pattern' })
  query: string

  @ApiPropertyOptional({ description: 'Path to search in (default: ".")' })
  path?: string

  @ApiPropertyOptional({ type: [String], description: 'File types to include (e.g., ["js", "ts"])' })
  file_types?: string[]

  @ApiPropertyOptional({ type: [String], description: 'Include glob patterns' })
  include_globs?: string[]

  @ApiPropertyOptional({ type: [String], description: 'Exclude glob patterns' })
  exclude_globs?: string[]

  @ApiPropertyOptional({ description: 'Case sensitive search (default: true)' })
  case_sensitive?: boolean

  @ApiPropertyOptional({ description: 'Enable multiline matching' })
  multiline?: boolean

  @ApiPropertyOptional({ description: 'Lines of context around matches' })
  context?: number

  @ApiPropertyOptional({ description: 'Return only match counts' })
  count_only?: boolean

  @ApiPropertyOptional({ description: 'Return only filenames with matches' })
  filenames_only?: boolean

  @ApiPropertyOptional({ description: 'Return structured JSON output' })
  json?: boolean

  @ApiPropertyOptional({ description: 'Maximum number of results' })
  max_results?: number

  @ApiPropertyOptional({ type: [String], description: 'Additional ripgrep arguments' })
  rg_args?: string[]
}

@ApiSchema({ name: 'SearchMatch' })
export class SearchMatchDto {
  @ApiProperty({ description: 'File path' })
  file: string

  @ApiProperty({ description: 'Line number' })
  line_number: number

  @ApiProperty({ description: 'Column number' })
  column: number

  @ApiProperty({ description: 'Line content' })
  line: string

  @ApiProperty({ description: 'Matched text' })
  match: string

  @ApiPropertyOptional({ type: [String], description: 'Context lines before match' })
  context_before?: string[]

  @ApiPropertyOptional({ type: [String], description: 'Context lines after match' })
  context_after?: string[]
}

@ApiSchema({ name: 'SearchResults' })
export class SearchResultsDto {
  @ApiProperty({ type: [SearchMatchDto], description: 'Array of matches' })
  matches: SearchMatchDto[]

  @ApiProperty({ description: 'Total number of matches' })
  total_matches: number

  @ApiProperty({ description: 'Total number of files with matches' })
  total_files: number

  @ApiPropertyOptional({ type: [String], description: 'List of files with matches' })
  files?: string[]

  @ApiPropertyOptional({ description: 'Raw output from search tool' })
  raw_output?: string
}

@ApiSchema({ name: 'ReplaceRequest' })
export class ReplaceRequestDto {
  @ApiProperty({ type: [String] })
  files: string[]

  @ApiProperty()
  pattern: string

  @ApiProperty()
  newValue: string
}

@ApiSchema({ name: 'ReplaceResult' })
export class ReplaceResultDto {
  @ApiPropertyOptional()
  file?: string

  @ApiPropertyOptional()
  success?: boolean

  @ApiPropertyOptional()
  error?: string
}

@ApiSchema({ name: 'GitAddRequest' })
export class GitAddRequestDto {
  @ApiProperty()
  path: string

  @ApiProperty({
    type: [String],
    description: 'files to add (use . for all files)',
  })
  files: string[]
}

@ApiSchema({ name: 'GitBranchRequest' })
export class GitBranchRequestDto {
  @ApiProperty()
  path: string

  @ApiProperty()
  name: string
}

@ApiSchema({ name: 'GitDeleteBranchRequest' })
export class GitDeleteBranchRequestDto {
  @ApiProperty()
  path: string

  @ApiProperty()
  name: string
}

@ApiSchema({ name: 'GitCloneRequest' })
export class GitCloneRequestDto {
  @ApiProperty()
  url: string

  @ApiProperty()
  path: string

  @ApiPropertyOptional()
  username?: string

  @ApiPropertyOptional()
  password?: string

  @ApiPropertyOptional()
  branch?: string

  @ApiPropertyOptional()
  commit_id?: string
}

@ApiSchema({ name: 'GitCommitRequest' })
export class GitCommitRequestDto {
  @ApiProperty()
  path: string

  @ApiProperty()
  message: string

  @ApiProperty()
  author: string

  @ApiProperty()
  email: string
}

@ApiSchema({ name: 'GitCommitResponse' })
export class GitCommitResponseDto {
  @ApiProperty()
  hash: string
}

@ApiSchema({ name: 'GitCheckoutRequest' })
export class GitCheckoutRequestDto {
  @ApiProperty()
  path: string

  @ApiProperty()
  branch: string
}

@ApiSchema({ name: 'GitRepoRequest' })
export class GitRepoRequestDto {
  @ApiProperty()
  path: string

  @ApiPropertyOptional()
  username?: string

  @ApiPropertyOptional()
  password?: string
}

@ApiSchema({ name: 'FileStatus' })
export class FileStatusDto {
  @ApiProperty()
  name: string

  @ApiProperty()
  staging: string

  @ApiProperty()
  worktree: string

  @ApiProperty()
  extra: string
}

@ApiSchema({ name: 'GitStatus' })
export class GitStatusDto {
  @ApiProperty()
  currentBranch: string

  @ApiProperty({
    type: [FileStatusDto],
  })
  fileStatus: FileStatusDto[]

  @ApiPropertyOptional()
  ahead?: number

  @ApiPropertyOptional()
  behind?: number

  @ApiPropertyOptional()
  branchPublished?: boolean
}

@ApiSchema({ name: 'ListBranchResponse' })
export class ListBranchResponseDto {
  @ApiProperty({ type: [String] })
  branches: string[]
}

@ApiSchema({ name: 'GitCommitInfo' })
export class GitCommitInfoDto {
  @ApiProperty()
  hash: string

  @ApiProperty()
  message: string

  @ApiProperty()
  author: string

  @ApiProperty()
  email: string

  @ApiProperty()
  timestamp: string
}

@ApiSchema({ name: 'ExecuteRequest' })
export class ExecuteRequestDto {
  @ApiProperty()
  command: string

  @ApiPropertyOptional({
    description: 'Current working directory',
  })
  cwd?: string

  @ApiPropertyOptional({
    description: 'Timeout in seconds, defaults to 10 seconds',
  })
  timeout?: number
}

@ApiSchema({ name: 'ExecuteResponse' })
export class ExecuteResponseDto {
  @ApiProperty({
    type: Number,
    description: 'Exit code',
    example: 0,
  })
  exitCode: number

  @ApiProperty({
    type: String,
    description: 'Command output',
    example: 'Command output here',
  })
  result: string
}

@ApiSchema({ name: 'ProjectDirResponse' })
export class ProjectDirResponseDto {
  @ApiPropertyOptional()
  dir?: string
}

@ApiSchema({ name: 'CreateSessionRequest' })
export class CreateSessionRequestDto {
  @ApiProperty({
    description: 'The ID of the session',
    example: 'session-123',
  })
  @IsString()
  sessionId: string
}

@ApiSchema({ name: 'SessionExecuteRequest' })
export class SessionExecuteRequestDto {
  @ApiProperty({
    description: 'The command to execute',
    example: 'ls -la',
  })
  @IsString()
  command: string

  @ApiPropertyOptional({
    description: 'Whether to execute the command asynchronously',
    example: false,
  })
  @IsBoolean()
  @IsOptional()
  runAsync?: boolean

  @ApiPropertyOptional({
    description: 'Deprecated: Use runAsync instead. Whether to execute the command asynchronously',
    example: false,
    deprecated: true,
  })
  @IsBoolean()
  @IsOptional()
  async?: boolean

  constructor(partial: Partial<SessionExecuteRequestDto>) {
    Object.assign(this, partial)
    // Migrate async to runAsync if async is set and runAsync is not set
    if (this.async !== undefined && this.runAsync === undefined) {
      this.runAsync = this.async
    }
  }
}

@ApiSchema({ name: 'SessionExecuteResponse' })
export class SessionExecuteResponseDto {
  @ApiPropertyOptional({
    description: 'The ID of the executed command',
    example: 'cmd-123',
  })
  @IsString()
  @IsOptional()
  cmdId?: string

  @ApiPropertyOptional({
    description: 'The output of the executed command',
    example: 'total 20\ndrwxr-xr-x  4 user group  128 Mar 15 10:30 .',
  })
  @IsString()
  @IsOptional()
  output?: string

  @ApiPropertyOptional({
    description: 'The exit code of the executed command',
    example: 0,
  })
  @IsOptional()
  exitCode?: number
}

@ApiSchema({ name: 'Command' })
export class CommandDto {
  @ApiProperty({
    description: 'The ID of the command',
    example: 'cmd-123',
  })
  @IsString()
  id: string

  @ApiProperty({
    description: 'The command that was executed',
    example: 'ls -la',
  })
  @IsString()
  command: string

  @ApiPropertyOptional({
    description: 'The exit code of the command',
    example: 0,
  })
  @IsOptional()
  exitCode?: number
}

@ApiSchema({ name: 'Session' })
export class SessionDto {
  @ApiProperty({
    description: 'The ID of the session',
    example: 'session-123',
  })
  @IsString()
  sessionId: string

  @ApiProperty({
    description: 'The list of commands executed in this session',
    type: [CommandDto],
    nullable: true,
  })
  @IsArray()
  @IsOptional()
  commands?: CommandDto[] | null
}
