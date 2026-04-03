/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SANDBOX_TARGET_DEFAULT_VALUE } from '@/constants/Playground'
import { PlaygroundCategories } from '@/enums/Playground'
import { useDeepCompareMemo } from '@/hooks/useDeepCompareMemo'
import { usePlayground } from '@/hooks/usePlayground'
import { useSandboxSession, UseSandboxSessionResult } from '@/hooks/useSandboxSession'
import { createContext, useEffect, useMemo, useRef } from 'react'

export const PlaygroundSandboxContext = createContext<UseSandboxSessionResult | null>(null)

export const PlaygroundSandboxProvider: React.FC<{
  activeTab: PlaygroundCategories
  children: React.ReactNode
}> = ({ activeTab, children }) => {
  const { sandboxParametersState, getSandboxParametersInfo } = usePlayground()
  const { createSandboxParams } = getSandboxParametersInfo()
  const stableCreateParams = useDeepCompareMemo(createSandboxParams)

  const target = useMemo(() => {
    const t = sandboxParametersState.sandboxTarget
    if (!t || t === SANDBOX_TARGET_DEFAULT_VALUE) return undefined
    return t
  }, [sandboxParametersState.sandboxTarget])

  const session = useSandboxSession({
    scope: 'playground',
    createParams: stableCreateParams,
    target,
    terminal: true,
    vnc: true,
    notify: { vnc: activeTab === PlaygroundCategories.VNC },
  })

  const createRef = useRef(session.sandbox.create)
  createRef.current = session.sandbox.create

  useEffect(() => {
    const needsSandbox = activeTab === PlaygroundCategories.TERMINAL || activeTab === PlaygroundCategories.VNC
    if (needsSandbox && !session.sandbox.instance && !session.sandbox.loading && !session.sandbox.error) {
      createRef.current()
    }
  }, [activeTab, session.sandbox.instance, session.sandbox.loading, session.sandbox.error])

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
