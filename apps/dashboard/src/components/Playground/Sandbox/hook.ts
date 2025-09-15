/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useContext } from 'react'
import { PlaygroundSandboxParamsContext } from '@/components/Playground/Sandbox/context'

export function usePlaygroundSandboxParams() {
  const context = useContext(PlaygroundSandboxParamsContext)

  if (!context) {
    throw new Error('usePlaygroundSandboxParams must be used within a <Playground />')
  }

  return context
}
