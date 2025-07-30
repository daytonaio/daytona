/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Plus } from 'lucide-react'
import React, { useEffect, useState, useCallback, useMemo } from 'react'
import { toast } from 'sonner'
import { OrganizationRolePermissionsEnum, VolumeDto, VolumeState } from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { VolumeTable } from '@/components/VolumeTable'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { handleApiError } from '@/lib/error-handling'

const Volumes: React.FC = () => {
  const { volumeApi } = useApi()
  const { notificationSocket } = useNotificationSocket()

  const [volumes, setVolumes] = useState<VolumeDto[]>([])
  const [loadingVolumes, setLoadingVolumes] = useState(true)

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newVolumeName, setNewVolumeName] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)

  const [volumeToDelete, setVolumeToDelete] = useState<VolumeDto | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [processingVolumeAction, setProcessingVolumeAction] = useState<Record<string, boolean>>({})

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchVolumes = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingVolumes(true)
      }
      try {
        const volumes = (await volumeApi.listVolumes(selectedOrganization.id)).data
        setVolumes(volumes)
      } catch (error) {
        handleApiError(error, 'Failed to fetch volumes')
      } finally {
        setLoadingVolumes(false)
      }
    },
    [volumeApi, selectedOrganization],
  )

  useEffect(() => {
    fetchVolumes()
  }, [fetchVolumes])

  useEffect(() => {
    const handleVolumeCreatedEvent = (volume: VolumeDto) => {
      if (!volumes.some((v) => v.id === volume.id)) {
        setVolumes((prev) => [volume, ...prev])
      }
    }

    const handleVolumeStateUpdatedEvent = (data: {
      volume: VolumeDto
      oldState: VolumeState
      newState: VolumeState
    }) => {
      if (data.newState === VolumeState.DELETED) {
        setVolumes((prev) => prev.filter((v) => v.id !== data.volume.id))
      } else if (!volumes.some((v) => v.id === data.volume.id)) {
        setVolumes((prev) => [data.volume, ...prev])
      } else {
        setVolumes((prev) => prev.map((v) => (v.id === data.volume.id ? data.volume : v)))
      }
    }

    const handleVolumeLastUsedAtUpdatedEvent = (volume: VolumeDto) => {
      if (!volumes.some((v) => v.id === volume.id)) {
        setVolumes((prev) => [volume, ...prev])
      } else {
        setVolumes((prev) => prev.map((v) => (v.id === volume.id ? volume : v)))
      }
    }

    if (!notificationSocket) {
      return
    }

    notificationSocket.on('volume.created', handleVolumeCreatedEvent)
    notificationSocket.on('volume.state.updated', handleVolumeStateUpdatedEvent)
    notificationSocket.on('volume.lastUsedAt.updated', handleVolumeLastUsedAtUpdatedEvent)

    return () => {
      notificationSocket.off('volume.created', handleVolumeCreatedEvent)
      notificationSocket.off('volume.state.updated', handleVolumeStateUpdatedEvent)
      notificationSocket.off('volume.lastUsedAt.updated', handleVolumeLastUsedAtUpdatedEvent)
    }
  }, [notificationSocket, volumes])

  const handleCreate = async () => {
    setLoadingCreate(true)
    try {
      await volumeApi.createVolume(
        {
          name: newVolumeName,
        },
        selectedOrganization?.id,
      )
      setShowCreateDialog(false)
      setNewVolumeName('')
      toast.success(`Creating volume ${newVolumeName}`)
    } catch (error) {
      handleApiError(error, 'Failed to create volume')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDelete = async (volume: VolumeDto) => {
    setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: true }))

    // Optimistically update the volume state
    setVolumes((prev) => prev.map((v) => (v.id === volume.id ? { ...v, state: VolumeState.PENDING_DELETE } : v)))

    try {
      await volumeApi.deleteVolume(volume.id, selectedOrganization?.id)
      setVolumeToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting volume ${volume.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete volume')
      // Revert the optimistic update
      setVolumes((prev) => prev.map((v) => (v.id === volume.id ? { ...v, state: volume.state } : v)))
    } finally {
      setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: false }))
    }
  }

  const handleBulkDelete = async (volumes: VolumeDto[]) => {
    setProcessingVolumeAction((prev) => ({ ...prev, ...volumes.reduce((acc, v) => ({ ...acc, [v.id]: true }), {}) }))

    for (const volume of volumes) {
      // Optimistically update the volume state
      setVolumes((prev) => prev.map((v) => (v.id === volume.id ? { ...v, state: VolumeState.PENDING_DELETE } : v)))

      try {
        await volumeApi.deleteVolume(volume.id, selectedOrganization?.id)
        toast.success(`Deleting volume ${volume.name}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete volume')

        // Revert the optimistic update
        setVolumes((prev) => prev.map((v) => (v.id === volume.id ? { ...v, state: volume.state } : v)))

        // Wait for user decision
        const shouldContinue = window.confirm(
          `Failed to delete volume ${volume.name}. Do you want to continue with the remaining volumes?`,
        )

        if (!shouldContinue) {
          break
        }
      } finally {
        setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: false }))
      }
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_VOLUMES),
    [authenticatedUserHasPermission],
  )

  return (
    <div className="px-6 py-2">
      <Dialog
        open={showCreateDialog}
        onOpenChange={(isOpen) => {
          setShowCreateDialog(isOpen)
          if (isOpen) {
            return
          }
          setNewVolumeName('')
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Volumes</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button
                variant="default"
                size="sm"
                disabled={loadingVolumes}
                className="w-auto px-4"
                title="Create Volume"
              >
                <Plus className="w-4 h-4" />
                Create Volume
              </Button>
            </DialogTrigger>
          )}
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Volume</DialogTitle>
              <DialogDescription>Instantly Access Shared Files with Volume Mounts</DialogDescription>
            </DialogHeader>
            <form
              id="create-volume-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleCreate()
              }}
            >
              <div className="space-y-3">
                <Label htmlFor="name">Volume Name</Label>
                <Input
                  id="name"
                  value={newVolumeName}
                  onChange={(e) => setNewVolumeName(e.target.value)}
                  placeholder="my-volume"
                />
              </div>
            </form>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" size="sm" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              {loadingCreate ? (
                <Button type="button" size="sm" variant="default" disabled>
                  Creating...
                </Button>
              ) : (
                <Button
                  type="submit"
                  size="sm"
                  form="create-volume-form"
                  variant="default"
                  disabled={!newVolumeName.trim()}
                >
                  Create
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </div>

        <VolumeTable
          data={volumes}
          loading={loadingVolumes}
          processingVolumeAction={processingVolumeAction}
          onDelete={(volume) => {
            setVolumeToDelete(volume)
            setShowDeleteDialog(true)
          }}
          onBulkDelete={handleBulkDelete}
        />
      </Dialog>

      {volumeToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setVolumeToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Volume Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this Volume? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              <Button
                variant="destructive"
                onClick={() => handleDelete(volumeToDelete)}
                disabled={processingVolumeAction[volumeToDelete.id]}
              >
                {processingVolumeAction[volumeToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Volumes
