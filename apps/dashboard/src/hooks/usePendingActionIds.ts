/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MutationKey, useMutationState } from '@tanstack/react-query'
import { useMemo } from 'react'

interface UsePendingActionIdsOptions<TVariables> {
  mutationKey: MutationKey
  getId: (variables: TVariables | undefined) => string | undefined
}

export function usePendingActionIds<TVariables>({ mutationKey, getId }: UsePendingActionIdsOptions<TVariables>) {
  return useMutationState<string | undefined>({
    filters: { mutationKey, status: 'pending' },
    select: (mutation) => {
      const variables = mutation.state.variables as TVariables | undefined
      return getId(variables)
    },
  })
}

interface PendingActionSelector {
  mutationKey: MutationKey
  getId: (variables: unknown) => string | undefined
}

const isMutationKeyPrefixMatch = (actual: MutationKey | undefined, expected: MutationKey) => {
  if (!actual || actual.length < expected.length) {
    return false
  }

  return expected.every((value, index) => actual[index] === value)
}

export function usePendingActionMap(selectors: PendingActionSelector[]) {
  const pendingRawIds = useMutationState<string | undefined>({
    filters: {
      status: 'pending',
      predicate: (mutation) =>
        selectors.some((selector) => isMutationKeyPrefixMatch(mutation.options.mutationKey, selector.mutationKey)),
    },
    select: (mutation) => {
      for (const selector of selectors) {
        if (!isMutationKeyPrefixMatch(mutation.options.mutationKey, selector.mutationKey)) {
          continue
        }

        return selector.getId(mutation.state.variables)
      }

      return undefined
    },
  })

  const pendingIds = useMemo(() => pendingRawIds.filter((id): id is string => Boolean(id)), [pendingRawIds])

  const loadingAction = useMemo<Record<string, boolean>>(() => {
    const loading: Record<string, boolean> = {}
    pendingIds.forEach((id) => {
      loading[id] = true
    })
    return loading
  }, [pendingIds])

  return {
    pendingIds,
    loadingAction,
  }
}
