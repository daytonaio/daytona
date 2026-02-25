/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ISandboxSessionContext, SandboxSessionContext } from '@/contexts/SandboxSessionContext'
import { ReactNode, useCallback, useRef } from 'react'

type SessionFlags = { terminal: boolean; vnc: boolean }

export function SandboxSessionProvider({ children }: { children: ReactNode }) {
  const flagsRef = useRef<Map<string, SessionFlags>>(new Map())

  const getFlags = useCallback((sandboxId: string): SessionFlags => {
    let flags = flagsRef.current.get(sandboxId)
    if (!flags) {
      flags = { terminal: false, vnc: false }
      flagsRef.current.set(sandboxId, flags)
    }
    return flags
  }, [])

  const value: ISandboxSessionContext = {
    isTerminalActivated: (sandboxId) => getFlags(sandboxId).terminal,
    activateTerminal: (sandboxId) => {
      getFlags(sandboxId).terminal = true
    },
    isVncActivated: (sandboxId) => getFlags(sandboxId).vnc,
    activateVnc: (sandboxId) => {
      getFlags(sandboxId).vnc = true
    },
  }

  return <SandboxSessionContext.Provider value={value}>{children}</SandboxSessionContext.Provider>
}
