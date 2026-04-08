/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { queryKeys } from '@/hooks/queries/queryKeys'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { PaginatedSnapshots, SnapshotDto, SnapshotState } from '@daytona/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

export function useSnapshotWsSync() {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!notificationSocket || !selectedOrganization?.id) return

    const queryKey = queryKeys.snapshots.list(selectedOrganization.id)

    const updateSnapshotInCacheIfPresent = (snapshot: SnapshotDto) => {
      queryClient.setQueriesData<PaginatedSnapshots>({ queryKey }, (previousSnapshots) => {
        if (!previousSnapshots) return previousSnapshots
        if (!previousSnapshots.items.some((existingSnapshot) => existingSnapshot.id === snapshot.id))
          return previousSnapshots

        return {
          ...previousSnapshots,
          items: previousSnapshots.items.map((existingSnapshot) =>
            existingSnapshot.id === snapshot.id ? snapshot : existingSnapshot,
          ),
        }
      })
    }

    const invalidate = (refetchType: 'active' | 'none' = 'none') => {
      queryClient.invalidateQueries({
        queryKey,
        refetchType,
      })
    }

    const handleSnapshotCreatedEvent = () => {
      invalidate('active')
    }

    const handleSnapshotStateUpdatedEvent = (data: {
      snapshot: SnapshotDto
      oldState: SnapshotState
      newState: SnapshotState
    }) => {
      updateSnapshotInCacheIfPresent(data.snapshot)
      invalidate()
    }

    const handleSnapshotRemovedEvent = () => {
      invalidate('active')
    }

    notificationSocket.on('snapshot.created', handleSnapshotCreatedEvent)
    notificationSocket.on('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
    notificationSocket.on('snapshot.removed', handleSnapshotRemovedEvent)

    return () => {
      notificationSocket.off('snapshot.created', handleSnapshotCreatedEvent)
      notificationSocket.off('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
      notificationSocket.off('snapshot.removed', handleSnapshotRemovedEvent)
    }
  }, [notificationSocket, queryClient, selectedOrganization?.id])
}
