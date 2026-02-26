/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSyncExternalStore, useMemo, useCallback } from 'react'

function getServerSnapshot(): boolean {
  return false
}

export function useMatchMedia(query: string): boolean {
  const mql = useMemo(() => {
    if (typeof window === 'undefined') return null
    return window.matchMedia(query)
  }, [query])

  const subscribe = useCallback(
    (callback: () => void) => {
      if (!mql) return () => undefined

      mql.addEventListener('change', callback)
      return () => mql.removeEventListener('change', callback)
    },
    [mql],
  )

  const getSnapshot = useCallback(() => {
    return mql ? mql.matches : false
  }, [mql])

  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
}
