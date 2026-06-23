/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { CreateRunnerSheet } from '@/components/CreateRunnerSheet'
import { PageContent, PageFooter, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { RefreshIntervalValue } from '@/components/RefreshSegmentedButton'
import RunnerDetailsSheet, { type RunnerDetailsSheetRef } from '@/components/RunnerDetailsSheet'
import { RunnerTable } from '@/components/RunnerTable'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useDeleteRunnerMutation } from '@/hooks/mutations/useDeleteRunnerMutation'
import { useMutatingRunners } from '@/hooks/mutations/useMutatingRunners'
import { useUpdateRunnerSchedulingMutation } from '@/hooks/mutations/useUpdateRunnerSchedulingMutation'
import { useAvailableRegionsQuery, useRegionLookup } from '@/hooks/queries/useRegionsQuery'
import { useRunnersQuery } from '@/hooks/queries/useRunnersQuery'
import { useRunnerWsSync } from '@/hooks/useRunnerWsSync'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { EMPTY_REGIONS, filterCustomRegions } from '@/lib/regions'
import { OrganizationRolePermissionsEnum, Runner } from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

const Runners: React.FC = () => {
  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const [refreshInterval, setRefreshInterval] = useState<RefreshIntervalValue>(false)

  useRunnerWsSync()

  const { data: availableRegions = EMPTY_REGIONS, isLoading: loadingRegions } = useAvailableRegionsQuery(
    selectedOrganization?.id,
  )
  const regions = useMemo(() => filterCustomRegions(availableRegions), [availableRegions])
  const { getRegionName } = useRegionLookup(selectedOrganization?.id)
  const {
    data: runners = [],
    dataUpdatedAt: runnersUpdatedAt,
    error: runnersDataError,
    isFetching: runnersDataIsFetching,
    isLoading: loadingRunnersData,
    refetch: refetchRunnersData,
  } = useRunnersQuery({
    refetchInterval: refreshInterval,
  })
  const updateRunnerSchedulingMutation = useUpdateRunnerSchedulingMutation()
  const deleteRunnerMutation = useDeleteRunnerMutation()
  const mutatingRunnerIds = useMutatingRunners()
  const runnerIsLoading = useMemo<Record<string, boolean>>(() => {
    return Object.fromEntries([...mutatingRunnerIds].map((runnerId) => [runnerId, true]))
  }, [mutatingRunnerIds])

  const [runnerToDelete, setRunnerToDelete] = useState<Runner | null>(null)
  const [deleteRunnerDialogIsOpen, setDeleteRunnerDialogIsOpen] = useState(false)

  const [runnerToToggleScheduling, setRunnerToToggleScheduling] = useState<Runner | null>(null)
  const [toggleRunnerSchedulingDialogIsOpen, setToggleRunnerSchedulingDialogIsOpen] = useState(false)

  const [selectedRunner, setSelectedRunner] = useState<Runner | null>(null)
  const [showRunnerDetails, setShowRunnerDetails] = useState(false)

  const createRunnerSheetRef = useRef<{ open: () => void }>(null)
  const runnerDetailsSheetRef = useRef<RunnerDetailsSheetRef>(null)

  useEffect(() => {
    if (runnersDataError) {
      handleApiError(runnersDataError, 'Failed to fetch runners')
    }
  }, [runnersDataError])

  useEffect(() => {
    if (!selectedRunner || !runners) return
    const found = runners.find((r) => r.id === selectedRunner.id)
    if (!found) {
      runnerDetailsSheetRef.current?.close()
      return
    }
    setSelectedRunner(found)
  }, [runners, selectedRunner])

  const selectedRunnerIndex = useMemo(() => {
    if (!selectedRunner) {
      return -1
    }

    return runners.findIndex((runner) => runner.id === selectedRunner.id)
  }, [runners, selectedRunner])

  const handleNavigateRunner = useCallback(
    (direction: 'prev' | 'next') => {
      if (selectedRunnerIndex === -1) {
        return
      }

      const nextIndex = direction === 'prev' ? selectedRunnerIndex - 1 : selectedRunnerIndex + 1
      const nextRunner = runners[nextIndex]
      if (nextRunner) {
        setSelectedRunner(nextRunner)
      }
    },
    [runners, selectedRunnerIndex],
  )

  const handleToggleEnabled = (runner: Runner) => {
    setRunnerToToggleScheduling(runner)
    setToggleRunnerSchedulingDialogIsOpen(true)
  }

  const confirmToggleScheduling = async () => {
    if (!runnerToToggleScheduling) return

    try {
      await updateRunnerSchedulingMutation.mutateAsync({
        runnerId: runnerToToggleScheduling.id,
        organizationId: selectedOrganization?.id,
        unschedulable: !runnerToToggleScheduling.unschedulable,
      })
      toast.success(
        `Runner is now ${runnerToToggleScheduling.unschedulable ? 'available' : 'unavailable'} for scheduling new sandboxes`,
      )
    } catch (error) {
      handleApiError(error, 'Failed to update runner scheduling status')
    } finally {
      setToggleRunnerSchedulingDialogIsOpen(false)
      setRunnerToToggleScheduling(null)
    }
  }

  const handleDelete = (runner: Runner) => {
    setRunnerToDelete(runner)
    setDeleteRunnerDialogIsOpen(true)
  }

  const confirmDelete = async () => {
    if (!runnerToDelete) return

    try {
      await deleteRunnerMutation.mutateAsync({
        runnerId: runnerToDelete.id,
        organizationId: selectedOrganization?.id,
      })
      toast.success('Runner deleted successfully')
    } catch (error) {
      handleApiError(error, 'Failed to delete runner')
    } finally {
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

  return (
    <PageLayout contained>
      <PageHeader />

      <PageContent size="full" className="overflow-hidden">
        <PageIntro
          title="Runners"
          actions={
            writePermitted && regions.length > 0 ? (
              <CreateRunnerSheet regions={regions} ref={createRunnerSheetRef} />
            ) : undefined
          }
        />
        <RunnerTable
          data={runners}
          regions={regions}
          loading={loadingRunnersData || loadingRegions}
          activeRunnerId={showRunnerDetails ? selectedRunner?.id : undefined}
          isLoadingRunner={(runner) => runnerIsLoading[runner.id] || false}
          writePermitted={writePermitted}
          deletePermitted={deletePermitted}
          onToggleEnabled={handleToggleEnabled}
          onDelete={handleDelete}
          getRegionName={getRegionName}
          onRowClick={(runner: Runner) => {
            setSelectedRunner(runner)
            runnerDetailsSheetRef.current?.open()
          }}
          refreshInterval={refreshInterval}
          onRefreshIntervalChange={setRefreshInterval}
          onRefresh={refetchRunnersData}
          isRefreshing={runnersDataIsFetching}
          lastUpdatedAt={runnersUpdatedAt || undefined}
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
                {runnerIsLoading[runnerToToggleScheduling.id] && <Spinner />}
                {runnerToToggleScheduling.unschedulable ? 'Mark as schedulable' : 'Mark as unschedulable'}
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
                {runnerIsLoading[runnerToDelete.id] && <Spinner />}
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      <RunnerDetailsSheet
        ref={runnerDetailsSheetRef}
        runner={selectedRunner}
        onOpenChange={setShowRunnerDetails}
        runnerIsLoading={runnerIsLoading}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        hasPrev={selectedRunnerIndex > 0}
        hasNext={selectedRunnerIndex >= 0 && selectedRunnerIndex < runners.length - 1}
        onNavigate={handleNavigateRunner}
        onToggleEnabled={handleToggleEnabled}
        onDelete={(runner) => {
          setRunnerToDelete(runner)
          setDeleteRunnerDialogIsOpen(true)
        }}
        getRegionName={getRegionName}
      />
    </PageLayout>
  )
}

export default Runners
