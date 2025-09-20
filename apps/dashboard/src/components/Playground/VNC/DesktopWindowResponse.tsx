/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import ResponseCard from '../ResponseCard'
import { toast } from 'sonner'
import { handleApiError } from '@/lib/error-handling'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { usePlayground } from '@/hooks/usePlayground'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useApi } from '@/hooks/useApi'
import { useState, useEffect, useCallback, ReactNode } from 'react'

type VNCDesktopWindowResponseProps = {
  sandboxId: string
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
}

const VNCDesktopWindowResponse: React.FC<VNCDesktopWindowResponseProps> = ({ sandboxId, getPortPreviewUrl }) => {
  const [loadingVNCUrl, setLoadingVNCUrl] = useState(true)
  const [VNCLoadingError, setVNCLoadingError] = useState<string | ReactNode>('')
  const [VNCUrl, setVNCUrl] = useState<string | null>(null)

  const { selectedOrganization } = useSelectedOrganization()
  const { toolboxApi } = useApi()
  const { VNCInteractionOptionsParamsState } = usePlayground()

  const getVNCUrl = useCallback(async (): Promise<string | null> => {
    try {
      const url = await getPortPreviewUrl(sandboxId, 6080)
      return url + '/vnc.html'
    } catch (error) {
      handleApiError(error, 'Failed to construct VNC URL')
      return null
    }
  }, [sandboxId, getPortPreviewUrl])

  const getVNCComputerUseUrl = useCallback(async () => {
    setLoadingVNCUrl(true)
    // Notify user immediately that we're checking VNC status
    toast.info('Checking VNC desktop status...')
    try {
      // First, check if computer use is already started
      const statusResponse = await toolboxApi.getComputerUseStatus(sandboxId, selectedOrganization?.id)
      const status = statusResponse.data.status

      // Check if computer use is active (all processes running)
      if (status === 'active') {
        const vncUrl = await getVNCUrl()
        if (vncUrl) setVNCUrl(vncUrl)
      } else {
        // Computer use is not active, try to start it
        try {
          await toolboxApi.startComputerUse(sandboxId, selectedOrganization?.id)
          toast.success('Starting VNC desktop...')

          // Wait a moment for processes to start, then open VNC
          await new Promise((resolve) => setTimeout(resolve, 5000))

          try {
            const newStatusResponse = await toolboxApi.getComputerUseStatus(sandboxId, selectedOrganization?.id)
            const newStatus = newStatusResponse.data.status

            if (newStatus === 'active') {
              const vncUrl = await getVNCUrl()

              if (vncUrl) setVNCUrl(vncUrl)
            } else {
              toast.error(`VNC desktop failed to start. Status: ${newStatus}`)
              setVNCLoadingError(`VNC desktop failed to start. Status: ${newStatus}`)
            }
          } catch (error) {
            handleApiError(error, 'Failed to check VNC status after start')
          }
        } catch (startError: any) {
          // Check if this is a computer-use availability error
          const errorMessage = startError?.response?.data?.message || startError?.message || String(startError)

          if (errorMessage === 'Computer-use functionality is not available') {
            toast.error('Computer-use functionality is not available', {
              description: (
                <div>
                  <div>Computer-use dependencies are missing in the runtime environment.</div>
                  <div className="mt-2">
                    <a
                      href={`${DAYTONA_DOCS_URL}/getting-started/computer-use`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      See documentation on how to configure the runtime for computer-use
                    </a>
                  </div>
                </div>
              ),
            })
            setVNCLoadingError(
              <div>
                <div>Computer-use dependencies are missing in the runtime environment.</div>
                <div className="mt-2">
                  <a
                    href={`${DAYTONA_DOCS_URL}/getting-started/computer-use`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    See documentation on how to configure the runtime for computer-use
                  </a>
                </div>
              </div>,
            )
          } else {
            handleApiError(startError, 'Failed to start VNC desktop')
          }
        }
      }
    } catch (error) {
      handleApiError(error, 'Failed to check VNC status')
    } finally {
      setLoadingVNCUrl(false)
    }
  }, [getVNCUrl, sandboxId, selectedOrganization, toolboxApi])

  useEffect(() => {
    getVNCComputerUseUrl()
  }, [getVNCComputerUseUrl])

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Desktop window</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="w-full aspect-[4/3] md:aspect-[16/9]">
            {loadingVNCUrl || VNCLoadingError || !VNCUrl ? (
              <div className="h-full flex items-center justify-center rounded-lg">
                <p>{loadingVNCUrl ? 'Loading VNC...' : VNCLoadingError || 'Unable to open VNC. Please try again.'}</p>
              </div>
            ) : (
              <iframe title="VNC desktop window" src={VNCUrl} className="w-full h-full" />
            )}
          </div>
        </CardContent>
      </Card>
      <ResponseCard responseText={VNCInteractionOptionsParamsState.responseText || ''} />
    </>
  )
}

export default VNCDesktopWindowResponse
