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
} from '@/contexts/PlaygroundContext'
import { ScreenshotFormatOption, MouseButton, MouseScrollDirection } from '@/enums/Playground'
import { useState } from 'react'

export const PlaygroundProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sandboxParametersState, setSandboxParametersState] = useState<SandboxParams>({
    // Default values
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

  return (
    <PlaygroundContext.Provider
      value={{
        sandboxParametersState,
        setSandboxParameterValue,
        VNCInteractionOptionsParamsState,
        setVNCInteractionOptionsParamValue,
      }}
    >
      {children}
    </PlaygroundContext.Provider>
  )
}
