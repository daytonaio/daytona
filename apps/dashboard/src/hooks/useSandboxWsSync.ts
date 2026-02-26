/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getSandboxesQueryKey } from '@/hooks/useSandboxes'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { PaginatedSandboxes, Sandbox, SandboxDesiredState, SandboxState } from '@daytonaio/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

interface UseSandboxWsSyncOptions {
  sandboxId?: string
  refetchOnCreate?: boolean
}

export function useSandboxWsSync({ sandboxId, refetchOnCreate = false }: UseSandboxWsSyncOptions = {}) {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!notificationSocket) return

    const orgId = selectedOrganization?.id ?? ''

    const updateStateInListCache = (targetId: string, state: SandboxState) => {
      queryClient.setQueriesData<PaginatedSandboxes>({ queryKey: getSandboxesQueryKey(orgId) }, (oldData) => {
        if (!oldData) return oldData
        return {
          ...oldData,
          items: oldData.items.map((s) => (s.id === targetId ? { ...s, state } : s)),
        }
      })
    }

    const updateStateInDetailCache = (targetId: string, state: SandboxState) => {
      queryClient.setQueryData<Sandbox>(queryKeys.sandboxes.detail(orgId, targetId), (oldData) => {
        if (!oldData) return oldData
        return { ...oldData, state }
      })
    }

    const optimisticUpdate = (targetId: string, state: SandboxState) => {
      updateStateInListCache(targetId, state)
      if (sandboxId) {
        updateStateInDetailCache(targetId, state)
      }
    }

    const invalidate = () => {
      queryClient.invalidateQueries({
        queryKey: getSandboxesQueryKey(orgId),
        refetchType: 'none',
      })

      if (sandboxId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.sandboxes.detail(orgId, sandboxId),
        })
      }
    }

    const handleCreated = (_sandbox: Sandbox) => {
      if (sandboxId) return

      queryClient.invalidateQueries({
        queryKey: getSandboxesQueryKey(orgId),
        refetchType: refetchOnCreate ? 'active' : 'none',
      })
    }

    const handleStateUpdated = (data: { sandbox: Sandbox; oldState: SandboxState; newState: SandboxState }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return

      // warm pool sandboxes — treat as created
      if (data.oldState === data.newState && data.newState === SandboxState.STARTED) {
        handleCreated(data.sandbox)
        return
      }

      let updatedState = data.newState

      // error/build_failed with desiredState=DESTROYED should display as destroyed
      if (
        data.sandbox.desiredState === SandboxDesiredState.DESTROYED &&
        (data.newState === SandboxState.ERROR || data.newState === SandboxState.BUILD_FAILED)
      ) {
        updatedState = SandboxState.DESTROYED
      }

      optimisticUpdate(data.sandbox.id, updatedState)
      invalidate()
    }

    const handleDesiredStateUpdated = (data: {
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }) => {
      if (sandboxId && data.sandbox.id !== sandboxId) return

      if (data.newDesiredState !== SandboxDesiredState.DESTROYED) return
      if (data.sandbox.state !== SandboxState.ERROR && data.sandbox.state !== SandboxState.BUILD_FAILED) return

      optimisticUpdate(data.sandbox.id, SandboxState.DESTROYED)
      invalidate()
    }

    notificationSocket.on('sandbox.created', handleCreated)
    notificationSocket.on('sandbox.state.updated', handleStateUpdated)
    notificationSocket.on('sandbox.desired-state.updated', handleDesiredStateUpdated)

    return () => {
      notificationSocket.off('sandbox.created', handleCreated)
      notificationSocket.off('sandbox.state.updated', handleStateUpdated)
      notificationSocket.off('sandbox.desired-state.updated', handleDesiredStateUpdated)
    }
  }, [notificationSocket, selectedOrganization?.id, sandboxId, refetchOnCreate, queryClient])
}
