/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeLanguage, Resources, CreateSandboxBaseParams, ScreenshotRegion } from '@daytonaio/sdk'
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
} from '@/enums/Playground'
import { createContext } from 'react'

export interface SandboxParams {
  language?: CodeLanguage
  resources: Resources
  createSandboxBaseParams: CreateSandboxBaseParams
  languageCodeToRun?: string
  shellCodeToRun?: string
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
  responseText?: string
}

export type SetVNCInteractionOptionsParamValue = <K extends keyof VNCInteractionOptionsParams>(
  key: K,
  value: VNCInteractionOptionsParams[K],
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

export interface IPlaygroundContext {
  sandboxParametersState: SandboxParams
  setSandboxParameterValue: SetSandboxParamsValue
  VNCInteractionOptionsParamsState: VNCInteractionOptionsParams
  setVNCInteractionOptionsParamValue: SetVNCInteractionOptionsParamValue
  runPlaygroundActionWithParams: RunPlaygroundActionWithParams
  runPlaygroundActionWithoutParams: RunPlaygroundActionBasic
  runningActionMethod: RunningActionMethodName
  actionRuntimeError: ActionRuntimeError
}

export const PlaygroundContext = createContext<IPlaygroundContext | null>(null)
