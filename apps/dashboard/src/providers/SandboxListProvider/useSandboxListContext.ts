/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListContext, SandboxListContextValue } from './SandboxListContext'
import { useContext } from 'react'

export function useSandboxListContext(): SandboxListContextValue {
  const context = useContext(SandboxListContext)
  if (!context) {
    throw new Error('useSandboxListContext must be used within a SandboxListProvider')
  }
  return context
}
