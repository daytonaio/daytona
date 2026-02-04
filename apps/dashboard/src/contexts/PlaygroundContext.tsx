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
} from '@daytonaio/sdk'
import {
  KeyboardHotKey,
  KeyboardPress,
  KeyboardType,
  MouseClick,
  MouseDrag,
  MouseMove,
  MouseScroll,
  CustomizedScreenshotOptions,
  PlaygroundActions,
  ParameterFormData,
  PlaygroundActionFormDataBasic,
  PlaygroundActionWithParamsFormData,
  GitCloneParams,
  GitStatusParams,
  GitBranchesParams,
  CodeRunParams,
  ShellCommandRunParams,
  ParameterFormItem,
  ListFilesParams,
  CreateFolderParams,
  DeleteFileParams,
} from '@/enums/Playground'
import { UsePlaygroundSandboxResult } from '@/hooks/usePlaygroundSandbox'
import { createContext, ReactNode } from 'react'

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
