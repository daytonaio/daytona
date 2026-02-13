/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useDeepCompareMemo } from '@/hooks/useDeepCompareMemo'
import { usePlayground } from '@/hooks/usePlayground'
import { useSandboxSession, UseSandboxSessionResult } from '@/hooks/useSandboxSession'
import { createContext, useEffect, useRef } from 'react'

export const PlaygroundSandboxContext = createContext<UseSandboxSessionResult | null>(null)

export const PlaygroundSandboxProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { getSandboxParametersInfo } = usePlayground()
  const { createSandboxParams } = getSandboxParametersInfo()
  const stableCreateParams = useDeepCompareMemo(createSandboxParams)

  const session = useSandboxSession({
    key: 'playground',
    createParams: stableCreateParams,
    autoCreate: true,
    terminal: true,
    vnc: true,
  })

  const vncSandboxId = useRef<string | null>(null)
  useEffect(() => {
    const id = session.sandbox.instance?.id
    if (id && vncSandboxId.current !== id) {
      vncSandboxId.current = id
      session.vnc.start()
    }
  }, [session.sandbox.instance?.id, session.vnc])

  return <PlaygroundSandboxContext.Provider value={session}>{children}</PlaygroundSandboxContext.Provider>
}
