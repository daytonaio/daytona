/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import PythonIcon from '@/assets/python.svg'
import TypescriptIcon from '@/assets/typescript.svg'
import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import { CodeLanguage, ScreenshotOptions, ScreenshotRegion, ComputerUse } from '@daytonaio/sdk'

export enum PlaygroundCategories {
  SANDBOX = 'sandbox',
  TERMINAL = 'terminal',
  VNC = 'vnc',
}

export const playgroundCategoriesData = [
  { value: PlaygroundCategories.SANDBOX, label: 'Sandbox' },
  { value: PlaygroundCategories.TERMINAL, label: 'Terminal' },
  { value: PlaygroundCategories.VNC, label: 'VNC' },
]

export interface ParameterFormItem {
  label: string
  placeholder: string
  key: string
  required?: boolean
}

export interface NumberParameterFormItem extends ParameterFormItem {
  min: number
  max: number
  step?: number
}

// keyof (A | B | C) gives intersections of types i.e. type = common properties to A,B,C
// KeysOf gives us keyof A | keyof B | keyof C behaviour
export type KeysOf<T> = T extends any ? keyof T : never

export type ParameterFormData<T> = ((ParameterFormItem | NumberParameterFormItem) & { key: KeysOf<T> })[]

export enum SandboxParametersSections {
  SANDBOX_MANAGMENT = 'sandbox_managment',
  FILE_SYSTEM = 'file_system',
  GIT_OPERATIONS = 'git_operations',
  PROCESS_CODE_EXECUTION = 'process_code_execution',
}

export const sandboxParametersSectionsData = [
  { value: SandboxParametersSections.SANDBOX_MANAGMENT, label: 'Managment' },
  { value: SandboxParametersSections.FILE_SYSTEM, label: 'File System' },
  { value: SandboxParametersSections.GIT_OPERATIONS, label: 'Git Operations' },
  { value: SandboxParametersSections.PROCESS_CODE_EXECUTION, label: 'Process & Code Execution' },
]

export const codeSnippetSupportedLanguages = [
  { value: CodeLanguage.PYTHON, label: 'Python', icon: PythonIcon },
  { value: CodeLanguage.TYPESCRIPT, label: 'TypeScript', icon: TypescriptIcon },
]

export enum VNCInteractionOptionsSections {
  DISPLAY = 'display',
  KEYBOARD = 'keyboard',
  MOUSE = 'mouse',
  SCREENSHOT = 'screenshot',
}

export type WrapVNCInvokeApiType = (
  invokeApi: PlaygroundActionInvokeApi,
) => <A, T>(
  actionFormData: PlaygroundActionFormDataBasic<A> | PlaygroundActionWithParamsFormData<A, T>,
) => Promise<void>

export type VNCInteractionOptionsSectionComponentProps = {
  disableActions: boolean
  ComputerUseClient: ComputerUse | null
  wrapVNCInvokeApi: WrapVNCInvokeApiType
}

export const VNCInteractionOptionsSectionsData = [
  { value: VNCInteractionOptionsSections.DISPLAY, label: 'Display' },
  { value: VNCInteractionOptionsSections.KEYBOARD, label: 'Keyboard' },
  { value: VNCInteractionOptionsSections.MOUSE, label: 'Mouse' },
  { value: VNCInteractionOptionsSections.SCREENSHOT, label: 'Screenshot' },
]

export enum DisplayActions {
  GET_INFO = 'getInfo',
  GET_WINDOWS = 'getWindows',
}

export type KeyboardHotKey = {
  keys: string
}

export type KeyboardPress = {
  key: string
  modifiers?: string
}

export type KeyboardType = {
  text: string
  delay?: number
}

export enum KeyboardActions {
  HOTKEY = 'hotkey',
  PRESS = 'press',
  TYPE = 'type',
}

export type KeyboardActionFormData<T extends KeyboardHotKey | KeyboardPress | KeyboardType> =
  PlaygroundActionWithParamsFormData<KeyboardActions, T>

export enum MouseButton {
  LEFT = 'left',
  RIGHT = 'right',
  MIDDLE = 'middle',
}

export enum MouseScrollDirection {
  UP = 'up',
  DOWN = 'down',
}

export type MouseClick = {
  x: number
  y: number
  button?: MouseButton
  double?: boolean
}

export type MouseDrag = {
  startX: number
  startY: number
  endX: number
  endY: number
  button?: MouseButton
}

export type MouseMove = {
  x: number
  y: number
}

export type MouseScroll = {
  x: number
  y: number
  direction: MouseScrollDirection
  amount?: number
}

export enum MouseActions {
  CLICK = 'click',
  DRAG = 'drag',
  MOVE = 'move',
  SCROLL = 'scroll',
  GET_POSITION = 'getPosition',
}

export type MouseActionFormData<T extends MouseClick | MouseDrag | MouseMove | MouseScroll> =
  PlaygroundActionWithParamsFormData<MouseActions, T>

export enum ScreenshotFormatOption {
  JPEG = 'jpeg',
  PNG = 'png',
  WEBP = 'webp',
}

export interface CustomizedScreenshotOptions extends Omit<ScreenshotOptions, 'format'> {
  format?: ScreenshotFormatOption
}

// Form data structure for actions which don't require any parameters for their execution
export interface PlaygroundActionFormDataBasic<A> {
  label: string
  description: string
  methodName: A
  onChangeParamsValidationDisabled?: boolean
}

// Form data structure for actions which use certain parameters for their execution
export type PlaygroundActionWithParamsFormData<A, T> = PlaygroundActionFormDataBasic<A> & {
  parametersFormItems: ParameterFormData<T>
  parametersState: T
}

export enum ScreenshotActions {
  TAKE_COMPRESSED = 'takeCompressed',
  TAKE_COMPRESSED_REGION = 'takeCompressedRegion',
  TAKE_FULL_SCREEN = 'takeFullScreen',
  TAKE_REGION = 'takeRegion',
}

export type ScreenshotActionFormData<T extends ScreenshotRegion | CustomizedScreenshotOptions> =
  PlaygroundActionWithParamsFormData<ScreenshotActions, T>

export type ListFilesParams = {
  directoryPath: string
}

export type CreateFolderParams = {
  folderDestinationPath: string
  permissions: string
}

export type DeleteFileParams = {
  filePath: string
  recursive?: boolean
}

export enum FileSystemActions {
  LIST_FILES = 'listFiles',
  CREATE_FOLDER = 'createFolder',
  DELETE_FILE = 'deleteFile',
}

export type FileSystemActionFormData<T extends ListFilesParams | CreateFolderParams | DeleteFileParams> =
  PlaygroundActionWithParamsFormData<FileSystemActions, T>

export type GitCloneParams = {
  repositoryURL: string
  cloneDestinationPath: string
  branchToClone?: string
  commitToClone?: string
  authUsername?: string
  authPassword?: string
}

export type GitStatusParams = {
  repositoryPath: string
}

export type GitBranchesParams = {
  repositoryPath: string
}

export enum GitOperationsActions {
  GIT_CLONE = 'clone',
  GIT_STATUS = 'status',
  GIT_BRANCHES_LIST = 'branches',
}

export type GitOperationsActionFormData<T extends GitCloneParams | GitStatusParams | GitBranchesParams> =
  PlaygroundActionWithParamsFormData<GitOperationsActions, T>

export enum ProcessCodeExecutionActions {
  CODE_RUN = 'codeRun',
  SHELL_COMMANDS_RUN = 'executeCommand',
}

export type CodeRunParams = {
  languageCode?: string
}

export type ShellCommandRunParams = {
  shellCommand?: string
}

export type ProcessCodeExecutionOperationsActionFormData<T extends CodeRunParams | ShellCommandRunParams> =
  PlaygroundActionWithParamsFormData<ProcessCodeExecutionActions, T>

export type SandboxCodeSnippetsActions = FileSystemActions | GitOperationsActions | ProcessCodeExecutionActions

export type VNCInteractionActions = DisplayActions | KeyboardActions | MouseActions | ScreenshotActions

// Actions enums values represent method names for TypeScript SDK
export type PlaygroundActions = VNCInteractionActions | SandboxCodeSnippetsActions
