/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { handleApiError } from '@/lib/error-handling'
import { useState, useEffect, useCallback } from 'react'

type WebTerminalProps = {
  sandboxId: string
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
}

const WebTerminal: React.FC<WebTerminalProps> = ({ sandboxId, getPortPreviewUrl }) => {
  const [loadingTerminalUrl, setLoadingTerminalUrl] = useState(true)
  const [terminalUrl, setTerminalUrl] = useState<string | null>(null)

  const getWebTerminalUrl = useCallback(async () => {
    try {
      setLoadingTerminalUrl(true)
      const url = await getPortPreviewUrl(sandboxId, 22222)
      setTerminalUrl(url)
    } catch (error) {
      handleApiError(error, 'Failed to construct web terminal URL')
      setTerminalUrl(null)
    } finally {
      setLoadingTerminalUrl(false)
    }
  }, [sandboxId, getPortPreviewUrl])

  useEffect(() => {
    getWebTerminalUrl()
  }, [getWebTerminalUrl])

  return (
    <div className="h-full flex flex-col justify-center">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Web Terminal</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="w-full h-[400px]">
            {loadingTerminalUrl || !terminalUrl ? (
              <div className="h-full bg-black text-white flex items-center justify-center rounded-lg">
                <p>{loadingTerminalUrl ? 'Loading terminal...' : 'Unable to open the terminal. Please try again.'}</p>
              </div>
            ) : (
              <iframe
                title="Interactive web terminal for sandbox"
                src={terminalUrl}
                width={'100%'}
                height={'100%'}
                style={{ backgroundColor: '#000' }}
              />
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default WebTerminal
