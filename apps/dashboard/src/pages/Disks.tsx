/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Plus } from 'lucide-react'
import React, { useEffect, useState, useCallback, useMemo } from 'react'
import { toast } from 'sonner'
import { OrganizationRolePermissionsEnum, DiskDto, DiskState } from '@daytonaio/api-client'
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
import { DiskTable } from '@/components/DiskTable'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { handleApiError } from '@/lib/error-handling'

const Disks: React.FC = () => {
  const { diskApi } = useApi()
  const { notificationSocket } = useNotificationSocket()

  const [disks, setDisks] = useState<DiskDto[]>([])
  const [loadingDisks, setLoadingDisks] = useState(true)

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newDiskName, setNewDiskName] = useState('')
  const [newDiskSize, setNewDiskSize] = useState<number>(10)
  const [loadingCreate, setLoadingCreate] = useState(false)

  const [diskToDelete, setDiskToDelete] = useState<DiskDto | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [processingDiskAction, setProcessingDiskAction] = useState<Record<string, boolean>>({})

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchDisks = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingDisks(true)
      }
      try {
        const disks = (await diskApi.listDisks(selectedOrganization.id)).data
        setDisks(disks)
      } catch (error) {
        handleApiError(error, 'Failed to fetch disks')
      } finally {
        setLoadingDisks(false)
      }
    },
    [diskApi, selectedOrganization],
  )

  useEffect(() => {
    fetchDisks()
  }, [fetchDisks])

  useEffect(() => {
    const handleDiskCreatedEvent = (disk: DiskDto) => {
      // Refresh the disks list to get the latest data from the server
      fetchDisks(false) // false = don't show loading state
    }

    const handleDiskStateUpdatedEvent = (data: { disk: DiskDto; oldState: DiskState; newState: DiskState }) => {
      if (data.newState === DiskState.STORED) {
        // Disk is being deleted
        setDisks((prev) => prev.filter((d) => d.id !== data.disk.id))
      } else {
        setDisks((prev) => {
          const existingDisk = prev.find((d) => d.id === data.disk.id)
          if (existingDisk) {
            return prev.map((d) => (d.id === data.disk.id ? data.disk : d))
          } else {
            return [data.disk, ...prev]
          }
        })
      }
    }

    if (!notificationSocket) {
      return
    }

    notificationSocket.on('disk.created', handleDiskCreatedEvent)
    notificationSocket.on('disk.state.updated', handleDiskStateUpdatedEvent)

    return () => {
      notificationSocket.off('disk.created', handleDiskCreatedEvent)
      notificationSocket.off('disk.state.updated', handleDiskStateUpdatedEvent)
    }
  }, [notificationSocket, fetchDisks])

  const handleCreate = async () => {
    if (!newDiskName.trim()) {
      toast.error('Disk name is required')
      return
    }

    if (newDiskSize < 1 || newDiskSize > 100) {
      toast.error('Disk size must be between 1 and 100 GB')
      return
    }

    if (!Number.isInteger(newDiskSize)) {
      toast.error('Disk size must be a whole number')
      return
    }

    setLoadingCreate(true)
    try {
      await diskApi.createDisk(
        {
          name: newDiskName,
          size: newDiskSize,
        },
        selectedOrganization?.id,
      )
      setShowCreateDialog(false)
      setNewDiskName('')
      setNewDiskSize(10)
      toast.success(`Creating disk ${newDiskName}`)
      // Refresh the disks list to show the new disk
      await fetchDisks(false)
    } catch (error) {
      handleApiError(error, 'Failed to create disk')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDelete = async (disk: DiskDto) => {
    setProcessingDiskAction((prev) => ({ ...prev, [disk.id]: true }))

    // Optimistically update the disk state
    setDisks((prev) => prev.map((d) => (d.id === disk.id ? { ...d, state: DiskState.STORED } : d)))

    try {
      await diskApi.deleteDisk(disk.id, selectedOrganization?.id)
      setDiskToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting disk ${disk.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete disk')
      // Revert the optimistic update
      setDisks((prev) => prev.map((d) => (d.id === disk.id ? { ...d, state: disk.state } : d)))
    } finally {
      setProcessingDiskAction((prev) => ({ ...prev, [disk.id]: false }))
    }
  }

  const handleBulkDelete = async (disks: DiskDto[]) => {
    setProcessingDiskAction((prev) => ({ ...prev, ...disks.reduce((acc, d) => ({ ...acc, [d.id]: true }), {}) }))

    for (const disk of disks) {
      // Optimistically update the disk state
      setDisks((prev) => prev.map((d) => (d.id === disk.id ? { ...d, state: DiskState.STORED } : d)))

      try {
        await diskApi.deleteDisk(disk.id, selectedOrganization?.id)
        toast.success(`Deleting disk ${disk.name}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete disk')

        // Revert the optimistic update
        setDisks((prev) => prev.map((d) => (d.id === disk.id ? { ...d, state: disk.state } : d)))

        // Wait for user decision
        const shouldContinue = window.confirm(
          `Failed to delete disk ${disk.name}. Do you want to continue with the remaining disks?`,
        )

        if (!shouldContinue) {
          break
        }
      } finally {
        setProcessingDiskAction((prev) => ({ ...prev, [disk.id]: false }))
      }
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES),
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
          setNewDiskName('')
          setNewDiskSize(10)
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Disks</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button variant="default" size="sm" disabled={loadingDisks} className="w-auto px-4" title="Create Disk">
                <Plus className="w-4 h-4" />
                Create Disk
              </Button>
            </DialogTrigger>
          )}
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Disk</DialogTitle>
              <DialogDescription>Create a new disk for your sandboxes</DialogDescription>
            </DialogHeader>
            <form
              id="create-disk-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleCreate()
              }}
            >
              <div className="space-y-3">
                <Label htmlFor="name">Disk Name</Label>
                <Input
                  id="name"
                  value={newDiskName}
                  onChange={(e) => setNewDiskName(e.target.value)}
                  placeholder="my-disk"
                />
              </div>
              <div className="space-y-3">
                <Label htmlFor="size">Size (GB)</Label>
                <Input
                  id="size"
                  type="number"
                  min="1"
                  max="100"
                  step="1"
                  value={newDiskSize}
                  onChange={(e) => setNewDiskSize(Number(e.target.value))}
                  placeholder="10"
                />
                <p className="text-sm text-muted-foreground">Size must be between 1 and 100 GB (whole numbers only)</p>
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
                  form="create-disk-form"
                  variant="default"
                  disabled={
                    !newDiskName.trim() || newDiskSize < 1 || newDiskSize > 100 || !Number.isInteger(newDiskSize)
                  }
                >
                  Create
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </div>

        <DiskTable
          data={disks}
          loading={loadingDisks}
          processingDiskAction={processingDiskAction}
          onDelete={(disk) => {
            setDiskToDelete(disk)
            setShowDeleteDialog(true)
          }}
          onBulkDelete={handleBulkDelete}
        />
      </Dialog>

      {diskToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setDiskToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Disk Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this Disk? This action cannot be undone.
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
                onClick={() => handleDelete(diskToDelete)}
                disabled={processingDiskAction[diskToDelete.id]}
              >
                {processingDiskAction[diskToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Disks
