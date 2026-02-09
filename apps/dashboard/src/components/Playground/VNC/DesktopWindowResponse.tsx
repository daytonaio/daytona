/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { useApi } from '@/hooks/useApi'
import { usePlayground } from '@/hooks/usePlayground'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { Sandbox } from '@daytonaio/sdk'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronUpIcon, RefreshCcw, XIcon } from 'lucide-react'
import { ReactNode, useCallback, useEffect, useState } from 'react'
import { Group, Panel, usePanelRef } from 'react-resizable-panels'
import { toast } from 'sonner'
import ResponseCard from '../ResponseCard'
import { Window, WindowContent, WindowTitleBar } from '../Window'

type VNCDesktopWindowResponseProps = {
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
  className?: string
}

const motionLoadingProps = {
  initial: { opacity: 0, y: 10 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -10 },
  transition: { duration: 0.175 },
}

function isComputerUseUnavailableError(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error)
  return message === 'Computer-use functionality is not available'
}

const computerUseMissingErrorMessage = (
  <div>
    <div>Computer-use dependencies are missing in the runtime environment.</div>
    <div className="mt-2">
      <a
        href={`${DAYTONA_DOCS_URL}/en/vnc-access/`}
        target="_blank"
        rel="noopener noreferrer"
        className="text-primary hover:underline"
      >
        See documentation on how to configure the runtime for computer-use
      </a>
    </div>
  </div>
)

const VNCDesktopWindowResponse: React.FC<VNCDesktopWindowResponseProps> = ({ getPortPreviewUrl, className }) => {
  const [loadingVNCUrl, setLoadingVNCUrl] = useState(true)
  const [VNCLoadingError, setVNCLoadingError] = useState<string | ReactNode>('')

  const { selectedOrganization } = useSelectedOrganization()
  const { toolboxApi } = useApi()
  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue } = usePlayground()
  const VNCSandboxData = VNCInteractionOptionsParamsState.VNCSandboxData
  const VNCUrl = VNCInteractionOptionsParamsState.VNCUrl

  const getVNCUrl = useCallback(
    async (sandbox: Sandbox): Promise<string | null> => {
      try {
        const url = await getPortPreviewUrl(sandbox.id, 6080)
        return url + '/vnc.html'
      } catch (error) {
        handleApiError(error, 'Failed to construct VNC URL')
        return null
      }
    },
    [getPortPreviewUrl],
  )

  const getVNCComputerUseUrl = useCallback(
    async (sandbox: Sandbox) => {
      // Notify user immediately that we're checking VNC status
      toast.info('Checking VNC desktop status...')
      try {
        // First, check if computer use is already started
        const statusResponse = await toolboxApi.getComputerUseStatusDeprecated(sandbox.id, selectedOrganization?.id)
        const status = statusResponse.data.status

        // Check if computer use is active (all processes running)
        if (status === 'active') {
          const vncUrl = await getVNCUrl(sandbox)
          if (vncUrl) setVNCInteractionOptionsParamValue('VNCUrl', vncUrl)
        } else {
          // Computer use is not active, try to start it
          try {
            await toolboxApi.startComputerUseDeprecated(sandbox.id, selectedOrganization?.id)
            toast.success('Starting VNC desktop...')

            // Wait a moment for processes to start, then open VNC
            await new Promise((resolve) => setTimeout(resolve, 5000))

            try {
              const newStatusResponse = await toolboxApi.getComputerUseStatusDeprecated(
                sandbox.id,
                selectedOrganization?.id,
              )
              const newStatus = newStatusResponse.data.status

              if (newStatus === 'active') {
                const vncUrl = await getVNCUrl(sandbox)
                if (vncUrl) setVNCInteractionOptionsParamValue('VNCUrl', vncUrl)
              } else {
                toast.error(`VNC desktop failed to start. Status: ${newStatus}`)
                setVNCLoadingError(`VNC desktop failed to start. Status: ${newStatus}`)
              }
            } catch (error) {
              handleApiError(error, 'Failed to check VNC status after start')
            }
          } catch (startError: any) {
            handleApiError(startError, 'Failed to start VNC desktop')
          }
        }
      } catch (error) {
        const isComputerUseError = isComputerUseUnavailableError(error)
        if (isComputerUseError) {
          toast.error('Computer-use functionality is not available', {
            description: computerUseMissingErrorMessage,
          })
          setVNCLoadingError(computerUseMissingErrorMessage)
          return
        }
        handleApiError(error, 'Failed to check VNC status')
      }
    },
    [getVNCUrl, selectedOrganization, toolboxApi, setVNCInteractionOptionsParamValue],
  )

  const setupVNCComputerUse = useCallback(
    async (sandbox: Sandbox) => {
      setLoadingVNCUrl(true)
      await getVNCComputerUseUrl(sandbox) // if (VNCSandboxData.sandbox) guarantes that value isn't null so we put as Sandbox to silence TS compiler
      setLoadingVNCUrl(false)
    },
    [getVNCComputerUseUrl],
  )

  useEffect(() => {
    setVNCInteractionOptionsParamValue('VNCUrl', null) // Reset VNCurl value
    if (!VNCSandboxData) return
    if (VNCSandboxData.sandbox) {
      // Sandbox created -> setup VNC
      setupVNCComputerUse(VNCSandboxData.sandbox)
    } else if (VNCSandboxData.error) setLoadingVNCUrl(false)
  }, [setVNCInteractionOptionsParamValue, VNCSandboxData, getVNCComputerUseUrl, setupVNCComputerUse])

  const resultPanelRef = usePanelRef()

  useEffect(() => {
    if (resultPanelRef.current?.isCollapsed()) {
      resultPanelRef.current?.resize('20%')
    }
  }, [VNCInteractionOptionsParamsState.responseContent, resultPanelRef])

  return (
    <Window className={className}>
      <WindowTitleBar>Desktop Window </WindowTitleBar>
      <WindowContent className="w-full flex flex-col items-center justify-center">
        <Group orientation="vertical" className="aspect-[4/3] md:aspect-[16/9] border-border rounded-b-md">
          <Panel minSize={'20%'} className="overflow-auto">
            <div className="aspect-[4/3] md:aspect-[16/9] bg-muted/40 dark:bg-muted/10 rounded-lg">
              {loadingVNCUrl || VNCLoadingError || !VNCUrl ? (
                <div className="h-full flex items-center justify-center rounded-lg">
                  <AnimatePresence mode="wait">
                    {loadingVNCUrl ? (
                      <motion.p className="flex items-center gap-2" key="loading" {...motionLoadingProps}>
                        <Spinner className="size-4 mr-2" /> Loading VNC...
                      </motion.p>
                    ) : (
                      <motion.p
                        key="error"
                        className="flex flex-col items-center justify-center gap-2"
                        {...motionLoadingProps}
                      >
                        {VNCLoadingError || 'There was an error loading VNC.'}
                        {VNCSandboxData?.sandbox && (
                          <Button
                            variant="outline"
                            className="ml-2"
                            onClick={() => {
                              setVNCLoadingError('')
                              setupVNCComputerUse(VNCSandboxData.sandbox!)
                            }}
                          >
                            <RefreshCcw className="size-4" />
                            Retry
                          </Button>
                        )}
                      </motion.p>
                    )}
                  </AnimatePresence>
                </div>
              ) : (
                <iframe title="VNC desktop window" src={`${VNCUrl}?resize=scale`} className="w-full h-full" />
              )}
            </div>
          </Panel>

          <Panel maxSize="80%" minSize="20%" panelRef={resultPanelRef} collapsedSize={0} collapsible defaultSize={0}>
            <div className="bg-background w-full border rounded-md overflow-auto flex flex-col h-full">
              <div className="flex justify-between border-b px-4 pr-2 py-1 text-xs items-center dark:bg-muted/50">
                <div className="text-muted-foreground font-mono">Result</div>
                <div className="flex items-center gap-2">
                  <TooltipButton
                    onClick={() => resultPanelRef.current?.resize('80%')}
                    tooltipText="Maximize"
                    className="h-6 w-6"
                    size="sm"
                    variant="ghost"
                  >
                    <ChevronUpIcon className="w-4 h-4" />
                  </TooltipButton>
                  <TooltipButton
                    tooltipText="Close"
                    className="h-6 w-6"
                    size="sm"
                    variant="ghost"
                    onClick={() => resultPanelRef.current?.collapse()}
                  >
                    <XIcon />
                  </TooltipButton>
                </div>
              </div>
              <div className="flex-1 overflow-y-auto">
                <ResponseCard
                  responseContent={
                    VNCInteractionOptionsParamsState.responseContent || (
                      <div className="text-muted-foreground font-mono">Interaction results will be shown here...</div>
                    )
                  }
                />
              </div>
            </div>
          </Panel>
        </Group>
      </WindowContent>
    </Window>
  )
}

export default VNCDesktopWindowResponse
