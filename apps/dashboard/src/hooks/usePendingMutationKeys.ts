/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MutationKey, useMutationState } from '@tanstack/react-query'
import { useMemo } from 'react'

interface PendingMutationSelector<T extends string, TVariables = unknown> {
  mutationKey: MutationKey
  getKey: (variables: TVariables | undefined) => T | undefined
}

const isMutationKeyPrefixMatch = (actual: MutationKey | undefined, expected: MutationKey) => {
  if (!actual || actual.length < expected.length) {
    return false
  }

  return expected.every((value, index) => actual[index] === value)
}

export function usePendingMutationKeys<T extends string, TVariables = unknown>(
  selectors: PendingMutationSelector<T, TVariables>[],
) {
  const pendingRawKeys = useMutationState<T | undefined>({
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

        return selector.getKey(mutation.state.variables as TVariables | undefined)
      }

      return undefined
    },
  })

  return useMemo(() => new Set(pendingRawKeys.filter((key): key is T => Boolean(key))), [pendingRawKeys])
}
