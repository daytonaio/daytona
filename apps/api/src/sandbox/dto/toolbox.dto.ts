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

@ApiSchema({ name: 'SearchFilesResponse' })
export class SearchFilesResponseDto {
  @ApiProperty({ type: [String] })
  files: string[]
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

// Computer Use DTOs
@ApiSchema({ name: 'MousePosition' })
export class MousePositionDto {
  @ApiProperty()
  x: number

  @ApiProperty()
  y: number
}

@ApiSchema({ name: 'MouseMoveRequest' })
export class MouseMoveRequestDto {
  @ApiProperty()
  x: number

  @ApiProperty()
  y: number
}

@ApiSchema({ name: 'MouseMoveResponse' })
export class MouseMoveResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  x: number

  @ApiProperty()
  y: number

  @ApiProperty()
  actual_x: number

  @ApiProperty()
  actual_y: number
}

@ApiSchema({ name: 'MouseClickRequest' })
export class MouseClickRequestDto {
  @ApiProperty()
  x: number

  @ApiProperty()
  y: number

  @ApiPropertyOptional()
  button?: string

  @ApiPropertyOptional()
  double?: boolean
}

@ApiSchema({ name: 'MouseClickResponse' })
export class MouseClickResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  action: string

  @ApiProperty()
  button: string

  @ApiProperty()
  double: boolean

  @ApiProperty()
  x: number

  @ApiProperty()
  y: number

  @ApiProperty()
  actual_x: number

  @ApiProperty()
  actual_y: number
}

@ApiSchema({ name: 'MouseDragRequest' })
export class MouseDragRequestDto {
  @ApiProperty()
  startX: number

  @ApiProperty()
  startY: number

  @ApiProperty()
  endX: number

  @ApiProperty()
  endY: number

  @ApiPropertyOptional()
  button?: string
}

@ApiSchema({ name: 'MouseDragResponse' })
export class MouseDragResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  action: string

  @ApiProperty()
  from: { x: number; y: number }

  @ApiProperty()
  to: { x: number; y: number }

  @ApiProperty()
  actual_x: number

  @ApiProperty()
  actual_y: number
}

@ApiSchema({ name: 'MouseScrollRequest' })
export class MouseScrollRequestDto {
  @ApiProperty()
  x: number

  @ApiProperty()
  y: number

  @ApiProperty()
  direction: string

  @ApiPropertyOptional()
  amount?: number
}

@ApiSchema({ name: 'MouseScrollResponse' })
export class MouseScrollResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  action: string

  @ApiProperty()
  direction: string

  @ApiProperty()
  amount: number

  @ApiProperty()
  x: number

  @ApiProperty()
  y: number
}

@ApiSchema({ name: 'KeyboardTypeRequest' })
export class KeyboardTypeRequestDto {
  @ApiProperty()
  text: string

  @ApiPropertyOptional()
  delay?: number
}

@ApiSchema({ name: 'KeyboardTypeResponse' })
export class KeyboardTypeResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  typed: string
}

@ApiSchema({ name: 'KeyboardPressRequest' })
export class KeyboardPressRequestDto {
  @ApiProperty()
  key: string

  @ApiPropertyOptional({ type: [String] })
  modifiers?: string[]
}

@ApiSchema({ name: 'KeyboardPressResponse' })
export class KeyboardPressResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  key: string

  @ApiProperty()
  modifiers: string[]
}

@ApiSchema({ name: 'KeyboardHotkeyRequest' })
export class KeyboardHotkeyRequestDto {
  @ApiProperty()
  keys: string
}

@ApiSchema({ name: 'KeyboardHotkeyResponse' })
export class KeyboardHotkeyResponseDto {
  @ApiProperty()
  success: boolean

  @ApiProperty()
  hotkey: string
}

@ApiSchema({ name: 'ScreenshotResponse' })
export class ScreenshotResponseDto {
  @ApiProperty()
  screenshot: string

  @ApiProperty()
  width: number

  @ApiProperty()
  height: number

  @ApiPropertyOptional()
  cursor_position?: { x: number; y: number }

  @ApiPropertyOptional()
  region?: { x: number; y: number; width: number; height: number }

  @ApiPropertyOptional()
  format?: string

  @ApiPropertyOptional()
  quality?: number

  @ApiPropertyOptional()
  scale?: number

  @ApiPropertyOptional()
  size_bytes?: number
}

@ApiSchema({ name: 'RegionScreenshotRequest' })
export class RegionScreenshotRequestDto {
  @ApiProperty()
  x: number

  @ApiProperty()
  y: number

  @ApiProperty()
  width: number

  @ApiProperty()
  height: number
}

@ApiSchema({ name: 'RegionScreenshotResponse' })
export class RegionScreenshotResponseDto {
  @ApiProperty()
  screenshot: string

  @ApiProperty()
  region: { x: number; y: number; width: number; height: number }

  @ApiPropertyOptional()
  cursor_position?: { x: number; y: number }
}

@ApiSchema({ name: 'CompressedScreenshotResponse' })
export class CompressedScreenshotResponseDto {
  @ApiProperty()
  screenshot: string

  @ApiProperty()
  width: number

  @ApiProperty()
  height: number

  @ApiProperty()
  format: string

  @ApiProperty()
  quality: number

  @ApiProperty()
  scale: number

  @ApiProperty()
  size_bytes: number

  @ApiPropertyOptional()
  cursor_position?: { x: number; y: number }
}

@ApiSchema({ name: 'DisplayInfoResponse' })
export class DisplayInfoResponseDto {
  @ApiProperty({ type: [Object] })
  displays: Array<{ id: number; x: number; y: number; width: number; height: number; is_active: boolean }>
}

@ApiSchema({ name: 'WindowsResponse' })
export class WindowsResponseDto {
  @ApiProperty({ type: [Object] })
  windows: Array<{ id: number; title: string }>

  @ApiProperty()
  count: number
}

// Computer Use Management Response DTOs
@ApiSchema({ name: 'EmptyResponse' })
export class EmptyResponseDto {
  @ApiProperty()
  success: boolean
}

@ApiSchema({ name: 'ComputerUseStartResponse' })
export class ComputerUseStartResponseDto {
  @ApiProperty()
  message: string

  @ApiProperty({ type: Object })
  status: Record<string, any>
}

@ApiSchema({ name: 'ComputerUseStopResponse' })
export class ComputerUseStopResponseDto {
  @ApiProperty()
  message: string

  @ApiProperty({ type: Object })
  status: Record<string, any>
}

@ApiSchema({ name: 'ComputerUseStatusResponse' })
export class ComputerUseStatusResponseDto {
  @ApiProperty({ type: Object })
  status: Record<string, any>
}

@ApiSchema({ name: 'ProcessStatusResponse' })
export class ProcessStatusResponseDto {
  @ApiProperty()
  processName: string

  @ApiProperty()
  running: boolean
}

@ApiSchema({ name: 'ProcessRestartResponse' })
export class ProcessRestartResponseDto {
  @ApiProperty()
  message: string

  @ApiProperty()
  processName: string
}

@ApiSchema({ name: 'ProcessLogsResponse' })
export class ProcessLogsResponseDto {
  @ApiProperty()
  processName: string

  @ApiProperty()
  logs: string
}

@ApiSchema({ name: 'ProcessErrorsResponse' })
export class ProcessErrorsResponseDto {
  @ApiProperty()
  processName: string

  @ApiProperty()
  errors: string
}
