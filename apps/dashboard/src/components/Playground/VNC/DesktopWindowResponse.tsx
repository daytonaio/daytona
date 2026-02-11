/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { usePlayground } from '@/hooks/usePlayground'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronUpIcon, RefreshCcw, XIcon } from 'lucide-react'
import { useEffect } from 'react'
import { Group, Panel, usePanelRef } from 'react-resizable-panels'
import ResponseCard from '../ResponseCard'
import { Window, WindowContent, WindowTitleBar } from '../Window'

const motionLoadingProps = {
  initial: { opacity: 0, y: 10 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -10 },
  transition: { duration: 0.175 },
}

const VNCDesktopWindowResponse: React.FC<{ className?: string }> = ({ className }) => {
  const { VNCInteractionOptionsParamsState } = usePlayground()
  const VNCUrl = VNCInteractionOptionsParamsState.VNCUrl
  const { sandbox, sandboxError, vncUrlLoading, vncUrlError, refetchVNCUrl } =
    VNCInteractionOptionsParamsState.VNCSandboxData ?? {}

  const loadingVNCUrl = vncUrlLoading || (!sandbox && !sandboxError)

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
              {loadingVNCUrl || vncUrlError || !VNCUrl ? (
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
                        {vncUrlError || 'There was an error loading VNC.'}
                        {sandbox ? (
                          <Button variant="outline" className="ml-2" onClick={() => refetchVNCUrl?.()}>
                            <RefreshCcw className="size-4" />
                            Retry
                          </Button>
                        ) : (
                          sandboxError && <span className="text-sm text-muted-foreground">{sandboxError}</span>
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
