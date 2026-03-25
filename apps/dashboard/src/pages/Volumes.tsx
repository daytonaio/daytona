/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
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
import { useCreateVolumeMutation } from '@/hooks/mutations/useCreateVolumeMutation'
import { useDeleteVolumeMutation } from '@/hooks/mutations/useDeleteVolumeMutation'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { useVolumesQuery } from '@/hooks/queries/useVolumesQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useVolumeWsSync } from '@/hooks/useVolumeWsSync'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { pluralize } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, VolumeDto, VolumeState } from '@daytonaio/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'

const Volumes: React.FC = () => {
  const queryClient = useQueryClient()

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newVolumeName, setNewVolumeName] = useState('')

  const [volumeToDelete, setVolumeToDelete] = useState<VolumeDto | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [processingVolumeAction, setProcessingVolumeAction] = useState<Record<string, boolean>>({})

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  useVolumeWsSync()

  const queryKey = useMemo(() => queryKeys.volumes.list(selectedOrganization?.id ?? ''), [selectedOrganization?.id])
  const { data: volumes = [], isLoading: loadingVolumes, error: volumesError } = useVolumesQuery()
  const createVolumeMutation = useCreateVolumeMutation()
  const deleteVolumeMutation = useDeleteVolumeMutation({ invalidateOnSuccess: false })

  useEffect(() => {
    if (volumesError) {
      handleApiError(volumesError, 'Failed to fetch volumes')
    }
  }, [volumesError])

  const updateVolumeStateInCache = useCallback(
    (volumeId: string, state: VolumeState) => {
      queryClient.setQueriesData<VolumeDto[]>({ queryKey }, (previousVolumes) => {
        if (!previousVolumes) return previousVolumes

        return previousVolumes.map((volume) => (volume.id === volumeId ? { ...volume, state } : volume))
      })
    },
    [queryClient, queryKey],
  )

  const handleCreate = async () => {
    const volumeName = newVolumeName.trim()
    if (!volumeName) {
      toast.error('Volume name is required')
      return
    }

    try {
      await createVolumeMutation.mutateAsync({
        volume: {
          name: volumeName,
        },
        organizationId: selectedOrganization?.id,
      })
      setShowCreateDialog(false)
      setNewVolumeName('')
      toast.success(`Creating volume ${volumeName}`)
    } catch (error) {
      handleApiError(error, 'Failed to create volume')
    }
  }

  const handleDelete = async (volume: VolumeDto) => {
    setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: true }))

    updateVolumeStateInCache(volume.id, VolumeState.PENDING_DELETE)

    try {
      await deleteVolumeMutation.mutateAsync({
        volumeId: volume.id,
        organizationId: selectedOrganization?.id,
      })
      if (selectedOrganization?.id) {
        await queryClient.invalidateQueries({ queryKey })
      }
      setVolumeToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting volume ${volume.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete volume')
      updateVolumeStateInCache(volume.id, volume.state)
    } finally {
      setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: false }))
    }
  }

  const handleBulkDelete = async (volumes: VolumeDto[]) => {
    const previousStatesById = new Map(volumes.map((volume) => [volume.id, volume.state]))
    let isCancelled = false
    let processedCount = 0
    let successCount = 0
    let failureCount = 0

    const totalLabel = pluralize(volumes.length, 'volume', 'volumes')
    const onCancel = () => {
      isCancelled = true
    }

    const bulkToast = createBulkActionToast(`Deleting 0 of ${totalLabel}.`, {
      action: { label: 'Cancel', onClick: onCancel },
    })

    try {
      for (const volume of volumes) {
        if (isCancelled) break

        processedCount += 1
        bulkToast.loading(`Deleting ${processedCount} of ${totalLabel}.`, {
          action: { label: 'Cancel', onClick: onCancel },
        })

        setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: true }))
        updateVolumeStateInCache(volume.id, VolumeState.PENDING_DELETE)

        try {
          await deleteVolumeMutation.mutateAsync({
            volumeId: volume.id,
            organizationId: selectedOrganization?.id,
          })
          successCount += 1
        } catch (error) {
          failureCount += 1
          updateVolumeStateInCache(volume.id, previousStatesById.get(volume.id) ?? volume.state)
          console.error('Deleting volume failed', volume.id, error)
        } finally {
          setProcessingVolumeAction((prev) => ({ ...prev, [volume.id]: false }))
        }
      }

      if (selectedOrganization?.id) {
        await queryClient.invalidateQueries({ queryKey })
      }

      bulkToast.result(
        { successCount, failureCount },
        {
          successTitle: `${pluralize(volumes.length, 'Volume', 'Volumes')} deleted.`,
          errorTitle: `Failed to delete ${pluralize(volumes.length, 'volume', 'volumes')}.`,
          warningTitle: 'Failed to delete some volumes.',
          canceledTitle: 'Delete canceled.',
        },
      )
    } catch (error) {
      console.error('Deleting volumes failed', error)
      bulkToast.error('Deleting volumes failed.')
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_VOLUMES),
    [authenticatedUserHasPermission],
  )

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Volumes</PageTitle>
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
          {writePermitted && (
            <DialogTrigger asChild>
              <Button variant="default" size="sm" disabled={loadingVolumes} className="ml-auto" title="Create Volume">
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
              {createVolumeMutation.isPending ? (
                <Button type="button" size="sm" variant="default" disabled>
                  Creating...
                </Button>
              ) : (
                <Button
                  type="submit"
                  size="sm"
                  form="create-volume-form"
                  variant="default"
                  disabled={!newVolumeName.trim() || createVolumeMutation.isPending}
                >
                  Create
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </PageHeader>

      <PageContent size="full">
        <VolumeTable
          data={volumes}
          loading={loadingVolumes}
          processingVolumeAction={processingVolumeAction}
          onCreateVolume={() => setShowCreateDialog(true)}
          onDelete={(volume) => {
            setVolumeToDelete(volume)
            setShowDeleteDialog(true)
          }}
          onBulkDelete={handleBulkDelete}
        />

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
      </PageContent>
    </PageLayout>
  )
}

export default Volumes
