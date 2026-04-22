/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Sandbox, SandboxState } from '@daytona/api-client'
import type { QueryKey } from '@tanstack/react-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

interface UseSandboxWsSyncOptions<TData> {
  sandboxId?: string
  queryKey: QueryKey
  sync: (oldData: TData | undefined, sandbox: Sandbox) => TData | undefined
}

export function useSandboxWsSync({ sandboxId, queryKey, sync }: UseSandboxWsSyncOptions<Sandbox>) {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()
  const queryKeyRef = useRef(queryKey)
  const syncRef = useRef(sync)

  queryKeyRef.current = queryKey
  syncRef.current = sync

  useEffect(() => {
    if (!notificationSocket || !selectedOrganization?.id) return

    const cancelSandboxQuery = () => {
      queryClient.cancelQueries({
        queryKey: queryKeyRef.current,
      })
    }

    const syncSandboxInCache = (sandbox: Sandbox) => {
      queryClient.setQueryData<Sandbox>(queryKeyRef.current, (oldData) => syncRef.current(oldData, sandbox))
    }

    const syncSandboxFromEvent = async (sandbox: Sandbox) => {
      cancelSandboxQuery()
      syncSandboxInCache(sandbox)
    }

    const handleCreated = async (sandbox: Sandbox) => {
      if (sandboxId && sandbox.id !== sandboxId) return
      await syncSandboxFromEvent(sandbox)
    }

    const handleStateUpdated = async (data: { sandbox: Sandbox; oldState: SandboxState; newState: SandboxState }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return
      await syncSandboxFromEvent(data.sandbox)
    }

    const handleDesiredStateUpdated = async (data: { sandbox: Sandbox }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return
      await syncSandboxFromEvent(data.sandbox)
    }

    notificationSocket.on('sandbox.created', handleCreated)
    notificationSocket.on('sandbox.state.updated', handleStateUpdated)
    notificationSocket.on('sandbox.desired-state.updated', handleDesiredStateUpdated)

    return () => {
      notificationSocket.off('sandbox.created', handleCreated)
      notificationSocket.off('sandbox.state.updated', handleStateUpdated)
      notificationSocket.off('sandbox.desired-state.updated', handleDesiredStateUpdated)
    }
  }, [notificationSocket, selectedOrganization?.id, sandboxId, queryClient])
}

export function useSandboxDetailsWsSync(sandboxId?: string) {
  const { selectedOrganization } = useSelectedOrganization()

  useSandboxWsSync({
    sandboxId,
    queryKey: queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId ?? ''),
    sync: (oldData, sandbox) => {
      if (!oldData) {
        return sandbox
      }

      return {
        ...oldData,
        ...sandbox,
      }
    },
  })
}
