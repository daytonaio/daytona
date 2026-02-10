/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { usePlaygroundSandbox } from '@/hooks/usePlaygroundSandbox'
import { useQuery } from '@tanstack/react-query'
import { AnimatePresence, motion } from 'framer-motion'
import { RefreshCcw } from 'lucide-react'
import { Window, WindowContent, WindowTitleBar } from '../Window'

const motionLoadingProps = {
  initial: { opacity: 0, y: 10 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -10 },
  transition: { duration: 0.175 },
}

type WebTerminalProps = {
  getPortPreviewUrl: (sandboxId: string, port: number) => Promise<string>
  className?: string
}

const WebTerminal: React.FC<WebTerminalProps> = ({ getPortPreviewUrl, className }) => {
  const { sandbox: terminalSandbox, error: terminalSandboxError } = usePlaygroundSandbox()

  const {
    data: terminalUrl,
    isLoading,
    refetch: refetchTerminalUrl,
  } = useQuery({
    queryKey: ['webTerminalUrl', terminalSandbox?.id],
    queryFn: async () => {
      try {
        if (!terminalSandbox) return null
        return await getPortPreviewUrl(terminalSandbox.id, 22222)
      } catch (error) {
        handleApiError(error, 'Failed to construct web terminal URL')
        return null
      }
    },
    enabled: !!terminalSandbox,
    retry: false,
  })

  // Loading terminal URL conditions:
  // - No sandbox yet, no error → true (waiting for sandbox)
  // - Sandbox arrived, query fetching → true (isLoading)
  // - Sandbox arrived, query done → false
  // - Sandbox errored → false
  const loadingTerminalUrl = isLoading || (!terminalSandbox && !terminalSandboxError)

  return (
    <Window className={className}>
      <WindowTitleBar>Sandbox Terminal</WindowTitleBar>
      <WindowContent>
        <div className="w-full bg-muted/40 dark:bg-muted/10 min-h-[500px] flex flex-col [&>*]:flex-1">
          {loadingTerminalUrl || !terminalUrl ? (
            <div className="h-full flex items-center justify-center rounded-lg">
              <AnimatePresence mode="wait">
                {loadingTerminalUrl ? (
                  <motion.p className="flex items-center gap-2" key="loading" {...motionLoadingProps}>
                    <Spinner className="size-4 mr-2" /> Loading terminal...
                  </motion.p>
                ) : (
                  <motion.p
                    key="error"
                    className="flex flex-col items-center justify-center gap-2"
                    {...motionLoadingProps}
                  >
                    There was an error loading the terminal.
                    {terminalSandbox && (
                      <Button variant="outline" className="ml-2" onClick={() => refetchTerminalUrl()}>
                        <RefreshCcw className="size-4" />
                        Retry
                      </Button>
                    )}
                  </motion.p>
                )}
              </AnimatePresence>
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
