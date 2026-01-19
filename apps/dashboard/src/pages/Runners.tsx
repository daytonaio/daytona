/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import {
  Runner,
  RunnerState,
  OrganizationRolePermissionsEnum,
  CreateRunner,
  CreateRunnerResponse,
} from '@daytonaio/api-client'
import { RunnerTable } from '@/components/RunnerTable'
import { CreateRunnerDialog } from '@/components/CreateRunnerDialog'
import RunnerDetailsSheet from '@/components/RunnerDetailsSheet'
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
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { handleApiError } from '@/lib/error-handling'
import { useRegions } from '@/hooks/useRegions'

const Runners: React.FC = () => {
  const { runnersApi } = useApi()
  const { notificationSocket } = useNotificationSocket()
  const { customRegions: regions, loadingAvailableRegions: loadingRegions, getRegionName } = useRegions()

  const [runners, setRunners] = useState<Runner[]>([])
  const [loadingRunnersData, setLoadingRunnersData] = useState(false)
  const [runnerIsLoading, setRunnerIsLoading] = useState<Record<string, boolean>>({})

  const [runnerToDelete, setRunnerToDelete] = useState<Runner | null>(null)
  const [deleteRunnerDialogIsOpen, setDeleteRunnerDialogIsOpen] = useState(false)

  const [runnerToToggleScheduling, setRunnerToToggleScheduling] = useState<Runner | null>(null)
  const [toggleRunnerSchedulingDialogIsOpen, setToggleRunnerSchedulingDialogIsOpen] = useState(false)

  const [selectedRunner, setSelectedRunner] = useState<Runner | null>(null)
  const [showRunnerDetails, setShowRunnerDetails] = useState(false)

  const [autoRefresh, setAutoRefresh] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchRunners = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingRunnersData(true)
      }
      try {
        const response = (await runnersApi.listRunners(selectedOrganization.id)).data
        setRunners(response || [])
      } catch (error) {
        handleApiError(error, 'Failed to fetch runners')
        setRunners([])
      } finally {
        setLoadingRunnersData(false)
      }
    },
    [runnersApi, selectedOrganization],
  )

  useEffect(() => {
    fetchRunners()
  }, [fetchRunners])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(() => {
      fetchRunners(false)
    }, 5000)
    return () => clearInterval(interval)
  }, [autoRefresh, fetchRunners])

  useEffect(() => {
    const handleRunnerCreatedEvent = (runner: Runner) => {
      if (!runners.some((r) => r.id === runner.id)) {
        setRunners((prev) => [runner, ...prev])
      }
    }

    const handleRunnerStateUpdatedEvent = (data: { runner: Runner; oldState: RunnerState; newState: RunnerState }) => {
      if (!runners.some((r) => r.id === data.runner.id)) {
        setRunners((prev) => [data.runner, ...prev])
      } else {
        setRunners((prev) =>
          prev.map((r) =>
            r.id === data.runner.id
              ? {
                  ...r,
                  state: data.newState,
                }
              : r,
          ),
        )
      }
    }

    const handleRunnerUnschedulableUpdatedEvent = (runner: Runner) => {
      if (!runners.some((r) => r.id === runner.id)) {
        setRunners((prev) => [runner, ...prev])
      } else {
        setRunners((prev) => prev.map((r) => (r.id === runner.id ? runner : r)))
      }
    }

    if (!notificationSocket) {
      return
    }

    notificationSocket.on('runner.created', handleRunnerCreatedEvent)
    notificationSocket.on('runner.state.updated', handleRunnerStateUpdatedEvent)
    notificationSocket.on('runner.unschedulable.updated', handleRunnerUnschedulableUpdatedEvent)

    return () => {
      notificationSocket.off('runner.created', handleRunnerCreatedEvent)
      notificationSocket.off('runner.state.updated', handleRunnerStateUpdatedEvent)
      notificationSocket.off('runner.unschedulable.updated', handleRunnerUnschedulableUpdatedEvent)
    }
  }, [notificationSocket, runners])

  useEffect(() => {
    if (!selectedRunner || !runners) return
    const found = runners.find((r) => r.id === selectedRunner.id)
    if (!found) {
      setSelectedRunner(null)
      setShowRunnerDetails(false)
      return
    }
    if (found !== selectedRunner) setSelectedRunner(found)
  }, [runners, selectedRunner])

  const handleCreateRunner = async (createRunnerData: CreateRunner): Promise<CreateRunnerResponse | null> => {
    try {
      const response = (await runnersApi.createRunner(createRunnerData, selectedOrganization?.id)).data
      toast.success('Runner created successfully')
      return response
    } catch (error) {
      handleApiError(error, 'Failed to create runner')
      return null
    }
  }

  const handleToggleEnabled = async (runner: Runner) => {
    setRunnerToToggleScheduling(runner)
    setToggleRunnerSchedulingDialogIsOpen(true)
  }

  const confirmToggleScheduling = async () => {
    if (!runnerToToggleScheduling) return

    setRunnerIsLoading((prev) => ({ ...prev, [runnerToToggleScheduling.id]: true }))
    try {
      await runnersApi.updateRunnerScheduling(runnerToToggleScheduling.id, selectedOrganization?.id, {
        data: { unschedulable: !runnerToToggleScheduling.unschedulable },
      })
      toast.success(
        `Runner is now ${runnerToToggleScheduling.unschedulable ? 'available' : 'unavailable'} for scheduling new sandboxes`,
      )
    } catch (error) {
      handleApiError(error, 'Failed to update runner scheduling status')
    } finally {
      setRunnerIsLoading((prev) => ({ ...prev, [runnerToToggleScheduling.id]: false }))
      setToggleRunnerSchedulingDialogIsOpen(false)
      setRunnerToToggleScheduling(null)
    }
  }

  const handleDelete = async (runner: Runner) => {
    setRunnerToDelete(runner)
    setDeleteRunnerDialogIsOpen(true)
  }

  const confirmDelete = async () => {
    if (!runnerToDelete) return

    setRunnerIsLoading((prev) => ({ ...prev, [runnerToDelete.id]: true }))
    try {
      await runnersApi.deleteRunner(runnerToDelete.id, selectedOrganization?.id)
      toast.success('Runner deleted successfully')
      await fetchRunners(false)
    } catch (error) {
      handleApiError(error, 'Failed to delete runner')
    } finally {
      setRunnerIsLoading((prev) => ({ ...prev, [runnerToDelete.id]: false }))
      setDeleteRunnerDialogIsOpen(false)
      setRunnerToDelete(null)
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_RUNNERS),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_RUNNERS),
    [authenticatedUserHasPermission],
  )

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Runners</h1>
        {writePermitted && regions.length > 0 && (
          <CreateRunnerDialog regions={regions} onCreateRunner={handleCreateRunner} />
        )}
      </div>

      <RunnerTable
        data={runners}
        regions={regions}
        loading={loadingRunnersData || loadingRegions}
        isLoadingRunner={(runner) => runnerIsLoading[runner.id] || false}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        onToggleEnabled={handleToggleEnabled}
        onDelete={handleDelete}
        getRegionName={getRegionName}
        onRowClick={(runner: Runner) => {
          setSelectedRunner(runner)
          setShowRunnerDetails(true)
        }}
        autoRefresh={autoRefresh}
        onAutoRefreshChange={setAutoRefresh}
      />

      {runnerToToggleScheduling && (
        <Dialog open={toggleRunnerSchedulingDialogIsOpen} onOpenChange={setToggleRunnerSchedulingDialogIsOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Update Runner</DialogTitle>
              <DialogDescription>
                Are you sure you want to update the scheduling status of this runner? This will make the runner{' '}
                {runnerToToggleScheduling.unschedulable ? 'available' : 'unavailable'} for scheduling new sandboxes.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              <Button
                variant={runnerToToggleScheduling.unschedulable ? 'default' : 'destructive'}
                onClick={confirmToggleScheduling}
                disabled={runnerIsLoading[runnerToToggleScheduling.id]}
              >
                {runnerIsLoading[runnerToToggleScheduling.id]
                  ? 'Updating...'
                  : runnerToToggleScheduling.unschedulable
                    ? 'Mark as schedulable'
                    : 'Mark as unschedulable'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {runnerToDelete && (
        <Dialog open={deleteRunnerDialogIsOpen} onOpenChange={setDeleteRunnerDialogIsOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Runner Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this runner? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              <Button variant="destructive" onClick={confirmDelete} disabled={runnerIsLoading[runnerToDelete.id]}>
                {runnerIsLoading[runnerToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      <RunnerDetailsSheet
        runner={selectedRunner}
        open={showRunnerDetails}
        onOpenChange={setShowRunnerDetails}
        runnerIsLoading={runnerIsLoading}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        onDelete={(runner) => {
          setRunnerToDelete(runner)
          setDeleteRunnerDialogIsOpen(true)
          setShowRunnerDetails(false)
        }}
        getRegionName={getRegionName}
      />
    </div>
  )
}

export default Runners
