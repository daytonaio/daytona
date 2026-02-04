/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePlayground } from '@/hooks/usePlayground'
import { Sandbox } from '@daytonaio/sdk'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

export type UsePlaygroundSandboxResult = {
  sandbox: Sandbox | null
  isLoading: boolean
  error: string | null
  updateSandbox: (sandbox: Sandbox) => Promise<void>
  createSandboxFromParams: () => Promise<Sandbox>
}

/**
 * This hook manages the playground sandbox lifecycle.
 *
 * The sandbox is stored in PlaygroundContext so it can be shared across all playground tabs.
 *
 * Sandbox creation behavior:
 * - Sandboxes are always created using parameters from the Sandbox tab Management section
 *   (resources, auto-stop interval, etc.) as the single source of truth.
 * - If no sandbox exists and `disableSandboxAutoCreate` is not true, the hook will automatically
 *   create a sandbox using the current parameters.
 * - Sandboxes created by clicking the "Run" button in the Sandbox tab take precedence -
 *   the playground sandbox will always be set to the one created from that action.
 *
 * The `disableSandboxAutoCreate` flag should be set to true for the Sandbox tab since sandbox
 * creation there is triggered manually via the "Run" button which executes the auto-generated
 * code snippet.
 */
export function usePlaygroundSandbox(disableSandboxAutoCreate?: boolean): UsePlaygroundSandboxResult {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const creatingRef = useRef(false)

  const { DaytonaClient, sandbox, setSandbox, getSandboxParametersInfo } = usePlayground()

  const updateSandbox = useCallback(
    async (newSandbox: Sandbox) => {
      setSandbox(newSandbox)
    },
    [setSandbox],
  )

  const createSandboxFromParams = useCallback(async (): Promise<Sandbox> => {
    const { createSandboxParams } = getSandboxParametersInfo()
    if (!DaytonaClient) throw new Error('Unable to create Daytona client: missing access token or organization ID.')
    return await DaytonaClient.create(createSandboxParams)
  }, [DaytonaClient, getSandboxParametersInfo])

  const createSandbox = useCallback(async () => {
    // Sandbox already created and stored in context -> skip creation
    if (sandbox) {
      return
    }
    // Prevent concurrent creation attempts
    if (creatingRef.current) {
      return
    }
    try {
      creatingRef.current = true
      setIsLoading(true)
      setError(null)
      const newSandbox = await createSandboxFromParams()
      setSandbox(newSandbox)
    } catch (error) {
      console.error('Failed to create sandbox:', error)
      toast.error('Failed to create sandbox', {
        action: {
          label: 'Try again',
          onClick: () => {
            createSandbox()
          },
        },
      })
      setError(error instanceof Error ? error.message : String(error))
    } finally {
      setIsLoading(false)
      creatingRef.current = false
    }
  }, [sandbox, setSandbox, createSandboxFromParams])

  useEffect(() => {
    if (!disableSandboxAutoCreate) createSandbox()
  }, [disableSandboxAutoCreate, createSandbox])

  return useMemo(
    () => ({
      sandbox,
      isLoading,
      error,
      updateSandbox,
      createSandboxFromParams,
    }),
    [sandbox, isLoading, error, updateSandbox, createSandboxFromParams],
  )
}
