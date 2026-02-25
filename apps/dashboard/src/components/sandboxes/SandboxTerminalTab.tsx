/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { useTerminalSessionQuery } from '@/hooks/queries/useTerminalSessionQuery'
import { useSandboxSessionContext } from '@/hooks/useSandboxSessionContext'
import { isStoppable } from '@/lib/utils/sandbox'
import { Sandbox } from '@daytonaio/api-client'
import { Spinner } from '@/components/ui/spinner'
import { Play, RefreshCw, TerminalSquare } from 'lucide-react'

export function SandboxTerminalTab({ sandbox }: { sandbox: Sandbox }) {
  const running = isStoppable(sandbox)
  const { isTerminalActivated, activateTerminal } = useSandboxSessionContext()

  const [activated, setActivated] = useState(() => isTerminalActivated(sandbox.id))

  const {
    data: session,
    isLoading,
    isError,
    isFetching,
    existingSession,
    reset,
  } = useTerminalSessionQuery(sandbox.id, running && activated)

  // Auto-reconnect: if activated and session is expired, refetch
  useEffect(() => {
    if (!activated || !existingSession) return
    if (existingSession.expiresAt <= Date.now()) {
      reset()
    }
  }, [activated, existingSession, reset])

  const handleConnect = () => {
    activateTerminal(sandbox.id)
    setActivated(true)
  }

  if (!running) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <TerminalSquare className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Sandbox is not running</EmptyTitle>
              <EmptyDescription>
                Start the sandbox to access the terminal.{' '}
                <a href={`${DAYTONA_DOCS_URL}/en/web-terminal`} target="_blank" rel="noopener noreferrer">
                  Learn more
                </a>
                .
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      </div>
    )
  }

  // Not yet activated — show connect button
  if (!activated) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <TerminalSquare className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Terminal</EmptyTitle>
              <EmptyDescription>
                Connect to an interactive terminal session in your sandbox.{' '}
                <a href={`${DAYTONA_DOCS_URL}/en/web-terminal`} target="_blank" rel="noopener noreferrer">
                  Learn more
                </a>
                .
              </EmptyDescription>
            </EmptyHeader>
            <Button onClick={handleConnect}>
              <Play className="size-4" />
              Connect
            </Button>
          </Empty>
        </div>
      </div>
    )
  }

  // Loading / fetching
  if (isLoading || isFetching) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex items-center justify-center gap-2 text-muted-foreground">
          <Spinner className="size-4" />
          <span className="text-sm">Connecting...</span>
        </div>
      </div>
    )
  }

  // Error
  if (isError || !session) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyTitle>Failed to connect</EmptyTitle>
              <EmptyDescription>Something went wrong while connecting to the terminal.</EmptyDescription>
            </EmptyHeader>
            <Button variant="outline" size="sm" onClick={() => reset()}>
              <RefreshCw className="size-4" />
              Retry
            </Button>
          </Empty>
        </div>
      </div>
    )
  }

  // Active session
  return (
    <div className="flex-1 flex flex-col p-4">
      <div className="flex-1 min-h-0 rounded-md border border-border bg-black overflow-hidden p-1">
        <iframe title="Sandbox terminal" src={session.url} className="w-full h-full border-0" />
      </div>
    </div>
  )
}
