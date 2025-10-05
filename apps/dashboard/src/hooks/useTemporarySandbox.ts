/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useRef, useCallback, useState, useMemo } from 'react'
import { usePlayground } from '@/hooks/usePlayground'
import { Sandbox } from '@daytonaio/sdk'

export type UseTemporarySandboxResult = {
  sandbox: Sandbox | null
  isLoading: boolean
  error: string | null
}

// This hook manages the full lifecycle of a temporary sandbox: it creates a sandbox when the component mounts and automatically deletes it when the component unmounts.
// NOTE: Using in development mode with React.Strict will trigger double call of useEffect and will cause 2 sandbox creations and only 1 sandbox deletion -> result is 1 dangling sandbox. To prevent this comment React.Strict element.
export function useTemporarySandbox(): UseTemporarySandboxResult {
  const sandboxRef = useRef<Sandbox | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { DaytonaClient } = usePlayground()

  const createSandbox = useCallback(async () => {
    // Delete previous sandbox if it exists
    if (sandboxRef.current) {
      try {
        await sandboxRef.current.delete()
      } catch (error) {
        console.error('Failed to delete previous sandbox:', error)
      }
      sandboxRef.current = null
    }

    // Create new sandbox
    try {
      setIsLoading(true)
      setError(null)
      const sandbox = await DaytonaClient.create()
      sandboxRef.current = sandbox
      return sandbox
    } catch (error) {
      console.error('Failed to create sandbox:', error)
      setError(error instanceof Error ? error.message : String(error))
      return null
    } finally {
      setIsLoading(false)
    }
  }, [DaytonaClient])

  useEffect(() => {
    createSandbox()

    // Cleanup on unmount: delete terminal sandbox if exists
    return () => {
      if (sandboxRef.current)
        sandboxRef.current.delete().catch((error) => console.error('Fauled to delete sandbox on unmount:', error))
      sandboxRef.current = null
    }
  }, [createSandbox])

  return useMemo(
    () => ({
      sandbox: sandboxRef.current,
      isLoading,
      error,
    }),
    [sandboxRef, isLoading, error],
  )
}
