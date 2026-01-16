/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useTemporarySandbox } from '@/hooks/useTemporarySandbox'
import { handleApiError } from '@/lib/error-handling'
import { Sandbox } from '@daytonaio/sdk'
import { useCallback, useEffect, useState } from 'react'
import { Window, WindowContent, WindowTitleBar } from '../Window'

type WebTerminalProps = {
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
  className?: string
}

const WebTerminal: React.FC<WebTerminalProps> = ({ getPortPreviewUrl, className }) => {
  const [loadingTerminalUrl, setLoadingTerminalUrl] = useState(true)
  const [terminalUrl, setTerminalUrl] = useState<string | null>(null)

  const { sandbox: terminalSandbox, error: terminalSandboxError } = useTemporarySandbox()

  const getWebTerminalUrl = useCallback(
    async (sandbox: Sandbox) => {
      try {
        const url = await getPortPreviewUrl(sandbox.id, 22222)
        setTerminalUrl(url)
      } catch (error) {
        handleApiError(error, 'Failed to construct web terminal URL')
        setTerminalUrl(null)
      }
    },
    [getPortPreviewUrl],
  )

  useEffect(() => {
    if (terminalSandbox) {
      // Temporary sandbox created -> setup terminal
      const setupWebTerminal = async () => {
        setLoadingTerminalUrl(true)
        await getWebTerminalUrl(terminalSandbox)
        setLoadingTerminalUrl(false)
      }
      setupWebTerminal()
    } else if (terminalSandboxError) setLoadingTerminalUrl(false)
  }, [terminalSandbox, terminalSandboxError, getWebTerminalUrl])

  return (
    <Window className={className}>
      <WindowTitleBar>Sandbox Terminal</WindowTitleBar>
      <WindowContent>
        <div className="w-full bg-muted/40 dark:bg-muted/10 min-h-[500px] flex flex-col [&>*]:flex-1">
          {loadingTerminalUrl || !terminalUrl ? (
            <div className="h-full flex items-center justify-center rounded-lg">
              <p>{loadingTerminalUrl ? 'Loading terminal...' : 'Unable to open the terminal. Please try again.'}</p>
            </div>
          ) : (
            <iframe title="Interactive web terminal for sandbox" src={terminalUrl} width={'100%'} height={'100%'} />
          )}
        </div>
      </WindowContent>
    </Window>
  )
}

export default WebTerminal
