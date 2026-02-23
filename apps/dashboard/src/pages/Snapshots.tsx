/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { CreateSnapshotDialog } from '@/components/snapshots/CreateSnapshotDialog'
import { SnapshotTable } from '@/components/snapshots/SnapshotTable'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { queryKeys } from '@/hooks/queries/queryKeys'
import {
  DEFAULT_SNAPSHOT_SORTING,
  SnapshotQueryParams,
  SnapshotSorting,
  useSnapshotsQuery,
} from '@/hooks/queries/useSnapshotsQuery'
import { useApi } from '@/hooks/useApi'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { pluralize } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, PaginatedSnapshots, SnapshotDto, SnapshotState } from '@daytonaio/api-client'
import { useQueryClient } from '@tanstack/react-query'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

const Snapshots: React.FC = () => {
  const { notificationSocket } = useNotificationSocket()
  const queryClient = useQueryClient()

  const { snapshotApi } = useApi()
  const { getRegionName } = useRegions()
  const [loadingSnapshots, setLoadingSnapshots] = useState<Record<string, boolean>>({})
  const [snapshotToDelete, setSnapshotToDelete] = useState<SnapshotDto | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

  const [sorting, setSorting] = useState<SnapshotSorting>(DEFAULT_SNAPSHOT_SORTING)

  const queryParams = useMemo<SnapshotQueryParams>(
    () => ({
      page: paginationParams.pageIndex + 1,
      pageSize: paginationParams.pageSize,
      sorting,
      filters: searchQuery.trim() ? { name: searchQuery.trim() } : undefined,
    }),
    [paginationParams, sorting, searchQuery],
  )

  const baseQueryKey = useMemo(() => queryKeys.snapshots.all, [])

  const queryKey = useMemo(
    () => queryKeys.snapshots.list(selectedOrganization?.id ?? '', queryParams),
    [selectedOrganization?.id, queryParams],
  )

  const {
    data: snapshotsData,
    isLoading: snapshotsDataIsLoading,
    error: snapshotsDataError,
  } = useSnapshotsQuery(queryParams)

  useEffect(() => {
    if (snapshotsDataError) {
      handleApiError(snapshotsDataError, 'Failed to fetch snapshots')
    }
  }, [snapshotsDataError])

  const prevSearchQueryRef = useRef<string>(searchQuery)

  // Reset to first page when search query changes
  useEffect(() => {
    if (searchQuery !== prevSearchQueryRef.current) {
      prevSearchQueryRef.current = searchQuery
      setPaginationParams((prev) => {
        if (prev.pageIndex > 0) {
          return { ...prev, pageIndex: 0 }
        }
        return prev
      })
    }
  }, [searchQuery])

  const updateSnapshotInCache = useCallback(
    (snapshotId: string, updates: Partial<SnapshotDto>) => {
      queryClient.setQueryData(queryKey, (oldData: PaginatedSnapshots | undefined) => {
        if (!oldData) return oldData
        return {
          ...oldData,
          items: oldData.items.map((snapshot) => (snapshot.id === snapshotId ? { ...snapshot, ...updates } : snapshot)),
        }
      })
    },
    [queryClient, queryKey],
  )

  const markAllSnapshotQueriesAsStale = useCallback(
    async (shouldRefetchActiveQueries = false) => {
      queryClient.invalidateQueries({
        queryKey: baseQueryKey,
        refetchType: shouldRefetchActiveQueries ? 'active' : 'none',
      })
    },
    [queryClient, baseQueryKey],
  )

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  const handleSortingChange = useCallback((newSorting: SnapshotSorting) => {
    setSorting(newSorting)
    setPaginationParams((prev) => ({ ...prev, pageIndex: 0 }))
  }, [])

  useEffect(() => {
    const handleSnapshotCreatedEvent = () => {
      markAllSnapshotQueriesAsStale(true)
    }

    const handleSnapshotStateUpdatedEvent = (data: {
      snapshot: SnapshotDto
      oldState: SnapshotState
      newState: SnapshotState
    }) => {
      updateSnapshotInCache(data.snapshot.id, data.snapshot)
    }

    const handleSnapshotRemovedEvent = () => {
      markAllSnapshotQueriesAsStale(true)
    }

    if (!notificationSocket) {
      return
    }

    notificationSocket.on('snapshot.created', handleSnapshotCreatedEvent)
    notificationSocket.on('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
    notificationSocket.on('snapshot.removed', handleSnapshotRemovedEvent)

    return () => {
      notificationSocket.off('snapshot.created', handleSnapshotCreatedEvent)
      notificationSocket.off('snapshot.state.updated', handleSnapshotStateUpdatedEvent)
      notificationSocket.off('snapshot.removed', handleSnapshotRemovedEvent)
    }
  }, [notificationSocket, markAllSnapshotQueriesAsStale, updateSnapshotInCache])

  useEffect(() => {
    if (snapshotsData?.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [snapshotsData?.items.length, paginationParams.pageIndex])

  const handleDelete = async (snapshot: SnapshotDto) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))
    updateSnapshotInCache(snapshot.id, { state: SnapshotState.REMOVING })

    try {
      await snapshotApi.removeSnapshot(snapshot.id, selectedOrganization?.id)
      setSnapshotToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete snapshot')
      updateSnapshotInCache(snapshot.id, { state: snapshot.state })
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const handleActivate = async (snapshot: SnapshotDto) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))
    updateSnapshotInCache(snapshot.id, { state: SnapshotState.PENDING })

    try {
      await snapshotApi.activateSnapshot(snapshot.id, selectedOrganization?.id)
      toast.success(`Activating snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to activate snapshot')
      updateSnapshotInCache(snapshot.id, { state: snapshot.state })
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const handleDeactivate = async (snapshot: SnapshotDto) => {
    setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: true }))
    updateSnapshotInCache(snapshot.id, { state: SnapshotState.INACTIVE })

    try {
      await snapshotApi.deactivateSnapshot(snapshot.id, selectedOrganization?.id)
      toast.success(`Deactivating snapshot ${snapshot.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to deactivate snapshot')
      updateSnapshotInCache(snapshot.id, { state: snapshot.state })
    } finally {
      setLoadingSnapshots((prev) => ({ ...prev, [snapshot.id]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const executeBulkAction = useCallback(
    async ({
      ids,
      actionName,
      optimisticState,
      apiCall,
      toastMessages,
    }: {
      ids: string[]
      actionName: string
      optimisticState: SnapshotState
      apiCall: (id: string) => Promise<unknown>
      toastMessages: {
        successTitle: string
        errorTitle: string
        warningTitle: string
        canceledTitle: string
      }
    }) => {
      const previousStatesById = new Map((snapshotsData?.items ?? []).map((snapshot) => [snapshot.id, snapshot.state]))

      let isCancelled = false
      let processedCount = 0
      let successCount = 0
      let failureCount = 0

      const totalLabel = pluralize(ids.length, 'snapshot', 'snapshots')
      const onCancel = () => {
        isCancelled = true
      }

      const bulkToast = createBulkActionToast(`${actionName} 0 of ${totalLabel}.`, {
        action: { label: 'Cancel', onClick: onCancel },
      })

      try {
        for (const id of ids) {
          if (isCancelled) break

          processedCount += 1
          bulkToast.loading(`${actionName} ${processedCount} of ${totalLabel}.`, {
            action: { label: 'Cancel', onClick: onCancel },
          })

          setLoadingSnapshots((prev) => ({ ...prev, [id]: true }))
          updateSnapshotInCache(id, { state: optimisticState })

          try {
            await apiCall(id)
            successCount += 1
          } catch (error) {
            failureCount += 1
            updateSnapshotInCache(id, { state: previousStatesById.get(id) })
            console.error(`${actionName} snapshot failed`, id, error)
          } finally {
            setLoadingSnapshots((prev) => ({ ...prev, [id]: false }))
          }
        }

        await markAllSnapshotQueriesAsStale(true)
        bulkToast.result({ successCount, failureCount }, toastMessages)
      } catch (error) {
        console.error(`${actionName} snapshots failed`, error)
        bulkToast.error(`${actionName} snapshots failed.`)
      }

      return { successCount, failureCount }
    },
    [snapshotsData?.items, updateSnapshotInCache, markAllSnapshotQueriesAsStale],
  )

  const handleBulkDelete = (snapshots: SnapshotDto[]) =>
    executeBulkAction({
      ids: snapshots.map((s) => s.id),
      actionName: 'Deleting',
      optimisticState: SnapshotState.REMOVING,
      apiCall: (id) => snapshotApi.removeSnapshot(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(snapshots.length, 'Snapshot', 'Snapshots')} deleted.`,
        errorTitle: `Failed to delete ${pluralize(snapshots.length, 'snapshot', 'snapshots')}.`,
        warningTitle: 'Failed to delete some snapshots.',
        canceledTitle: 'Delete canceled.',
      },
    })

  const handleBulkDeactivate = (snapshots: SnapshotDto[]) =>
    executeBulkAction({
      ids: snapshots.map((s) => s.id),
      actionName: 'Deactivating',
      optimisticState: SnapshotState.INACTIVE,
      apiCall: (id) => snapshotApi.deactivateSnapshot(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(snapshots.length, 'Snapshot', 'Snapshots')} deactivated.`,
        errorTitle: `Failed to deactivate ${pluralize(snapshots.length, 'snapshot', 'snapshots')}.`,
        warningTitle: 'Failed to deactivate some snapshots.',
        canceledTitle: 'Deactivate canceled.',
      },
    })

  const handleBulkActivate = (snapshots: SnapshotDto[]) =>
    executeBulkAction({
      ids: snapshots.map((s) => s.id),
      actionName: 'Activating',
      optimisticState: SnapshotState.ACTIVE,
      apiCall: (id) => snapshotApi.activateSnapshot(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(snapshots.length, 'Snapshot', 'Snapshots')} activated.`,
        errorTitle: `Failed to activate ${pluralize(snapshots.length, 'snapshot', 'snapshots')}.`,
        warningTitle: 'Failed to activate some snapshots.',
        canceledTitle: 'Activate canceled.',
      },
    })

  const dialogRef = useRef<{ open: () => void }>(null)

  const handleCreateSnapshot = () => {
    dialogRef.current?.open()
  }

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Snapshots</PageTitle>
        {writePermitted && <CreateSnapshotDialog className="ml-auto" ref={dialogRef} />}
      </PageHeader>

      <PageContent size="full">
        <SearchInput
          placeholder="Search snapshots..."
          value={searchQuery}
          onChange={setSearchQuery}
          resultCount={snapshotsData?.total ?? 0}
          entityName="snapshot"
          data-testid="snapshots-search"
        />

        <SnapshotTable
          data={snapshotsData?.items ?? []}
          loading={snapshotsDataIsLoading}
          loadingSnapshots={loadingSnapshots}
          getRegionName={getRegionName}
          onDelete={(snapshot) => {
            setSnapshotToDelete(snapshot)
            setShowDeleteDialog(true)
          }}
          onBulkDelete={handleBulkDelete}
          onBulkDeactivate={handleBulkDeactivate}
          onBulkActivate={handleBulkActivate}
          onActivate={handleActivate}
          onDeactivate={handleDeactivate}
          onCreateSnapshot={handleCreateSnapshot}
          pageCount={snapshotsData?.totalPages ?? 0}
          totalItems={snapshotsData?.total ?? 0}
          onPaginationChange={handlePaginationChange}
          pagination={{
            pageIndex: paginationParams.pageIndex,
            pageSize: paginationParams.pageSize,
          }}
          sorting={sorting}
          onSortingChange={handleSortingChange}
        />

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
      </PageContent>
    </PageLayout>
  )
}

export default Snapshots
