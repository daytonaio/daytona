/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeLanguage, Resources, CreateSandboxFromImageParams } from '@daytonaio/sdk-typescript/src'
import { createContext } from 'react'

export interface PlaygroundSandboxParams {
  apiKey?: string
  language?: CodeLanguage
  resources?: Resources
  createFromImageParams?: CreateSandboxFromImageParams
  autoStopInterval?: number
  autoArchiveInterval?: number
  autoDeleteInterval?: number
  languageCodeToRun?: string
  shellCodeToRun?: string
}

export type SetPlaygroundSandboxParamsValue = <K extends keyof PlaygroundSandboxParams>(
  key: K,
  value: PlaygroundSandboxParams[K],
) => void

export interface IPlaygroundSandboxParamsContext {
  playgroundSandboxParametersState: PlaygroundSandboxParams
  setPlaygroundSandboxParameterValue: SetPlaygroundSandboxParamsValue
}

export const PlaygroundSandboxParamsContext = createContext<IPlaygroundSandboxParamsContext | null>(null)
