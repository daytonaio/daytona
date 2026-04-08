/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { queryKeys } from '@/hooks/queries/queryKeys'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { VolumeDto, VolumeState } from '@daytona/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

export function useVolumeWsSync() {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!notificationSocket || !selectedOrganization?.id) return

    const queryKey = queryKeys.volumes.list(selectedOrganization.id)

    const upsertVolumeInCache = (volume: VolumeDto) => {
      queryClient.setQueriesData<VolumeDto[]>({ queryKey }, (previousVolumes) => {
        if (!previousVolumes) return [volume]

        if (!previousVolumes.some((existingVolume) => existingVolume.id === volume.id)) {
          return [volume, ...previousVolumes]
        }

        return previousVolumes.map((existingVolume) => (existingVolume.id === volume.id ? volume : existingVolume))
      })
    }

    const removeVolumeFromCache = (volumeId: string) => {
      queryClient.setQueriesData<VolumeDto[]>({ queryKey }, (previousVolumes) => {
        if (!previousVolumes) return []
        return previousVolumes.filter((volume) => volume.id !== volumeId)
      })
    }

    const invalidate = (refetchType: 'active' | 'none' = 'none') => {
      queryClient.invalidateQueries({
        queryKey,
        refetchType,
      })
    }

    const handleVolumeCreatedEvent = (volume: VolumeDto) => {
      upsertVolumeInCache(volume)
      invalidate('active')
    }

    const handleVolumeStateUpdatedEvent = (data: {
      volume: VolumeDto
      oldState: VolumeState
      newState: VolumeState
    }) => {
      if (data.newState === VolumeState.DELETED) {
        removeVolumeFromCache(data.volume.id)
        invalidate('active')
      } else {
        upsertVolumeInCache(data.volume)
        invalidate()
      }
    }

    const handleVolumeLastUsedAtUpdatedEvent = (volume: VolumeDto) => {
      upsertVolumeInCache(volume)
      invalidate()
    }

    notificationSocket.on('volume.created', handleVolumeCreatedEvent)
    notificationSocket.on('volume.state.updated', handleVolumeStateUpdatedEvent)
    notificationSocket.on('volume.lastUsedAt.updated', handleVolumeLastUsedAtUpdatedEvent)

    return () => {
      notificationSocket.off('volume.created', handleVolumeCreatedEvent)
      notificationSocket.off('volume.state.updated', handleVolumeStateUpdatedEvent)
      notificationSocket.off('volume.lastUsedAt.updated', handleVolumeLastUsedAtUpdatedEvent)
    }
  }, [notificationSocket, queryClient, selectedOrganization?.id])
}
