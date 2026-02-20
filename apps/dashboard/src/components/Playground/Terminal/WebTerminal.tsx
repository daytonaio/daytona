/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { usePlaygroundSandbox } from '@/hooks/usePlaygroundSandbox'
import { AnimatePresence, motion } from 'framer-motion'
import { RefreshCcw } from 'lucide-react'
import { Window, WindowContent, WindowTitleBar } from '../Window'

const motionLoadingProps = {
  initial: { opacity: 0, y: 10 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -10 },
  transition: { duration: 0.175 },
}

const WebTerminal: React.FC<{ className?: string }> = ({ className }) => {
  const { sandbox, terminal } = usePlaygroundSandbox()
  const loadingTerminalUrl = terminal.loading || (!sandbox.instance && !sandbox.error)

  return (
    <Window className={className}>
      <WindowTitleBar>Sandbox Terminal</WindowTitleBar>
      <WindowContent>
        <div className="w-full bg-muted/40 dark:bg-muted/10 min-h-[500px] flex flex-col [&>*]:flex-1">
          {loadingTerminalUrl || !terminal.url ? (
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
                    {sandbox.instance ? (
                      <Button variant="outline" className="ml-2" onClick={() => terminal.refetch()}>
                        <RefreshCcw className="size-4" />
                        Retry
                      </Button>
                    ) : (
                      sandbox.error && <span className="text-sm text-muted-foreground">{sandbox.error}</span>
                    )}
                  </motion.p>
                )}
              </AnimatePresence>
            </div>
          ) : (
            <iframe title="Interactive web terminal for sandbox" src={terminal.url} width={'100%'} height={'100%'} />
          )}
        </div>
      </WindowContent>
    </Window>
  )
}

export default WebTerminal
