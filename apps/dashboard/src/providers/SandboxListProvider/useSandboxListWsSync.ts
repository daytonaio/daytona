/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { Sandbox, SandboxDesiredState, SandboxState } from '@daytonaio/api-client'
import { useEffect } from 'react'

interface SandboxStateUpdatedEvent {
  sandbox: Sandbox
  oldState: SandboxState
  newState: SandboxState
}

interface SandboxDesiredStateUpdatedEvent {
  sandbox: Sandbox
  oldDesiredState: SandboxDesiredState
  newDesiredState: SandboxDesiredState
}

interface UseSandboxListWsSyncOptions {
  onSandboxCreated: (sandbox: Sandbox) => void
  onSandboxStateUpdated: (data: SandboxStateUpdatedEvent) => void
  onSandboxDesiredStateUpdated: (data: SandboxDesiredStateUpdatedEvent) => void
}

export function useSandboxListWsSync({
  onSandboxCreated,
  onSandboxStateUpdated,
  onSandboxDesiredStateUpdated,
}: UseSandboxListWsSyncOptions) {
  const { notificationSocket } = useNotificationSocket()

  useEffect(() => {
    if (!notificationSocket) {
      return
    }

    notificationSocket.on('sandbox.created', onSandboxCreated)
    notificationSocket.on('sandbox.state.updated', onSandboxStateUpdated)
    notificationSocket.on('sandbox.desired-state.updated', onSandboxDesiredStateUpdated)

    return () => {
      notificationSocket.off('sandbox.created', onSandboxCreated)
      notificationSocket.off('sandbox.state.updated', onSandboxStateUpdated)
      notificationSocket.off('sandbox.desired-state.updated', onSandboxDesiredStateUpdated)
    }
  }, [notificationSocket, onSandboxCreated, onSandboxDesiredStateUpdated, onSandboxStateUpdated])
}
