/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Sandbox } from '@daytonaio/sdk'
import { handleApiError } from '@/lib/error-handling'
import { useTemporarySandbox } from '@/hooks/useTemporarySandbox'
import { useState, useEffect, useCallback } from 'react'

type WebTerminalProps = {
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
}

const WebTerminal: React.FC<WebTerminalProps> = ({ getPortPreviewUrl }) => {
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
    <div className="h-full flex flex-col justify-center">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Web Terminal</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="w-full h-[400px]">
            {loadingTerminalUrl || !terminalUrl ? (
              <div className="h-full flex items-center justify-center rounded-lg">
                <p>{loadingTerminalUrl ? 'Loading terminal...' : 'Unable to open the terminal. Please try again.'}</p>
              </div>
            ) : (
              <iframe title="Interactive web terminal for sandbox" src={terminalUrl} width={'100%'} height={'100%'} />
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default WebTerminal
