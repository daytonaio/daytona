/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  PlaygroundContext,
  SandboxParams,
  SetSandboxParamsValue,
  VNCInteractionOptionsParams,
  SetVNCInteractionOptionsParamValue,
  RunningActionMethodName,
  ActionRuntimeError,
  ValidatePlaygroundActionRequiredParams,
  RunPlaygroundActionBasic,
  RunPlaygroundActionWithParams,
} from '@/contexts/PlaygroundContext'
import { ScreenshotFormatOption, MouseButton, MouseScrollDirection } from '@/enums/Playground'
import { useState } from 'react'

export const PlaygroundProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sandboxParametersState, setSandboxParametersState] = useState<SandboxParams>({
    resources: {
      cpu: 2,
      // gpu: 0,
      memory: 4,
      disk: 8,
    },
    createSandboxBaseParams: {
      autoStopInterval: 15,
      autoArchiveInterval: 7,
      autoDeleteInterval: -1,
    },
  })
  const [VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamsState] = useState<VNCInteractionOptionsParams>(
    {
      keyboardHotKeyParams: { keys: '' },
      keyboardPressParams: { key: '' },
      keyboardTypeParams: { text: '' },
      mouseClickParams: {
        x: 100,
        y: 100,
        button: MouseButton.LEFT,
        double: false,
      },
      mouseDragParams: {
        startX: 100,
        startY: 100,
        endX: 200,
        endY: 200,
        button: MouseButton.LEFT,
      },
      mouseMoveParams: {
        x: 100,
        y: 100,
      },
      mouseScrollParams: {
        x: 100,
        y: 100,
        direction: MouseScrollDirection.DOWN,
        amount: 1,
      },
      screenshotOptionsConfig: {
        showCursor: false,
        format: ScreenshotFormatOption.PNG,
        quality: 100,
        scale: 1,
      },
      screenshotRegionConfig: {
        x: 100,
        y: 100,
        width: 300,
        height: 200,
      },
    },
  )

  const setSandboxParameterValue: SetSandboxParamsValue = (key, value) => {
    setSandboxParametersState((prev) => ({ ...prev, [key]: value }))
  }

  const setVNCInteractionOptionsParamValue: SetVNCInteractionOptionsParamValue = (key, value) => {
    setVNCInteractionOptionsParamsState((prev) => ({ ...prev, [key]: value }))
  }

  const [runningActionMethod, setRunningActionMethod] = useState<RunningActionMethodName>(null)
  const [actionRuntimeError, setActionRuntimeError] = useState<ActionRuntimeError>({})

  const validatePlaygroundActionRequiredParams: ValidatePlaygroundActionRequiredParams = (
    actionParamsFormData,
    actionParamsState,
  ) => {
    if (actionParamsFormData.some((formItem) => formItem.required)) {
      const emptyFormItem = actionParamsFormData
        .filter((formItem) => formItem.required)
        .find((formItem) => {
          const value = actionParamsState[formItem.key]
          return value === '' || value === undefined
        })

      if (emptyFormItem) {
        return `${emptyFormItem.label} parameter is required for this action`
      }
    }

    return undefined
  }

  const runPlaygroundAction: RunPlaygroundActionBasic = async (actionFormData, invokeApi) => {
    setRunningActionMethod(actionFormData.methodName)
    try {
      await invokeApi(actionFormData)
    } catch (error) {
      console.log('API call error', error)
    }
    setTimeout(() => setRunningActionMethod(null), 5000)
  }

  const runPlaygroundActionWithParams: RunPlaygroundActionWithParams = async (actionFormData, invokeApi) => {
    const validationError = validatePlaygroundActionRequiredParams(
      actionFormData.parametersFormItems,
      actionFormData.parametersState,
    )
    if (validationError) {
      setActionRuntimeError((prev) => ({
        ...prev,
        [actionFormData.methodName]: validationError,
      }))
      setRunningActionMethod(null)
      return
    }
    // Reset error
    setActionRuntimeError((prev) => ({
      ...prev,
      [actionFormData.methodName]: null,
    }))
    return await runPlaygroundAction(actionFormData, invokeApi)
  }

  return (
    <PlaygroundContext.Provider
      value={{
        sandboxParametersState,
        setSandboxParameterValue,
        VNCInteractionOptionsParamsState,
        setVNCInteractionOptionsParamValue,
        runPlaygroundActionWithParams,
        runPlaygroundActionWithoutParams: runPlaygroundAction,
        runningActionMethod,
        actionRuntimeError,
      }}
    >
      {children}
    </PlaygroundContext.Provider>
  )
}
