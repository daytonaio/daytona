/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus, X } from 'lucide-react'
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
import { Slider } from '@/components/ui/slider'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*$/

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
  const [newImageName, setNewImageName] = useState('')
  const [newEntrypoint, setNewEntrypoint] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [cpu, setCpu] = useState<number | undefined>(undefined)
  const [memory, setMemory] = useState<number | undefined>(undefined)
  const [disk, setDisk] = useState<number | undefined>(undefined)
  const [targetPropagations, setTargetPropagations] = useState<Array<{ target: string; userOverride: number }>>([])
  const [newTarget, setNewTarget] = useState('')
  const [snapshotToEditTargets, setSnapshotToEditTargets] = useState<SnapshotDto | null>(null)
  const [showTargetsDialog, setShowTargetsDialog] = useState(false)
  const [editingTargetPropagations, setEditingTargetPropagations] = useState<
    Array<{ target: string; userOverride: number }>
  >([])
  const [editingNewTarget, setEditingNewTarget] = useState('')
  const [loadingTargetUpdate, setLoadingTargetUpdate] = useState(false)

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

    if (!notificationSocket) {
      return
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
    if (name.includes(' ')) {
      return 'Spaces are not allowed in snapshot names'
    }

    if (!IMAGE_NAME_REGEX.test(name)) {
      return 'Invalid snapshot name format. May contain letters, digits, dots, colons, slashes and dashes'
    }

    return null
  }

  const validateImageName = (name: string): string | null => {
    if (name.includes(' ')) {
      return 'Spaces are not allowed in image names'
    }

    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Image name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Images with tag ":latest" are not allowed'
    }

    if (!IMAGE_NAME_REGEX.test(name)) {
      return 'Invalid image name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  const handleCreate = async () => {
    const nameValidationError = validateSnapshotName(newSnapshotName)
    if (nameValidationError) {
      toast.warning(nameValidationError)
      return
    }

    const imageValidationError = validateImageName(newImageName)
    if (imageValidationError) {
      toast.warning(imageValidationError)
      return
    }

    setLoadingCreate(true)
    try {
      await snapshotApi.createSnapshot(
        {
          name: newSnapshotName,
          imageName: newImageName,
          entrypoint: newEntrypoint.trim() ? newEntrypoint.trim().split(' ') : undefined,
          cpu,
          memory,
          disk,
        },
        selectedOrganization?.id,
      )
      setShowCreateDialog(false)
      setNewSnapshotName('')
      setNewImageName('')
      setNewEntrypoint('')
      setCpu(undefined)
      setMemory(undefined)
      setDisk(undefined)
      setTargetPropagations([])
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

  const addTargetPropagation = () => {
    if (!newTarget.trim()) return

    if (targetPropagations.some((tp) => tp.target === newTarget.trim())) {
      toast.warning('Target already added')
      return
    }

    setTargetPropagations([...targetPropagations, { target: newTarget.trim(), userOverride: 1 }])
    setNewTarget('')
  }

  const removeTargetPropagation = (target: string) => {
    setTargetPropagations(targetPropagations.filter((tp) => tp.target !== target))
  }

  const updateTargetConcurrency = (target: string, value: number) => {
    setTargetPropagations(targetPropagations.map((tp) => (tp.target === target ? { ...tp, userOverride: value } : tp)))
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

  const handleActivate = async (snapshot: SnapshotDto) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))

    // Optimistically update the snapshot state
    setSnapshotsData((prev) => ({
      ...prev,
      items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: SnapshotState.ACTIVE } : i)),
    }))

    try {
      await snapshotApi.activateSnapshot(snapshot.id, selectedOrganization?.id)
      toast.success(`Activating snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to activate snapshot')
      // Revert the optimistic update
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: snapshot.state } : i)),
      }))
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const handleBulkDelete = async (snapshots: SnapshotDto[]) => {
    setLoadingSnapshots((prev) => ({ ...prev, ...snapshots.reduce((acc, img) => ({ ...acc, [img.id]: true }), {}) }))

    for (const snapshot of snapshots) {
      setSnapshotsData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: SnapshotState.REMOVING } : i)),
      }))

      try {
        await snapshotApi.removeSnapshot(snapshot.id, selectedOrganization?.id)
        toast.success(`Deleting snapshot ${snapshot.name}`)
      } catch (error) {
        handleApiError(error, `Failed to delete snapshot ${snapshot.name}`)

        setSnapshotsData((prev) => ({
          ...prev,
          items: prev.items.map((i) => (i.id === snapshot.id ? { ...i, state: snapshot.state } : i)),
        }))

        if (snapshots.indexOf(snapshot) < snapshots.length - 1) {
          const shouldContinue = window.confirm(
            `Failed to delete snapshot ${snapshot.name}. Do you want to continue with the remaining snapshots?`,
          )

          if (!shouldContinue) {
            break
          }
        }
      } finally {
        setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
      }
    }
  }

  const handleEditTargets = (snapshot: SnapshotDto) => {
    setSnapshotToEditTargets(snapshot)
    setEditingTargetPropagations(snapshot.targetPropagations || [])
    setShowTargetsDialog(true)
  }

  const addEditingTargetPropagation = () => {
    if (!editingNewTarget.trim()) return

    if (editingTargetPropagations.some((tp) => tp.target === editingNewTarget.trim())) {
      toast.warning('Target already added')
      return
    }

    setEditingTargetPropagations([...editingTargetPropagations, { target: editingNewTarget.trim(), userOverride: 1 }])
    setEditingNewTarget('')
  }

  const removeEditingTargetPropagation = (target: string) => {
    setEditingTargetPropagations(editingTargetPropagations.filter((tp) => tp.target !== target))
  }

  const updateEditingTargetConcurrency = (target: string, value: number) => {
    setEditingTargetPropagations(
      editingTargetPropagations.map((tp) => (tp.target === target ? { ...tp, userOverride: value } : tp)),
    )
  }

  const handleUpdateTargets = async () => {
    if (!snapshotToEditTargets) return

    setLoadingTargetUpdate(true)
    try {
      const response = await snapshotApi.setSnapshotTargetPropagations(
        snapshotToEditTargets.id,
        {
          targetPropagations: editingTargetPropagations,
        },
        selectedOrganization?.id,
      )
      setShowTargetsDialog(false)
      toast.success(`Updated targets for snapshot ${snapshotToEditTargets.name}`)

      // Update the snapshot in the local state using the actual response data
      setSnapshotsData((prev) => {
        const updatedItems = prev.items.map((i) => {
          if (i.id === snapshotToEditTargets.id) {
            // Use the actual response data instead of manually constructing it
            return response.data
          }
          return i
        })

        return {
          ...prev,
          items: updatedItems,
        }
      })
    } catch (error) {
      handleApiError(error, 'Failed to update snapshot targets')
    } finally {
      setLoadingTargetUpdate(false)
    }
  }

  return (
    <div className="px-6 py-2">
      <Dialog
        open={showCreateDialog}
        onOpenChange={(isOpen) => {
          setShowCreateDialog(isOpen)
          if (isOpen) {
            return
          }
          setNewSnapshotName('')
          setNewImageName('')
          setNewEntrypoint('')
          setCpu(undefined)
          setMemory(undefined)
          setDisk(undefined)
          setTargetPropagations([])
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Snapshots</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button
                variant="default"
                size="sm"
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
                  placeholder="ubuntu-4vcpu-8ram-100gb"
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  The name you will use in your client app (SDK, CLI) to reference the snapshot.
                </p>
              </div>
              <div className="space-y-3">
                <Label htmlFor="name">Image</Label>
                <Input
                  id="name"
                  value={newImageName}
                  onChange={(e) => setNewImageName(e.target.value)}
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
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Resources</h3>
                <div className="space-y-4 px-4 py-2">
                  <div className="flex items-center gap-4">
                    <Label htmlFor="cpu" className="w-32 flex-shrink-0">
                      Compute (vCPU):
                    </Label>
                    <Input
                      id="cpu"
                      type="number"
                      className="w-full"
                      min="1"
                      placeholder="1"
                      onChange={(e) => setCpu(parseInt(e.target.value) || undefined)}
                    />
                  </div>
                  <div className="flex items-center gap-4">
                    <Label htmlFor="memory" className="w-32 flex-shrink-0">
                      Memory (GiB):
                    </Label>
                    <Input
                      id="memory"
                      type="number"
                      className="w-full"
                      min="1"
                      placeholder="1"
                      onChange={(e) => setMemory(parseInt(e.target.value) || undefined)}
                    />
                  </div>
                  <div className="flex items-center gap-4">
                    <Label htmlFor="disk" className="w-32 flex-shrink-0">
                      Storage (GiB):
                    </Label>
                    <Input
                      id="disk"
                      type="number"
                      className="w-full"
                      min="1"
                      placeholder="3"
                      onChange={(e) => setDisk(parseInt(e.target.value) || undefined)}
                    />
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  If not specified, default values will be used (1 vCPU, 1 GiB memory, 3 GiB storage).
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
                  disabled={
                    !newSnapshotName.trim() ||
                    !newImageName.trim() ||
                    validateSnapshotName(newSnapshotName) !== null ||
                    validateImageName(newImageName) !== null
                  }
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
          onActivate={handleActivate}
          onEditTargets={handleEditTargets}
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

      {snapshotToEditTargets && (
        <Dialog
          open={showTargetsDialog}
          onOpenChange={(isOpen) => {
            setShowTargetsDialog(isOpen)
            if (!isOpen) {
              setSnapshotToEditTargets(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Targets for Snapshot - {snapshotToEditTargets.name} </DialogTitle>
              <DialogDescription>Update the distribution per target settings for the snapshot.</DialogDescription>
            </DialogHeader>
            <form
              id="edit-targets-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleUpdateTargets()
              }}
            >
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Distribution per Target</h3>
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Specify targets and the desired number of concurrent sandboxes for each target.
                </p>
                <div className="flex items-center gap-2">
                  <Input
                    value={editingNewTarget}
                    onChange={(e) => setEditingNewTarget(e.target.value)}
                    placeholder="Target environment (e.g., eu, us, asia)"
                    className="flex-grow"
                  />
                  <Button
                    type="button"
                    onClick={addEditingTargetPropagation}
                    disabled={!editingNewTarget.trim()}
                    variant="outline"
                  >
                    Add
                  </Button>
                </div>

                {editingTargetPropagations.length > 0 && (
                  <div className="space-y-3 mt-2">
                    {editingTargetPropagations.map((tp) => (
                      <div key={tp.target} className="border rounded-md p-3 space-y-2">
                        <div className="flex justify-between items-center">
                          <span className="font-medium">{tp.target}</span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => removeEditingTargetPropagation(tp.target)}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        </div>
                        <div className="grid grid-cols-[auto_1fr] gap-4 items-center">
                          <span className="text-sm text-muted-foreground whitespace-nowrap">
                            Desired Concurrent Sandboxes:
                          </span>
                          <div className="flex flex-row">
                            <Slider
                              value={[tp.userOverride]}
                              min={1}
                              max={snapshotToEditTargets?.maximumUserOverride}
                              step={1}
                              onValueChange={(value) => updateEditingTargetConcurrency(tp.target, value[0])}
                            />
                            <span className="text-right font-medium">{tp.userOverride}</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </form>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              {loadingTargetUpdate ? (
                <Button type="button" variant="default" disabled>
                  Updating...
                </Button>
              ) : (
                <Button type="submit" form="edit-targets-form" variant="default">
                  Update
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Snapshots
