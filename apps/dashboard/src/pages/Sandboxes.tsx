/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState, useCallback } from 'react'
import { useApi } from '@/hooks/useApi'
import { OrganizationSuspendedError } from '@/api/errors'
import { OrganizationUserRoleEnum, Sandbox, SandboxDesiredState, SandboxState } from '@daytonaio/api-client'
import { SandboxTable } from '@/components/SandboxTable'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNavigate } from 'react-router-dom'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { handleApiError } from '@/lib/error-handling'
import { RoutePath } from '@/enums/RoutePath'
import { useAuth } from 'react-oidc-context'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()

  const [sandboxes, setSandboxes] = useState<Sandbox[]>([])
  const [loadingSandboxes, setLoadingSandboxes] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const fetchSandboxes = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const sandboxes = (await sandboxApi.listSandboxes(selectedOrganization.id)).data
        setSandboxes(sandboxes)
      } catch (error) {
        handleApiError(error, 'Failed to fetch sandboxes')
      } finally {
        setLoadingTable(false)
      }
    },
    [sandboxApi, selectedOrganization],
  )

  useEffect(() => {
    fetchSandboxes()
  }, [fetchSandboxes])

  useEffect(() => {
    const handleSandboxCreatedEvent = (sandbox: Sandbox) => {
      if (!sandboxes.some((s) => s.id === sandbox.id)) {
        setSandboxes((prev) => [sandbox, ...prev])
      }
    }

    const handleSandboxStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldState: SandboxState
      newState: SandboxState
    }) => {
      if (data.newState === SandboxState.DESTROYED) {
        setSandboxes((prev) => prev.filter((s) => s.id !== data.sandbox.id))
      } else if (!sandboxes.some((s) => s.id === data.sandbox.id)) {
        setSandboxes((prev) => [data.sandbox, ...prev])
      } else {
        setSandboxes((prev) => prev.map((s) => (s.id === data.sandbox.id ? data.sandbox : s)))
      }
    }

    const handleSandboxDesiredStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }) => {
      if (
        data.newDesiredState === SandboxDesiredState.DESTROYED &&
        data.sandbox.state &&
        ([SandboxState.ERROR, SandboxState.BUILD_FAILED] as SandboxState[]).includes(data.sandbox.state)
      ) {
        setSandboxes((prev) => prev.filter((s) => s.id !== data.sandbox.id))
      }
    }

    notificationSocket.on('sandbox.created', handleSandboxCreatedEvent)
    notificationSocket.on('sandbox.state.updated', handleSandboxStateUpdatedEvent)
    notificationSocket.on('sandbox.desired-state.updated', handleSandboxDesiredStateUpdatedEvent)

    return () => {
      notificationSocket.off('sandbox.created', handleSandboxCreatedEvent)
      notificationSocket.off('sandbox.state.updated', handleSandboxStateUpdatedEvent)
      notificationSocket.off('sandbox.desired-state.updated', handleSandboxDesiredStateUpdatedEvent)
    }
  }, [notificationSocket, sandboxes])

  const handleStart = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const sandboxToStart = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToStart?.state

    // Optimistically update the sandbox state
    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STARTING } : s)))

    try {
      await sandboxApi.startSandbox(id, selectedOrganization?.id)
      toast.success(`Starting sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(
        error,
        'Failed to start sandbox',
        error instanceof OrganizationSuspendedError &&
          import.meta.env.VITE_BILLING_API_URL &&
          authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER ? (
          <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING)}>
            Go to billing
          </Button>
        ) : undefined,
      )
      // Revert the optimistic update
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleStop = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const sandboxToStop = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    // Optimistically update the sandbox state
    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STOPPING } : s)))

    try {
      await sandboxApi.stopSandbox(id, selectedOrganization?.id)
      toast.success(`Stopping sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      // Revert the optimistic update
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleDelete = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const sandboxToDelete = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToDelete?.state

    // Optimistically update the sandbox state
    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)))

    try {
      await sandboxApi.deleteSandbox(id, true, selectedOrganization?.id)
      setSandboxToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting sandbox with ID:  ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      // Revert the optimistic update
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleBulkDelete = async (ids: string[]) => {
    setLoadingSandboxes((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))

    for (const id of ids) {
      // Save the current state
      const sandboxToDelete = sandboxes.find((s) => s.id === id)
      const previousState = sandboxToDelete?.state

      // Optimistically update the sandbox state
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)))

      try {
        await sandboxApi.deleteSandbox(id, true, selectedOrganization?.id)
        toast.success(`Deleting sandbox with ID: ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')

        // Revert the optimistic update
        setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))

        // Wait for user decision
        const shouldContinue = window.confirm(
          `Failed to delete sandbox with ID: ${id}. Do you want to continue with the remaining sandboxes?`,
        )

        if (!shouldContinue) {
          break
        }
      } finally {
        setLoadingSandboxes((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: false }), {}) }))
      }
    }
  }

  const handleArchive = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const sandboxToArchive = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToArchive?.state

    // Optimistically update the sandbox state
    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.ARCHIVING } : s)))

    try {
      await sandboxApi.archiveSandbox(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      // Revert the optimistic update
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const getVncUrl = (sandboxId: string) => {
    const sandbox = sandboxes.find((s) => s.id === sandboxId)
    if (!sandbox) {
      return null
    }

    if (!sandbox.daemonVersion) {
      return `https://6080-${sandbox.id}.${sandbox.runnerDomain}/vnc.html`
    }

    return (
      import.meta.env.VITE_PROXY_TEMPLATE_URL?.replace('{{PORT}}', '6080').replace('{{sandboxId}}', sandbox.id) +
      '/vnc.html'
    )
  }

  const handleVnc = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    // Notify user immediately that we're checking VNC status
    toast.info('Checking VNC desktop status...')

    try {
      // First, check if computer use is already started
      const statusResponse = await toolboxApi.getComputerUseStatus(id, selectedOrganization?.id)
      const status = statusResponse.data.status

      // Check if computer use is active (all processes running)
      if (status === 'active') {
        const vncUrl = getVncUrl(id)
        if (vncUrl) {
          window.open(vncUrl, '_blank')
          toast.success('Opening VNC desktop...')
        } else {
          toast.error('Failed to construct VNC URL')
        }
      } else {
        // Computer use is not active, try to start it
        try {
          await toolboxApi.startComputerUse(id, selectedOrganization?.id)
          toast.success('Starting VNC desktop...')

          // Wait a moment for processes to start, then open VNC
          await new Promise((resolve) => setTimeout(resolve, 5000))

          try {
            const newStatusResponse = await toolboxApi.getComputerUseStatus(id, selectedOrganization?.id)
            const newStatus = newStatusResponse.data.status

            if (newStatus === 'active') {
              const vncUrl = getVncUrl(id)

              if (vncUrl) {
                window.open(vncUrl, '_blank')
                toast.success('VNC desktop is ready!', {
                  action: (
                    <Button variant="secondary" onClick={() => window.open(vncUrl, '_blank')}>
                      Open in new tab
                    </Button>
                  ),
                })
              }
            } else {
              toast.error(`VNC desktop failed to start. Status: ${newStatus}`)
            }
          } catch (error) {
            handleApiError(error, 'Failed to check VNC status after start')
          }
        } catch (startError: any) {
          // Check if this is a computer-use availability error
          const errorMessage = startError?.response?.data?.message || startError?.message || String(startError)

          if (errorMessage === 'Computer-use functionality is not available') {
            toast.error('Computer-use functionality is not available', {
              description: (
                <div>
                  <div>Computer-use dependencies are missing in the runtime environment.</div>
                  <div className="mt-2">
                    <a
                      href={`${DAYTONA_DOCS_URL}/getting-started/computer-use`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      See documentation on how to configure the runtime for computer-use
                    </a>
                  </div>
                </div>
              ),
            })
          } else {
            handleApiError(startError, 'Failed to start VNC desktop')
          }
        }
      }
    } catch (error) {
      handleApiError(error, 'Failed to check VNC status')
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  // Redirect user to the onboarding page if they haven't created an api key yet
  // Perform only once per user
  useEffect(() => {
    const onboardIfNeeded = async () => {
      if (!selectedOrganization) {
        return
      }

      const skipOnboardingKey = `${LocalStorageKey.SkipOnboardingPrefix}${user?.profile.sub}`
      const shouldSkipOnboarding = getLocalStorageItem(skipOnboardingKey) === 'true'

      if (shouldSkipOnboarding) {
        return
      }

      try {
        const keys = (await apiKeyApi.listApiKeys(selectedOrganization.id)).data
        if (keys.length === 0) {
          setLocalStorageItem(skipOnboardingKey, 'true')
          navigate(RoutePath.ONBOARDING)
        } else {
          // Future onboarding checks can be skipped for this user because they already created an api key
          setLocalStorageItem(skipOnboardingKey, 'true')
        }
      } catch (error) {
        console.error('Failed to check if user needs onboarding', error)
      }
    }

    onboardIfNeeded()
  }, [navigate, user, selectedOrganization, apiKeyApi])

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Sandboxes</h1>
        {!loadingTable && sandboxes.length === 0 && (
          <p className="text-sm text-muted-foreground mt-1">
            To get started, check out the{' '}
            <button className="text-primary" onClick={() => navigate(RoutePath.ONBOARDING)}>
              Onboarding
            </button>{' '}
            guide. For more examples, check out the{' '}
            <a href={DAYTONA_DOCS_URL} target="_blank" rel="noopener noreferrer" className="text-primary">
              Docs
            </a>
            .
          </p>
        )}
      </div>
      <SandboxTable
        loadingSandboxes={loadingSandboxes}
        handleStart={handleStart}
        handleStop={handleStop}
        handleDelete={(id) => {
          setSandboxToDelete(id)
          setShowDeleteDialog(true)
        }}
        handleBulkDelete={handleBulkDelete}
        handleArchive={handleArchive}
        handleVnc={handleVnc}
        data={sandboxes}
        loading={loadingTable}
      />

      {sandboxToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setSandboxToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Sandbox Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this sandbox? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              <Button
                variant="destructive"
                onClick={() => handleDelete(sandboxToDelete)}
                disabled={loadingSandboxes[sandboxToDelete]}
              >
                {loadingSandboxes[sandboxToDelete] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Sandboxes
