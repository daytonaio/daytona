/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { queryKeys } from '@/hooks/queries/queryKeys'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import type { Runner, RunnerState } from '@daytona/api-client'
import type { QueryKey } from '@tanstack/react-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

const getRunnerListRegionId = (queryKey: QueryKey) => {
  const filters = queryKey[3]

  if (!filters || typeof filters !== 'object' || !('regionId' in filters)) {
    return undefined
  }

  return typeof filters.regionId === 'string' ? filters.regionId : undefined
}

export function useRunnerWsSync() {
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!notificationSocket || !selectedOrganization?.id) return

    const queryKey = queryKeys.runners.list(selectedOrganization.id)

    const upsertRunnerInCache = (runner: Runner) => {
      queryClient.setQueriesData<Runner[]>(
        {
          queryKey,
          predicate: (query) => {
            const regionId = getRunnerListRegionId(query.queryKey)
            return !regionId || regionId === runner.region
          },
        },
        (previousRunners) => {
          if (!previousRunners) return [runner]

          if (!previousRunners.some((existingRunner) => existingRunner.id === runner.id)) {
            return [runner, ...previousRunners]
          }

          return previousRunners.map((existingRunner) => (existingRunner.id === runner.id ? runner : existingRunner))
        },
      )
    }

    const updateRunnerStateInCache = (runner: Runner, state: RunnerState) => {
      queryClient.setQueriesData<Runner[]>(
        {
          queryKey,
          predicate: (query) => {
            const regionId = getRunnerListRegionId(query.queryKey)
            return !regionId || regionId === runner.region
          },
        },
        (previousRunners) => {
          const updatedRunner = { ...runner, state }

          if (!previousRunners) return [updatedRunner]

          if (!previousRunners.some((existingRunner) => existingRunner.id === runner.id)) {
            return [updatedRunner, ...previousRunners]
          }

          return previousRunners.map((existingRunner) =>
            existingRunner.id === runner.id
              ? {
                  ...existingRunner,
                  state,
                }
              : existingRunner,
          )
        },
      )
    }

    const invalidate = (refetchType: 'active' | 'none' = 'none') => {
      queryClient.invalidateQueries({
        queryKey,
        refetchType,
      })
    }

    const handleRunnerCreatedEvent = (runner: Runner) => {
      upsertRunnerInCache(runner)
      invalidate('active')
    }

    const handleRunnerStateUpdatedEvent = (data: { runner: Runner; oldState: RunnerState; newState: RunnerState }) => {
      updateRunnerStateInCache(data.runner, data.newState)
      invalidate()
    }

    const handleRunnerUnschedulableUpdatedEvent = (runner: Runner) => {
      upsertRunnerInCache(runner)
      invalidate()
    }

    notificationSocket.on('runner.created', handleRunnerCreatedEvent)
    notificationSocket.on('runner.state.updated', handleRunnerStateUpdatedEvent)
    notificationSocket.on('runner.unschedulable.updated', handleRunnerUnschedulableUpdatedEvent)

    return () => {
      notificationSocket.off('runner.created', handleRunnerCreatedEvent)
      notificationSocket.off('runner.state.updated', handleRunnerStateUpdatedEvent)
      notificationSocket.off('runner.unschedulable.updated', handleRunnerUnschedulableUpdatedEvent)
    }
  }, [notificationSocket, queryClient, selectedOrganization?.id])
}
