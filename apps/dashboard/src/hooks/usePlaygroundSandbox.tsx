/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePlayground } from '@/hooks/usePlayground'
import { PlaygroundCategories } from '@/enums/Playground'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { Sandbox } from '@daytonaio/sdk'
import { ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

export type UsePlaygroundSandboxResult = {
  sandbox: Sandbox | null
  sandboxLoading: boolean
  sandboxError: string | null
  updateSandbox: (sandbox: Sandbox) => Promise<void>
  createSandboxFromParams: () => Promise<Sandbox>
  terminalUrlLoading: boolean
  terminalUrlError: string | null
  refetchTerminalUrl: () => void
  vncUrlLoading: boolean
  vncUrlError: string | ReactNode
  refetchVNCUrl: () => void
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
export function usePlaygroundSandbox(
  playgroundCategory: PlaygroundCategories,
  disableSandboxAutoCreate?: boolean,
): UsePlaygroundSandboxResult {
  const [sandboxLoading, setSandboxLoading] = useState(false)
  const [sandboxError, setSandboxError] = useState<string | null>(null)
  const sandboxCreatingRef = useRef(false)

  const [terminalUrlLoading, setTerminalUrlLoading] = useState(false)
  const [terminalUrlError, setTerminalUrlError] = useState<string | null>(null)
  const [vncUrlLoading, setVncUrlLoading] = useState(false)
  const [vncUrlError, setVncUrlError] = useState<string | ReactNode>(null)

  const {
    DaytonaClient,
    sandbox,
    setSandbox,
    getSandboxParametersInfo,
    setTerminalUrl,
    setVNCInteractionOptionsParamValue,
    terminalUrl,
    VNCInteractionOptionsParamsState,
  } = usePlayground()
  const VNCUrl = VNCInteractionOptionsParamsState.VNCUrl

  const { sandboxApi, toolboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number) =>
      (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url,
    [sandboxApi, selectedOrganization],
  )

  const fetchTerminalUrl = useCallback(
    async (sandboxId: string) => {
      setTerminalUrl(null)
      setTerminalUrlLoading(true)
      setTerminalUrlError(null)
      try {
        setTerminalUrl(await getPortPreviewUrl(sandboxId, 22222))
      } catch (error) {
        handleApiError(error, 'Failed to construct web terminal URL')
        setTerminalUrlError(error instanceof Error ? error.message : String(error))
      } finally {
        setTerminalUrlLoading(false)
      }
    },
    [getPortPreviewUrl, setTerminalUrl],
  )

  const fetchVNCUrl = useCallback(
    async (fetchSandbox: Sandbox) => {
      const showToast = playgroundCategory === PlaygroundCategories.VNC
      setVNCInteractionOptionsParamValue('VNCUrl', null)
      setVncUrlLoading(true)
      setVncUrlError(null)
      try {
        if (showToast) toast.info('Checking VNC desktop status...')
        const {
          data: { status },
        } = await toolboxApi.getComputerUseStatusDeprecated(fetchSandbox.id, selectedOrganization?.id)
        if (status === 'active') {
          const url = await getPortPreviewUrl(fetchSandbox.id, 6080)
          setVNCInteractionOptionsParamValue('VNCUrl', url + '/vnc.html')
        } else {
          await toolboxApi.startComputerUseDeprecated(fetchSandbox.id, selectedOrganization?.id)
          if (showToast) toast.success('Starting VNC desktop...')
          await new Promise((resolve) => setTimeout(resolve, 5000))
          const newStatusResponse = await toolboxApi.getComputerUseStatusDeprecated(
            fetchSandbox.id,
            selectedOrganization?.id,
          )
          if (newStatusResponse.data.status === 'active') {
            const url = await getPortPreviewUrl(fetchSandbox.id, 6080)
            setVNCInteractionOptionsParamValue('VNCUrl', url + '/vnc.html')
          } else {
            if (showToast) toast.error(`VNC desktop failed to start. Status: ${newStatusResponse.data.status}`)
            setVncUrlError(`VNC desktop failed to start. Status: ${newStatusResponse.data.status}`)
          }
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error)
        if (message === 'Computer-use functionality is not available') {
          const errorContent = (
            <div>
              <div>Computer-use dependencies are missing in the runtime environment.</div>
              <div className="mt-2">
                <a
                  href={`${DAYTONA_DOCS_URL}/en/vnc-access/`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  See documentation on how to configure the runtime for computer-use
                </a>
              </div>
            </div>
          )
          if (showToast)
            toast.error('Computer-use functionality is not available', {
              description: errorContent,
            })
          setVncUrlError(errorContent)
        } else {
          handleApiError(error, 'Failed to check VNC status')
        }
      } finally {
        setVncUrlLoading(false)
      }
    },
    [toolboxApi, selectedOrganization, getPortPreviewUrl, setVNCInteractionOptionsParamValue],
  )

  const refetchTerminalUrl = useCallback(() => {
    if (sandbox) fetchTerminalUrl(sandbox.id)
  }, [sandbox, fetchTerminalUrl])

  const refetchVNCUrl = useCallback(() => {
    if (sandbox) fetchVNCUrl(sandbox)
  }, [sandbox, fetchVNCUrl])

  const updateSandbox = useCallback(
    async (newSandbox: Sandbox) => {
      setSandbox(newSandbox)
      fetchTerminalUrl(newSandbox.id)
      fetchVNCUrl(newSandbox)
    },
    [setSandbox, fetchTerminalUrl, fetchVNCUrl],
  )

  const createSandboxFromParams = useCallback(async (): Promise<Sandbox> => {
    const { createSandboxParams } = getSandboxParametersInfo()
    if (!DaytonaClient) throw new Error('Unable to create Daytona client: missing access token or organization ID.')
    return await DaytonaClient.create(createSandboxParams)
  }, [DaytonaClient, getSandboxParametersInfo])

  const createSandbox = useCallback(async () => {
    // Sandbox already created and stored in context -> retry any missing URLs
    if (sandbox) {
      if (playgroundCategory === PlaygroundCategories.TERMINAL && !terminalUrl) fetchTerminalUrl(sandbox.id)
      if (playgroundCategory === PlaygroundCategories.VNC && !VNCUrl) fetchVNCUrl(sandbox)
      return
    }
    // Prevent concurrent creation attempts
    if (sandboxCreatingRef.current) {
      return
    }
    try {
      sandboxCreatingRef.current = true
      setSandboxLoading(true)
      setSandboxError(null)
      const newSandbox = await createSandboxFromParams()
      setSandbox(newSandbox)
      fetchTerminalUrl(newSandbox.id)
      fetchVNCUrl(newSandbox)
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error)
      toast.error('Failed to create sandbox', {
        description: (
          <div>
            <div>{errorMessage}</div>
          </div>
        ),
        action: {
          label: 'Try again',
          onClick: () => {
            createSandbox()
          },
        },
      })
      setSandboxError(errorMessage)
    } finally {
      setSandboxLoading(false)
      sandboxCreatingRef.current = false
    }
  }, [sandbox, setSandbox, createSandboxFromParams, fetchTerminalUrl, fetchVNCUrl, terminalUrl, VNCUrl])

  useEffect(() => {
    if (!disableSandboxAutoCreate) createSandbox()
  }, [disableSandboxAutoCreate, createSandbox])

  return useMemo(
    () => ({
      sandbox,
      sandboxLoading,
      sandboxError,
      updateSandbox,
      createSandboxFromParams,
      terminalUrlLoading,
      terminalUrlError,
      refetchTerminalUrl,
      vncUrlLoading,
      vncUrlError,
      refetchVNCUrl,
    }),
    [
      sandbox,
      sandboxLoading,
      sandboxError,
      updateSandbox,
      createSandboxFromParams,
      terminalUrlLoading,
      terminalUrlError,
      refetchTerminalUrl,
      vncUrlLoading,
      vncUrlError,
      refetchVNCUrl,
    ],
  )
}
