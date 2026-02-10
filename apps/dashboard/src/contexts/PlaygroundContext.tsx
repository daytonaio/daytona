/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  CodeLanguage,
  Resources,
  CreateSandboxBaseParams,
  ScreenshotRegion,
  Daytona,
  Sandbox,
  CreateSandboxFromImageParams,
  CreateSandboxFromSnapshotParams,
  ScreenshotOptions,
  ComputerUse,
} from '@daytonaio/sdk'
import {
  FileSystemActions,
  GitOperationsActions,
  KeyboardActions,
  MouseActions,
  MouseButton,
  MouseScrollDirection,
  PlaygroundActions,
  ProcessCodeExecutionActions,
  ScreenshotActions,
  ScreenshotFormatOption,
} from '@/enums/Playground'
import { UsePlaygroundSandboxResult } from '@/hooks/usePlaygroundSandbox'
import { createContext, ReactNode } from 'react'

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

// --- VNC param types ---

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

export interface CustomizedScreenshotOptions extends Omit<ScreenshotOptions, 'format'> {
  format?: ScreenshotFormatOption
}

// --- VNC component types ---

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

// --- Action-specific form data types ---

export type KeyboardActionFormData<T extends KeyboardHotKey | KeyboardPress | KeyboardType> =
  PlaygroundActionWithParamsFormData<KeyboardActions, T>

export type MouseActionFormData<T extends MouseClick | MouseDrag | MouseMove | MouseScroll> =
  PlaygroundActionWithParamsFormData<MouseActions, T>

export type ScreenshotActionFormData<T extends ScreenshotRegion | CustomizedScreenshotOptions> =
  PlaygroundActionWithParamsFormData<ScreenshotActions, T>

// --- Sandbox param types ---

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

export type CodeRunParams = {
  languageCode?: string
}

export type ShellCommandRunParams = {
  shellCommand?: string
}

export type FileSystemActionFormData<T extends ListFilesParams | CreateFolderParams | DeleteFileParams> =
  PlaygroundActionWithParamsFormData<FileSystemActions, T>

export type GitOperationsActionFormData<T extends GitCloneParams | GitStatusParams | GitBranchesParams> =
  PlaygroundActionWithParamsFormData<GitOperationsActions, T>

export type ProcessCodeExecutionOperationsActionFormData<T extends CodeRunParams | ShellCommandRunParams> =
  PlaygroundActionWithParamsFormData<ProcessCodeExecutionActions, T>

export interface SandboxParams {
  language?: CodeLanguage
  snapshotName?: string
  resources: Resources
  createSandboxBaseParams: CreateSandboxBaseParams
  // File system operations params
  listFilesParams: ListFilesParams
  createFolderParams: CreateFolderParams
  deleteFileParams: DeleteFileParams
  // Git operations params
  gitCloneParams: GitCloneParams
  gitStatusParams: GitStatusParams
  gitBranchesParams: GitBranchesParams
  // Process and Code Execution params
  codeRunParams: CodeRunParams
  shellCommandRunParams: ShellCommandRunParams
}

export type SetSandboxParamsValue = <K extends keyof SandboxParams>(key: K, value: SandboxParams[K]) => void

export interface VNCInteractionOptionsParams {
  keyboardHotKeyParams: KeyboardHotKey
  keyboardPressParams: KeyboardPress
  keyboardTypeParams: KeyboardType
  mouseClickParams: MouseClick
  mouseDragParams: MouseDrag
  mouseMoveParams: MouseMove
  mouseScrollParams: MouseScroll
  screenshotOptionsConfig: CustomizedScreenshotOptions
  screenshotRegionConfig: ScreenshotRegion
  responseContent?: string | ReactNode
  VNCSandboxData?: UsePlaygroundSandboxResult
  VNCUrl: string | null
}

export type SetVNCInteractionOptionsParamValue = <K extends keyof VNCInteractionOptionsParams>(
  key: K,
  value: VNCInteractionOptionsParams[K],
) => void

export type PlaygroundActionParams = SandboxParams & VNCInteractionOptionsParams

export type SetPlaygroundActionParamValue = <K extends keyof PlaygroundActionParams>(
  key: K,
  value: PlaygroundActionParams[K],
) => void

// Currently running action, or none
export type RunningActionMethodName = PlaygroundActions | null

// Mapping between action and runtime error message (if any)
export type ActionRuntimeError = Partial<Record<PlaygroundActions, string>>

// Method for validation of required params for a given action
export type ValidatePlaygroundActionRequiredParams = <T>(
  actionParamsFormData: ParameterFormData<T>,
  actionParamsState: T,
) => string | undefined

// Basic method which runs an action that has no params
export type RunPlaygroundActionBasic = <A extends PlaygroundActions>(
  actionFormData: PlaygroundActionFormDataBasic<A>,
  invokeApi: PlaygroundActionInvokeApi,
) => Promise<void>

// Runs an action that requires params
export type RunPlaygroundActionWithParams = <A extends PlaygroundActions, T>(
  actionFormData: PlaygroundActionWithParamsFormData<A, T>,
  invokeApi: PlaygroundActionInvokeApi,
) => Promise<void>

export type PlaygroundActionInvokeApi = <A, T>(
  actionFormData: PlaygroundActionFormDataBasic<A> | PlaygroundActionWithParamsFormData<A, T>,
) => Promise<void>

export type ValidatePlaygroundActionWithParams = <A extends PlaygroundActions, T>(
  actionFormData: PlaygroundActionWithParamsFormData<A, T>,
  parametersState: T,
) => void

export type PlaygroundActionParamValueSetter = <A extends PlaygroundActions, T, S extends T>( // S type is one of types contained inside T union
  actionFormData: PlaygroundActionWithParamsFormData<A, T>,
  paramFormData: ParameterFormItem,
  setState: React.Dispatch<React.SetStateAction<S>>,
  actionParamsKey: keyof PlaygroundActionParams,
  value: any,
) => void

export type SandboxParametersInfo = {
  useLanguageParam: boolean
  useResources: boolean
  useResourcesCPU: boolean
  useResourcesMemory: boolean
  useResourcesDisk: boolean
  createSandboxParamsExist: boolean
  useAutoStopInterval: boolean
  useAutoArchiveInterval: boolean
  useAutoDeleteInterval: boolean
  useSandboxCreateParams: boolean
  useCustomSandboxSnapshotName: boolean
  createSandboxFromImage: boolean
  createSandboxFromSnapshot: boolean
  createSandboxParams: CreateSandboxBaseParams | CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams
}

export interface IPlaygroundContext {
  sandboxParametersState: SandboxParams
  setSandboxParameterValue: SetSandboxParamsValue
  VNCInteractionOptionsParamsState: VNCInteractionOptionsParams
  setVNCInteractionOptionsParamValue: SetVNCInteractionOptionsParamValue
  runPlaygroundActionWithParams: RunPlaygroundActionWithParams
  runPlaygroundActionWithoutParams: RunPlaygroundActionBasic
  validatePlaygroundActionWithParams: ValidatePlaygroundActionWithParams
  playgroundActionParamValueSetter: PlaygroundActionParamValueSetter
  runningActionMethod: RunningActionMethodName
  actionRuntimeError: ActionRuntimeError
  DaytonaClient: Daytona | null
  sandbox: Sandbox | null
  setSandbox: React.Dispatch<React.SetStateAction<Sandbox | null>>
  getSandboxParametersInfo: () => SandboxParametersInfo
}

export const PlaygroundContext = createContext<IPlaygroundContext | null>(null)
