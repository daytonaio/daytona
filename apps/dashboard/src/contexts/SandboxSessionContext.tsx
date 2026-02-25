/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'

export interface ISandboxSessionContext {
  isTerminalActivated: (sandboxId: string) => boolean
  activateTerminal: (sandboxId: string) => void
  isVncActivated: (sandboxId: string) => boolean
  activateVnc: (sandboxId: string) => void
}

export const SandboxSessionContext = createContext<ISandboxSessionContext | null>(null)
