/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxMutationVariables } from '@/hooks/mutations/mutationKeys'
import { useArchiveSandboxMutation } from '@/hooks/mutations/useArchiveSandboxMutation'
import { useDeleteSandboxMutation } from '@/hooks/mutations/useDeleteSandboxMutation'
import { useRecoverSandboxMutation } from '@/hooks/mutations/useRecoverSandboxMutation'
import { useStartSandboxMutation } from '@/hooks/mutations/useStartSandboxMutation'
import { useStopSandboxMutation } from '@/hooks/mutations/useStopSandboxMutation'
import { SandboxState } from '@daytona/api-client'
import { useCallback } from 'react'

interface UseSandboxListMutationsOptions {
  getSandboxState: (sandboxId: string) => SandboxState | undefined
  performSandboxStateOptimisticUpdate: (sandboxId: string, newState: SandboxState) => void
  revertSandboxStateOptimisticUpdate: (sandboxId: string, previousState?: SandboxState) => void
  cancelOutgoingRefetches: () => Promise<void>
  markAllQueriesAsStale: (shouldRefetchActive?: boolean) => Promise<void>
}

interface SandboxMutationContext {
  previousState?: SandboxState
}

export function useSandboxListMutations({
  getSandboxState,
  performSandboxStateOptimisticUpdate,
  revertSandboxStateOptimisticUpdate,
  cancelOutgoingRefetches,
  markAllQueriesAsStale,
}: UseSandboxListMutationsOptions) {
  const startMutation = useStartSandboxMutation()
  const recoverMutation = useRecoverSandboxMutation()
  const stopMutation = useStopSandboxMutation()
  const archiveMutation = useArchiveSandboxMutation()
  const deleteMutation = useDeleteSandboxMutation()

  const runSandboxMutation = useCallback(
    async ({
      sandboxId,
      optimisticState,
      mutateAsync,
    }: {
      sandboxId: string
      optimisticState: SandboxState
      mutateAsync: (variables: SandboxMutationVariables) => Promise<void>
    }) => {
      await cancelOutgoingRefetches()

      const context: SandboxMutationContext = {
        previousState: getSandboxState(sandboxId),
      }

      performSandboxStateOptimisticUpdate(sandboxId, optimisticState)

      try {
        await mutateAsync({ sandboxId })
      } catch (error) {
        revertSandboxStateOptimisticUpdate(sandboxId, context.previousState)
        throw error
      }

      await markAllQueriesAsStale()
    },
    [
      cancelOutgoingRefetches,
      getSandboxState,
      markAllQueriesAsStale,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
    ],
  )

  const startSandbox = useCallback(
    async (sandboxId: string) => {
      await runSandboxMutation({
        sandboxId,
        optimisticState: SandboxState.STARTING,
        mutateAsync: startMutation.mutateAsync,
      })
    },
    [runSandboxMutation, startMutation.mutateAsync],
  )

  const recoverSandbox = useCallback(
    async (sandboxId: string) => {
      await runSandboxMutation({
        sandboxId,
        optimisticState: SandboxState.STARTING,
        mutateAsync: recoverMutation.mutateAsync,
      })
    },
    [recoverMutation.mutateAsync, runSandboxMutation],
  )

  const stopSandbox = useCallback(
    async (sandboxId: string) => {
      await runSandboxMutation({
        sandboxId,
        optimisticState: SandboxState.STOPPING,
        mutateAsync: stopMutation.mutateAsync,
      })
    },
    [runSandboxMutation, stopMutation.mutateAsync],
  )

  const archiveSandbox = useCallback(
    async (sandboxId: string) => {
      await runSandboxMutation({
        sandboxId,
        optimisticState: SandboxState.ARCHIVING,
        mutateAsync: archiveMutation.mutateAsync,
      })
    },
    [archiveMutation.mutateAsync, runSandboxMutation],
  )

  const deleteSandbox = useCallback(
    async (sandboxId: string) => {
      await runSandboxMutation({
        sandboxId,
        optimisticState: SandboxState.DESTROYING,
        mutateAsync: deleteMutation.mutateAsync,
      })
    },
    [deleteMutation.mutateAsync, runSandboxMutation],
  )

  return {
    startSandbox,
    recoverSandbox,
    stopSandbox,
    archiveSandbox,
    deleteSandbox,
  }
}
