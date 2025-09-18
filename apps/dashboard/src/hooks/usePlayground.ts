/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useContext } from 'react'
import { PlaygroundContext } from '@/contexts/PlaygroundContext'

export function usePlayground() {
  const context = useContext(PlaygroundContext)

  if (!context) {
    throw new Error('usePlayground must be used within a <Playground />')
  }

  return context
}
