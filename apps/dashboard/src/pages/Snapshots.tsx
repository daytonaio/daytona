/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import {
  SnapshotDto,
  SnapshotState,
  OrganizationRolePermissionsEnum,
  PaginatedSnapshotsDto,
} from '@daytonaio/api-client'
import { SnapshotTable } from '@/components/SnapshotTable'
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
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'

const Snapshots: React.FC = () => {
  const { notificationSocket } = useNotificationSocket()

  const { snapshotApi } = useApi()
  const [snapshotsData, setSnapshotsData] = useState<PaginatedSnapshotsDto>({
    items: [],
    total: 0,
    page: 1,
    totalPages: 0,
  })
  const [loadingSnapshots, setLoadingSnapshots] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [snapshotToDelete, setSnapshotToDelete] = useState<SnapshotDto | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newSnapshotName, setNewSnapshotName] = useState('')
  const [newEntrypoint, setNewEntrypoint] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

  const fetchSnapshots = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const response = (
          await snapshotApi.getAllSnapshots(
            selectedOrganization.id,
            paginationParams.pageSize,
            paginationParams.pageIndex + 1,
          )
        ).data
        setSnapshotsData(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch snapshots')
      } finally {
        setLoadingTable(false)
      }
    },
    [snapshotApi, selectedOrganization, paginationParams.pageIndex, paginationParams.pageSize],
  )

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  useEffect(() => {
    fetchSnapshots()
  }, [fetchSnapshots])

  useEffect(() => {
    const handleSnapshotCreatedEvent = (snapshot: SnapshotDto) => {
      if (paginationParams.pageIndex === 0) {
        setSnapshotsData((prev) => {
          if (prev.items.some((i) => i.id === snapshot.id)) {
            return prev
          }

          // Find the insertion point - used snapshots should remain at the top
          const insertIndex =
            prev.items.findIndex((i) => !i.lastUsedAt && i.createdAt <= snapshot.createdAt) || prev.items.length

          const newSnapshots = [...prev.items]
          newSnapshots.splice(insertIndex, 0, snapshot)

          const newTotal = prev.total + 1
          return {
            ...prev,
            items: newSnapshots.slice(0, paginationParams.pageSize),
            total: newTotal,
            totalPages: Math.ceil(newTotal / paginationParams.pageSize),
          }
        })
      }
    }

    const handleSnapshotStateUpdatedEvent = (data: {
      snapshot: SnapshotDto
      oldState: SnapshotState
      newState: SnapshotState
    }) => {
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === data.snapshot.id ? data.snapshot : i)),
      }))
    }

    const handleSnapshotEnabledToggledEvent = (snapshot: SnapshotDto) => {
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === snapshot.id ? snapshot : i)),
      }))
    }

    const handleSnapshotRemovedEvent = (snapshotId: string) => {
      setSnapshotsData((prev) => {
        const newTotal = Math.max(0, prev.total - 1)
        const newItems = prev.items.filter((i) => i.id !== snapshotId)

        return {
          ...prev,
          items: newItems,
          total: newTotal,
          totalPages: Math.ceil(newTotal / paginationParams.pageSize),
        }
      })
    }

    notificationSocket.on('snapshot.created', handleSnapshotCreatedEvent)
    notificationSocket.on('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
    notificationSocket.on('snapshot.enabled.toggled', handleSnapshotEnabledToggledEvent)
    notificationSocket.on('snapshot.removed', handleSnapshotRemovedEvent)

    return () => {
      notificationSocket.off('snapshot.created', handleSnapshotCreatedEvent)
      notificationSocket.off('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
      notificationSocket.off('snapshot.enabled.toggled', handleSnapshotEnabledToggledEvent)
      notificationSocket.off('snapshot.removed', handleSnapshotRemovedEvent)
    }
  }, [notificationSocket, paginationParams.pageIndex, paginationParams.pageSize])

  useEffect(() => {
    if (snapshotsData.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [snapshotsData.items.length, paginationParams.pageIndex])

  const validateSnapshotName = (name: string): string | null => {
    // Basic format check
    const snapshotNameRegex =
      /^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:\/[a-z0-9]+(?:[._-][a-z0-9]+)*)*:[a-z0-9]+(?:[._-][a-z0-9]+)*$/

    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Snapshot name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Snapshots with tag ":latest" are not allowed'
    }

    if (!snapshotNameRegex.test(name)) {
      return 'Invalid snapshot name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  const handleCreate = async () => {
    const validationError = validateSnapshotName(newSnapshotName)
    if (validationError) {
      toast.warning(validationError)
      return
    }

    setLoadingCreate(true)
    try {
      await snapshotApi.createSnapshot(
        {
          name: newSnapshotName,
          entrypoint: newEntrypoint.trim() ? newEntrypoint.trim().split(' ') : undefined,
        },
        selectedOrganization?.id,
      )
      setShowCreateDialog(false)
      setNewSnapshotName('')
      setNewEntrypoint('')
      toast.success(`Creating snapshot ${newSnapshotName}`)

      if (paginationParams.pageIndex !== 0) {
        setPaginationParams((prev) => ({
          ...prev,
          pageIndex: 0,
        }))
      }
    } catch (error) {
      handleApiError(error, 'Failed to create snapshot')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDelete = async (snapshot: SnapshotDto) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))

    // Optimistically update the snapshot state
    setSnapshotsData((prev) => ({
      ...prev,
      items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: SnapshotState.REMOVING } : i)),
    }))

    try {
      await snapshotApi.removeSnapshot(snapshot.id, selectedOrganization?.id)
      setSnapshotToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete snapshot')
      // Revert the optimistic update
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: snapshot.state } : i)),
      }))
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const handleToggleEnabled = async (snapshot: SnapshotDto, enabled: boolean) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))

    // Optimistically update the snapshot enabled flag
    setSnapshotsData((prev) => ({
      ...prev,
      items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, enabled } : i)),
    }))

    try {
      await snapshotApi.toggleSnapshotState(snapshot.id, { enabled }, selectedOrganization?.id)
      toast.success(`${enabled ? 'Enabling' : 'Disabling'} snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, enabled ? 'Failed to enable snapshot' : 'Failed to disable snapshot')
      // Revert the optimistic update
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, enabled: snapshot.enabled } : i)),
      }))
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const handleBulkDelete = async (images: ImageDto[]) => {
    setLoadingImages((prev) => ({ ...prev, ...images.reduce((acc, img) => ({ ...acc, [img.id]: true }), {}) }))

    for (const image of images) {
      setImagesData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === image.id ? { ...i, state: ImageState.REMOVING } : i)),
      }))

      try {
        await imageApi.removeImage(image.id, selectedOrganization?.id)
        toast.success(`Deleting image ${image.name}`)
      } catch (error) {
        handleApiError(error, `Failed to delete image ${image.name}`)

        setImagesData((prev) => ({
          ...prev,
          items: prev.items.map((i) => (i.id === image.id ? { ...i, state: image.state } : i)),
        }))

        if (images.indexOf(image) < images.length - 1) {
          const shouldContinue = window.confirm(
            `Failed to delete image ${image.name}. Do you want to continue with the remaining images?`,
          )

          if (!shouldContinue) {
            break
          }
        }
      } finally {
        setLoadingImages((prev) => ({ ...prev, [image.id]: false }))
      }
    }
  }

  return (
    <div className="p-6">
      <Dialog
        open={showCreateDialog}
        onOpenChange={(isOpen) => {
          setShowCreateDialog(isOpen)
          if (isOpen) {
            return
          }
          setNewSnapshotName('')
          setNewEntrypoint('')
        }}
      >
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Snapshots</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button
                variant="default"
                size="icon"
                disabled={loadingTable}
                className="w-auto px-4"
                title="Create Snapshot"
              >
                <Plus className="w-4 h-4" />
                Create Snapshot
              </Button>
            </DialogTrigger>
          )}
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Snapshot</DialogTitle>
              <DialogDescription>
                Register a new snapshot to be used for spinning up sandboxes in your organization.
              </DialogDescription>
            </DialogHeader>
            <form
              id="create-snapshot-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleCreate()
              }}
            >
              <div className="space-y-3">
                <Label htmlFor="name">Snapshot Name</Label>
                <Input
                  id="name"
                  value={newSnapshotName}
                  onChange={(e) => setNewSnapshotName(e.target.value)}
                  placeholder="ubuntu:22.04"
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Must include a tag (e.g., ubuntu:22.04). The tag "latest" is not allowed.
                </p>
              </div>
              <div className="space-y-3">
                <Label htmlFor="entrypoint">Entrypoint (optional)</Label>
                <Input
                  id="entrypoint"
                  value={newEntrypoint}
                  onChange={(e) => setNewEntrypoint(e.target.value)}
                  placeholder="sleep infinity"
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Ensure that the entrypoint is a long running command. If not provided, or if the snapshot does not
                  have an entrypoint, 'sleep infinity' will be used as the default.
                </p>
              </div>
            </form>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              {loadingCreate ? (
                <Button type="button" variant="default" disabled>
                  Creating...
                </Button>
              ) : (
                <Button
                  type="submit"
                  form="create-snapshot-form"
                  variant="default"
                  disabled={!newSnapshotName.trim() || validateSnapshotName(newSnapshotName) !== null}
                >
                  Create
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </div>

        <SnapshotTable
          data={snapshotsData.items}
          loading={loadingTable}
          loadingSnapshots={loadingSnapshots}
          onDelete={(snapshot) => {
            setSnapshotToDelete(snapshot)
            setShowDeleteDialog(true)
          }}
          onBulkDelete={handleBulkDelete}
          onToggleEnabled={handleToggleEnabled}
          pageCount={snapshotsData.totalPages}
          onPaginationChange={handlePaginationChange}
          pagination={{
            pageIndex: paginationParams.pageIndex,
            pageSize: paginationParams.pageSize,
          }}
        />
      </Dialog>

      {snapshotToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setSnapshotToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Snapshot Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this snapshot? This action cannot be undone.
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
                onClick={() => handleDelete(snapshotToDelete)}
                disabled={loadingSnapshots[snapshotToDelete.id]}
              >
                {loadingSnapshots[snapshotToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Snapshots
