/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum PlaygroundCategories {
  SANDBOX = 'sandbox',
  TERMINAL = 'terminal',
  VNC = 'vnc',
}

export enum SandboxParametersSections {
  SANDBOX_MANAGEMENT = 'sandbox_management',
  FILE_SYSTEM = 'file_system',
  GIT_OPERATIONS = 'git_operations',
  PROCESS_CODE_EXECUTION = 'process_code_execution',
}

export enum VNCInteractionOptionsSections {
  DISPLAY = 'display',
  KEYBOARD = 'keyboard',
  MOUSE = 'mouse',
  SCREENSHOT = 'screenshot',
}

export enum DisplayActions {
  GET_INFO = 'getInfo',
  GET_WINDOWS = 'getWindows',
}

export enum KeyboardActions {
  HOTKEY = 'hotkey',
  PRESS = 'press',
  TYPE = 'type',
}

export enum MouseButton {
  LEFT = 'left',
  RIGHT = 'right',
  MIDDLE = 'middle',
}

export enum MouseScrollDirection {
  UP = 'up',
  DOWN = 'down',
}

export enum MouseActions {
  CLICK = 'click',
  DRAG = 'drag',
  MOVE = 'move',
  SCROLL = 'scroll',
  GET_POSITION = 'getPosition',
}

export enum ScreenshotFormatOption {
  JPEG = 'jpeg',
  PNG = 'png',
  WEBP = 'webp',
}

export enum ScreenshotActions {
  TAKE_COMPRESSED = 'takeCompressed',
  TAKE_COMPRESSED_REGION = 'takeCompressedRegion',
  TAKE_FULL_SCREEN = 'takeFullScreen',
  TAKE_REGION = 'takeRegion',
}

export enum FileSystemActions {
  LIST_FILES = 'listFiles',
  CREATE_FOLDER = 'createFolder',
  DELETE_FILE = 'deleteFile',
}

export enum GitOperationsActions {
  GIT_CLONE = 'clone',
  GIT_STATUS = 'status',
  GIT_BRANCHES_LIST = 'branches',
}

export enum ProcessCodeExecutionActions {
  CODE_RUN = 'codeRun',
  SHELL_COMMANDS_RUN = 'executeCommand',
}

export type SandboxCodeSnippetsActions = FileSystemActions | GitOperationsActions | ProcessCodeExecutionActions

export type VNCInteractionActions = DisplayActions | KeyboardActions | MouseActions | ScreenshotActions

// Actions enums values represent method names for TypeScript SDK
export type PlaygroundActions = VNCInteractionActions | SandboxCodeSnippetsActions
