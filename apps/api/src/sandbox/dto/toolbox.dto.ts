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
  @ApiProperty({
    description: 'The X coordinate of the mouse cursor position',
    example: 100,
  })
  x: number

  @ApiProperty({
    description: 'The Y coordinate of the mouse cursor position',
    example: 200,
  })
  y: number
}

@ApiSchema({ name: 'MouseMoveRequest' })
export class MouseMoveRequestDto {
  @ApiProperty({
    description: 'The target X coordinate to move the mouse cursor to',
    example: 150,
  })
  x: number

  @ApiProperty({
    description: 'The target Y coordinate to move the mouse cursor to',
    example: 250,
  })
  y: number
}

@ApiSchema({ name: 'MouseMoveResponse' })
export class MouseMoveResponseDto {
  @ApiProperty({
    description: 'The actual X coordinate where the mouse cursor ended up',
    example: 150,
  })
  x: number

  @ApiProperty({
    description: 'The actual Y coordinate where the mouse cursor ended up',
    example: 250,
  })
  y: number
}

@ApiSchema({ name: 'MouseClickRequest' })
export class MouseClickRequestDto {
  @ApiProperty({
    description: 'The X coordinate where to perform the mouse click',
    example: 100,
  })
  x: number

  @ApiProperty({
    description: 'The Y coordinate where to perform the mouse click',
    example: 200,
  })
  y: number

  @ApiPropertyOptional({
    description: 'The mouse button to click (left, right, middle). Defaults to left',
    example: 'left',
  })
  button?: string

  @ApiPropertyOptional({
    description: 'Whether to perform a double-click instead of a single click',
    example: false,
  })
  double?: boolean
}

@ApiSchema({ name: 'MouseClickResponse' })
export class MouseClickResponseDto {
  @ApiProperty({
    description: 'The actual X coordinate where the click occurred',
    example: 100,
  })
  x: number

  @ApiProperty({
    description: 'The actual Y coordinate where the click occurred',
    example: 200,
  })
  y: number
}

@ApiSchema({ name: 'MouseDragRequest' })
export class MouseDragRequestDto {
  @ApiProperty({
    description: 'The starting X coordinate for the drag operation',
    example: 100,
  })
  startX: number

  @ApiProperty({
    description: 'The starting Y coordinate for the drag operation',
    example: 200,
  })
  startY: number

  @ApiProperty({
    description: 'The ending X coordinate for the drag operation',
    example: 300,
  })
  endX: number

  @ApiProperty({
    description: 'The ending Y coordinate for the drag operation',
    example: 400,
  })
  endY: number

  @ApiPropertyOptional({
    description: 'The mouse button to use for dragging (left, right, middle). Defaults to left',
    example: 'left',
  })
  button?: string
}

@ApiSchema({ name: 'MouseDragResponse' })
export class MouseDragResponseDto {
  @ApiProperty({
    description: 'The actual X coordinate where the drag ended',
    example: 300,
  })
  x: number

  @ApiProperty({
    description: 'The actual Y coordinate where the drag ended',
    example: 400,
  })
  y: number
}

@ApiSchema({ name: 'MouseScrollRequest' })
export class MouseScrollRequestDto {
  @ApiProperty({
    description: 'The X coordinate where to perform the scroll operation',
    example: 100,
  })
  x: number

  @ApiProperty({
    description: 'The Y coordinate where to perform the scroll operation',
    example: 200,
  })
  y: number

  @ApiProperty({
    description: 'The scroll direction (up, down)',
    example: 'down',
  })
  direction: string

  @ApiPropertyOptional({
    description: 'The number of scroll units to scroll. Defaults to 1',
    example: 3,
  })
  amount?: number
}

@ApiSchema({ name: 'MouseScrollResponse' })
export class MouseScrollResponseDto {
  @ApiProperty({
    description: 'Whether the mouse scroll operation was successful',
    example: true,
  })
  success: boolean
}

@ApiSchema({ name: 'KeyboardTypeRequest' })
export class KeyboardTypeRequestDto {
  @ApiProperty({
    description: 'The text to type using the keyboard',
    example: 'Hello, World!',
  })
  text: string

  @ApiPropertyOptional({
    description: 'Delay in milliseconds between keystrokes. Defaults to 0',
    example: 100,
  })
  delay?: number
}

@ApiSchema({ name: 'KeyboardPressRequest' })
export class KeyboardPressRequestDto {
  @ApiProperty({
    description: 'The key to press (e.g., a, b, c, enter, space, etc.)',
    example: 'enter',
  })
  key: string

  @ApiPropertyOptional({
    description: 'Array of modifier keys to press along with the main key (ctrl, alt, shift, cmd)',
    type: [String],
    example: ['ctrl', 'shift'],
  })
  modifiers?: string[]
}

@ApiSchema({ name: 'KeyboardHotkeyRequest' })
export class KeyboardHotkeyRequestDto {
  @ApiProperty({
    description: 'The hotkey combination to press (e.g., "ctrl+c", "cmd+v", "alt+tab")',
    example: 'ctrl+c',
  })
  keys: string
}

@ApiSchema({ name: 'ScreenshotResponse' })
export class ScreenshotResponseDto {
  @ApiProperty({
    description: 'Base64 encoded screenshot image data',
    example: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
  })
  screenshot: string

  @ApiPropertyOptional({
    description: 'The current cursor position when the screenshot was taken',
    example: { x: 500, y: 300 },
  })
  cursorPosition?: { x: number; y: number }

  @ApiPropertyOptional({
    description: 'The size of the screenshot data in bytes',
    example: 24576,
  })
  sizeBytes?: number
}

@ApiSchema({ name: 'RegionScreenshotRequest' })
export class RegionScreenshotRequestDto {
  @ApiProperty({
    description: 'The X coordinate of the top-left corner of the region to capture',
    example: 100,
  })
  x: number

  @ApiProperty({
    description: 'The Y coordinate of the top-left corner of the region to capture',
    example: 100,
  })
  y: number

  @ApiProperty({
    description: 'The width of the region to capture in pixels',
    example: 800,
  })
  width: number

  @ApiProperty({
    description: 'The height of the region to capture in pixels',
    example: 600,
  })
  height: number
}

@ApiSchema({ name: 'RegionScreenshotResponse' })
export class RegionScreenshotResponseDto {
  @ApiProperty({
    description: 'Base64 encoded screenshot image data of the specified region',
    example: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
  })
  screenshot: string

  @ApiPropertyOptional({
    description: 'The current cursor position when the region screenshot was taken',
    example: { x: 500, y: 300 },
  })
  cursorPosition?: { x: number; y: number }

  @ApiPropertyOptional({
    description: 'The size of the screenshot data in bytes',
    example: 24576,
  })
  sizeBytes?: number
}

@ApiSchema({ name: 'CompressedScreenshotResponse' })
export class CompressedScreenshotResponseDto {
  @ApiProperty({
    description: 'Base64 encoded compressed screenshot image data',
    example: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
  })
  screenshot: string

  @ApiPropertyOptional({
    description: 'The current cursor position when the compressed screenshot was taken',
    example: { x: 250, y: 150 },
  })
  cursorPosition?: { x: number; y: number }

  @ApiPropertyOptional({
    description: 'The size of the compressed screenshot data in bytes',
    example: 12288,
  })
  sizeBytes?: number
}

@ApiSchema({ name: 'DisplayInfoResponse' })
export class DisplayInfoResponseDto {
  @ApiProperty({
    description: 'Array of display information for all connected displays',
    type: [Object],
    example: [
      {
        id: 0,
        x: 0,
        y: 0,
        width: 1920,
        height: 1080,
        is_active: true,
      },
    ],
  })
  displays: Array<{ id: number; x: number; y: number; width: number; height: number; is_active: boolean }>
}

@ApiSchema({ name: 'WindowsResponse' })
export class WindowsResponseDto {
  @ApiProperty({
    description: 'Array of window information for all visible windows',
    type: [Object],
    example: [
      {
        id: 12345,
        title: 'Terminal',
      },
    ],
  })
  windows: Array<{ id: number; title: string }>

  @ApiProperty({
    description: 'The total number of windows found',
    example: 5,
  })
  count: number
}

// Computer Use Management Response DTOs

@ApiSchema({ name: 'ComputerUseStartResponse' })
export class ComputerUseStartResponseDto {
  @ApiProperty({
    description: 'A message indicating the result of starting computer use processes',
    example: 'Computer use processes started successfully',
  })
  message: string

  @ApiProperty({
    description: 'Status information about all VNC desktop processes after starting',
    type: Object,
    example: {
      xvfb: { running: true, priority: 100, autoRestart: true, pid: 12345 },
      xfce4: { running: true, priority: 200, autoRestart: true, pid: 12346 },
      x11vnc: { running: true, priority: 300, autoRestart: true, pid: 12347 },
      novnc: { running: true, priority: 400, autoRestart: true, pid: 12348 },
    },
  })
  status: Record<string, any>
}

@ApiSchema({ name: 'ComputerUseStopResponse' })
export class ComputerUseStopResponseDto {
  @ApiProperty({
    description: 'A message indicating the result of stopping computer use processes',
    example: 'Computer use processes stopped successfully',
  })
  message: string

  @ApiProperty({
    description: 'Status information about all VNC desktop processes after stopping',
    type: Object,
    example: {
      xvfb: { running: false, priority: 100, autoRestart: true },
      xfce4: { running: false, priority: 200, autoRestart: true },
      x11vnc: { running: false, priority: 300, autoRestart: true },
      novnc: { running: false, priority: 400, autoRestart: true },
    },
  })
  status: Record<string, any>
}

@ApiSchema({ name: 'ComputerUseStatusResponse' })
export class ComputerUseStatusResponseDto {
  @ApiProperty({
    description: 'Status of computer use services (active, partial, inactive, error)',
    example: 'active',
    enum: ['active', 'partial', 'inactive', 'error'],
  })
  status: string
}

@ApiSchema({ name: 'ProcessStatusResponse' })
export class ProcessStatusResponseDto {
  @ApiProperty({
    description: 'The name of the VNC process being checked',
    example: 'xfce4',
  })
  processName: string

  @ApiProperty({
    description: 'Whether the specified VNC process is currently running',
    example: true,
  })
  running: boolean
}

@ApiSchema({ name: 'ProcessRestartResponse' })
export class ProcessRestartResponseDto {
  @ApiProperty({
    description: 'A message indicating the result of restarting the process',
    example: 'Process xfce4 restarted successfully',
  })
  message: string

  @ApiProperty({
    description: 'The name of the VNC process that was restarted',
    example: 'xfce4',
  })
  processName: string
}

@ApiSchema({ name: 'ProcessLogsResponse' })
export class ProcessLogsResponseDto {
  @ApiProperty({
    description: 'The name of the VNC process whose logs were retrieved',
    example: 'novnc',
  })
  processName: string

  @ApiProperty({
    description: 'The log output from the specified VNC process',
    example: '2024-01-15 10:30:45 [INFO] NoVNC server started on port 6080',
  })
  logs: string
}

@ApiSchema({ name: 'ProcessErrorsResponse' })
export class ProcessErrorsResponseDto {
  @ApiProperty({
    description: 'The name of the VNC process whose error logs were retrieved',
    example: 'x11vnc',
  })
  processName: string

  @ApiProperty({
    description: 'The error log output from the specified VNC process',
    example: '2024-01-15 10:30:45 [ERROR] Failed to bind to port 5901',
  })
  errors: string
}
