/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { CreateRunnerSheet } from '@/components/CreateRunnerSheet'
import {
  PageBreadcrumbs,
  PageContent,
  PageDocsLink,
  PageFooter,
  PageHeader,
  PageIntro,
  PageLayout,
  PageStats,
} from '@/components/PageLayout'
import { RefreshIntervalValue } from '@/components/RefreshSegmentedButton'
import RunnerDetailsSheet from '@/components/RunnerDetailsSheet'
import { RunnerTable } from '@/components/RunnerTable'
import { Button } from '@/components/ui/button'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useApi } from '@/hooks/useApi'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import {
  CreateRunner,
  CreateRunnerResponse,
  OrganizationRolePermissionsEnum,
  Runner,
  RunnerState,
} from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

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

  const [refreshInterval, setRefreshInterval] = useState<RefreshIntervalValue>(false)
  const [runnersUpdatedAt, setRunnersUpdatedAt] = useState<number | undefined>()
  const createRunnerSheetRef = useRef<{ open: () => void }>(null)

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
        const response = (await runnersApi.listRunners(undefined, selectedOrganization.id)).data
        setRunners(response || [])
        setRunnersUpdatedAt(Date.now())
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
    if (typeof refreshInterval !== 'number') return
    const interval = setInterval(() => {
      fetchRunners(false)
    }, refreshInterval)
    return () => clearInterval(interval)
  }, [refreshInterval, fetchRunners])

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
    setSelectedRunner(found)
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

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!writePermitted || regions.length === 0) {
      return []
    }

    return [
      {
        id: 'create-runner',
        label: 'Create Runner',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => createRunnerSheetRef.current?.open(),
      },
    ]
  }, [writePermitted, regions.length])

  useRegisterCommands(rootCommands, { groupId: 'runner-actions', groupLabel: 'Runner actions', groupOrder: 0 })

  const runnerStats = useMemo(() => {
    const counts = runners.reduce((acc, runner) => {
      acc.set(runner.state, (acc.get(runner.state) ?? 0) + 1)
      return acc
    }, new Map<RunnerState, number>())

    const markerColors: Partial<Record<RunnerState, string>> = {
      [RunnerState.READY]: 'bg-success-foreground',
      [RunnerState.INITIALIZING]: 'bg-warning-foreground',
      [RunnerState.UNRESPONSIVE]: 'bg-destructive-foreground',
      [RunnerState.DISABLED]: 'bg-muted-foreground/50',
      [RunnerState.DECOMMISSIONED]: 'bg-muted-foreground/50',
    }

    return [
      { label: 'total', value: runners.length },
      ...Array.from(counts.entries()).map(([state, count]) => ({
        label: state
          .split('_')
          .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
          .join(' '),
        value: count,
        markerClassName: markerColors[state] ?? 'bg-muted-foreground/50',
      })),
    ]
  }, [runners])

  return (
    <PageLayout contained>
      <PageHeader>
        <PageBreadcrumbs current="Runners" />
        <PageDocsLink href={`${DAYTONA_DOCS_URL}/en/runners/`} label="Runner Docs" />
      </PageHeader>

      <PageContent size="full" className="overflow-hidden">
        <PageIntro
          title="Runners"
          description="Manage compute workers connected to your custom regions."
          titleActions={
            <PageStats
              items={runnerStats}
              loadingText={loadingRunnersData || loadingRegions ? 'Loading runners...' : undefined}
            />
          }
        />
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
          refreshInterval={refreshInterval}
          onRefreshIntervalChange={setRefreshInterval}
          onRefresh={() => fetchRunners(false)}
          isRefreshing={loadingRunnersData}
          lastUpdatedAt={runnersUpdatedAt}
          toolbarActions={
            writePermitted &&
            regions.length > 0 && (
              <CreateRunnerSheet regions={regions} onCreateRunner={handleCreateRunner} ref={createRunnerSheetRef} />
            )
          }
        />
      </PageContent>
      <PageFooter />

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
    </PageLayout>
  )
}

export default Runners
