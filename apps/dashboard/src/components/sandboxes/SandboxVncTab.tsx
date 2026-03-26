/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { useStartVncMutation } from '@/hooks/mutations/useStartVncMutation'
import { useVncInitialStatusQuery, useVncPollStatusQuery } from '@/hooks/queries/useVncStatusQuery'
import { useVncSessionQuery } from '@/hooks/queries/useVncSessionQuery'
import { isStoppable } from '@/lib/utils/sandbox'
import { Sandbox } from '@daytonaio/api-client'
import { Spinner } from '@/components/ui/spinner'
import { Monitor, Play, RefreshCw } from 'lucide-react'

const VNC_MISSING_DEPS_MSG = 'Computer-use functionality is not available'

export function SandboxVncTab({ sandbox }: { sandbox: Sandbox }) {
  const running = isStoppable(sandbox)

  // 1. Check initial VNC availability & status
  const initialStatusQuery = useVncInitialStatusQuery(sandbox.id, running)

  const isMissingDeps = (initialStatusQuery.error as Error | null)?.message === VNC_MISSING_DEPS_MSG
  const alreadyActive = initialStatusQuery.data === 'active'

  // 2. Start VNC
  const startMutation = useStartVncMutation(sandbox.id)

  const startError = startMutation.error?.message
  const startMissingDeps = startError === VNC_MISSING_DEPS_MSG

  // 3. Poll until active after starting
  const pollStatusQuery = useVncPollStatusQuery(sandbox.id, startMutation.isSuccess)

  const vncReady = alreadyActive || pollStatusQuery.data === 'active'

  // 4. Get signed URL once ready
  const {
    data: session,
    isLoading: sessionLoading,
    isError: sessionError,
    existingSession,
    reset,
  } = useVncSessionQuery(sandbox.id, vncReady)

  // Auto-reconnect: if session is expired, refetch
  useEffect(() => {
    if (!existingSession) return
    if (existingSession.expiresAt <= Date.now()) {
      reset()
    }
  }, [existingSession, reset])

  const isStarting =
    startMutation.isPending ||
    (startMutation.isSuccess && !pollStatusQuery.data && !pollStatusQuery.error) ||
    (vncReady && sessionLoading)

  const unavailable = isMissingDeps || startMissingDeps
  const pollError = pollStatusQuery.error?.message
  const anyError = startError && !startMissingDeps ? startError : pollError

  if (!running) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Monitor className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Sandbox is not running</EmptyTitle>
              <EmptyDescription>
                Start the sandbox to access the VNC desktop.{' '}
                <a href={`${DAYTONA_DOCS_URL}/en/vnc-access`} target="_blank" rel="noopener noreferrer">
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

  if (initialStatusQuery.isLoading) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex items-center justify-center gap-2 text-muted-foreground">
          <Spinner className="size-4" />
          <span className="text-sm">Checking VNC status...</span>
        </div>
      </div>
    )
  }

  if (unavailable) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Monitor className="size-4" />
              </EmptyMedia>
              <EmptyTitle>VNC not available</EmptyTitle>
              <EmptyDescription>
                Computer-use dependencies are not installed in this sandbox.{' '}
                <a href={`${DAYTONA_DOCS_URL}/en/vnc-access`} target="_blank" rel="noopener noreferrer">
                  Read the setup guide
                </a>
                .
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      </div>
    )
  }

  // Not yet started — show start button
  if (!vncReady && !isStarting) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Monitor className="size-4" />
              </EmptyMedia>
              <EmptyTitle>VNC Desktop</EmptyTitle>
              <EmptyDescription>
                Start the VNC server to access a graphical desktop.{' '}
                <a href={`${DAYTONA_DOCS_URL}/en/vnc-access`} target="_blank" rel="noopener noreferrer">
                  Learn more
                </a>
                .
              </EmptyDescription>
            </EmptyHeader>
            <Button onClick={() => startMutation.mutate()}>
              <Play className="size-4" />
              Start VNC
            </Button>
          </Empty>
        </div>
      </div>
    )
  }

  // Starting / polling / getting URL
  if (isStarting) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border bg-neutral-950 flex items-center justify-center gap-2 text-muted-foreground">
          <Spinner className="size-4" />
          <span className="text-sm">
            {startMutation.isPending
              ? 'Starting VNC desktop...'
              : vncReady && sessionLoading
                ? 'Getting preview URL...'
                : 'Waiting for VNC to become ready...'}
          </span>
        </div>
      </div>
    )
  }

  // Error
  if (anyError || sessionError) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border flex">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyTitle>Failed to connect</EmptyTitle>
              <EmptyDescription>{anyError || 'Something went wrong while connecting to VNC.'}</EmptyDescription>
            </EmptyHeader>
            <Button variant="outline" size="sm" onClick={reset}>
              <RefreshCw className="size-4" />
              Retry
            </Button>
          </Empty>
        </div>
      </div>
    )
  }

  // Active session
  if (session) {
    return (
      <div className="flex-1 flex flex-col p-4">
        <div className="flex-1 min-h-0 rounded-md border border-border bg-neutral-950 overflow-hidden">
          <iframe
            title="VNC desktop"
            src={`${session.url}/vnc.html?autoconnect=true&resize=scale`}
            className="w-full h-full border-0"
          />
        </div>
      </div>
    )
  }

  return null
}
