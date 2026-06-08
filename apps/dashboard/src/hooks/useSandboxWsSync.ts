/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { queryKeys } from '@/hooks/queries/queryKeys'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Sandbox, SandboxDesiredState, SandboxState } from '@daytona/api-client'
import type { QueryKey } from '@tanstack/react-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

export type SandboxWsSyncEvent =
  | {
      type: 'created'
      sandbox: Sandbox
    }
  | {
      type: 'state.updated'
      sandbox: Sandbox
      oldState: SandboxState
      newState: SandboxState
    }
  | {
      type: 'desired-state.updated'
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }

interface UseSandboxWsSyncOptions<TData = Sandbox> {
  enabled?: boolean
  sandboxId?: string
  queryKey?: QueryKey
  sync?: (oldData: TData | undefined, sandbox: Sandbox, event: SandboxWsSyncEvent) => TData | undefined
  onSync?: (event: SandboxWsSyncEvent) => void | Promise<void>
}

export function useSandboxWsSync<TData = Sandbox>({
  enabled = true,
  sandboxId,
  queryKey,
  sync,
  onSync,
}: UseSandboxWsSyncOptions<TData>) {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()
  const queryKeyRef = useRef(queryKey)
  const syncRef = useRef(sync)
  const onSyncRef = useRef(onSync)

  queryKeyRef.current = queryKey
  syncRef.current = sync
  onSyncRef.current = onSync

  useEffect(() => {
    if (!enabled || !notificationSocket || !selectedOrganization?.id) return

    const cancelSandboxQuery = async () => {
      if (!queryKeyRef.current) return

      await queryClient.cancelQueries({
        queryKey: queryKeyRef.current,
      })
    }

    const syncSandboxInCache = (event: SandboxWsSyncEvent) => {
      if (!queryKeyRef.current || !syncRef.current) return

      queryClient.setQueryData<TData>(queryKeyRef.current, (oldData) =>
        syncRef.current?.(oldData, event.sandbox, event),
      )
    }

    const syncSandboxFromEvent = async (event: SandboxWsSyncEvent) => {
      await cancelSandboxQuery()
      syncSandboxInCache(event)
      await onSyncRef.current?.(event)
    }

    const handleCreated = async (sandbox: Sandbox) => {
      if (sandboxId && sandbox.id !== sandboxId) return
      await syncSandboxFromEvent({ type: 'created', sandbox })
    }

    const handleStateUpdated = async (data: { sandbox: Sandbox; oldState: SandboxState; newState: SandboxState }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return
      await syncSandboxFromEvent({
        type: 'state.updated',
        sandbox: data.sandbox,
        oldState: data.oldState,
        newState: data.newState,
      })
    }

    const handleDesiredStateUpdated = async (data: {
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return
      await syncSandboxFromEvent({
        type: 'desired-state.updated',
        sandbox: data.sandbox,
        oldDesiredState: data.oldDesiredState,
        newDesiredState: data.newDesiredState,
      })
    }

    notificationSocket.on('sandbox.created', handleCreated)
    notificationSocket.on('sandbox.state.updated', handleStateUpdated)
    notificationSocket.on('sandbox.desired-state.updated', handleDesiredStateUpdated)

    return () => {
      notificationSocket.off('sandbox.created', handleCreated)
      notificationSocket.off('sandbox.state.updated', handleStateUpdated)
      notificationSocket.off('sandbox.desired-state.updated', handleDesiredStateUpdated)
    }
  }, [enabled, notificationSocket, selectedOrganization?.id, sandboxId, queryClient])
}

export function useSandboxDetailsWsSync(sandboxId?: string) {
  const { selectedOrganization } = useSelectedOrganization()

  useSandboxWsSync({
    enabled: Boolean(selectedOrganization?.id && sandboxId),
    sandboxId,
    queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId ?? ''),
    sync: (oldData, sandbox, event) => {
      if (!oldData) {
        return sandbox
      }

      if (event.type === 'desired-state.updated') {
        if (
          event.newDesiredState === SandboxDesiredState.DESTROYED &&
          (sandbox.state === SandboxState.ERROR || sandbox.state === SandboxState.BUILD_FAILED)
        ) {
          return {
            ...oldData,
            ...sandbox,
            state: SandboxState.DESTROYED,
          }
        }

        const { state: _ignoredState, ...sandboxWithoutState } = sandbox
        return {
          ...oldData,
          ...sandboxWithoutState,
        }
      }

      return {
        ...oldData,
        ...sandbox,
      }
    },
  })
}
