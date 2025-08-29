/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState, useCallback } from 'react'
import { useApi } from '@/hooks/useApi'
import { OrganizationSuspendedError } from '@/api/errors'
import {
  OrganizationUserRoleEnum,
  Sandbox,
  SandboxDesiredState,
  SandboxState,
  SnapshotDto,
} from '@daytonaio/api-client'
import { SandboxTable } from '@/components/SandboxTable'
import { Button, buttonVariants } from '@/components/ui/button'
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import SandboxDetailsSheet from '@/components/SandboxDetailsSheet'
import { formatDuration } from '@/lib/utils'
import { Label } from '@/components/ui/label'
import { Check, Copy } from 'lucide-react'

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi, snapshotApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()

  const [sandboxes, setSandboxes] = useState<Sandbox[]>([])
  const [snapshots, setSnapshots] = useState<SnapshotDto[]>([])
  const [loadingSandboxes, setLoadingSandboxes] = useState<Record<string, boolean>>({})
  const [transitioningSandboxes, setTransitioningSandboxes] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [loadingSnapshots, setLoadingSnapshots] = useState(true)
  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)
  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshToken, setSshToken] = useState<string>('')
  const [sshExpiryMinutes, setSshExpiryMinutes] = useState<number>(60)
  const [revokeSshToken, setRevokeSshToken] = useState<string>('')
  const [sshSandboxId, setSshSandboxId] = useState<string>('')
  const [copied, setCopied] = useState<string | null>(null)

  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const fetchSnapshots = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setLoadingSnapshots(true)
    try {
      // TODO: Implement snapshot search by input
      // e.g. "Search to load more results"
      const response = await snapshotApi.getAllSnapshots(selectedOrganization.id, 100)
      setSnapshots(response.data.items ?? [])
    } catch (error) {
      console.error('Failed to fetch snapshots', error)
    } finally {
      setLoadingSnapshots(false)
    }
  }, [selectedOrganization, snapshotApi])

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
    fetchSnapshots()
  }, [fetchSandboxes, fetchSnapshots])

  useEffect(() => {
    if (selectedSandbox) {
      const updatedSandbox = sandboxes.find((s) => s.id === selectedSandbox.id)
      if (updatedSandbox && updatedSandbox !== selectedSandbox) {
        setSelectedSandbox(updatedSandbox)
      }
    }
  }, [sandboxes, selectedSandbox])

  useEffect(() => {
    if (selectedSandbox && !sandboxes.some((s) => s.id === selectedSandbox.id)) {
      setSelectedSandbox(null)
      setShowSandboxDetails(false)
    }
  }, [sandboxes, selectedSandbox])

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

    if (!notificationSocket) {
      return
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
    setTransitioningSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToStart = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToStart?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STARTING } : s)))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.STARTING } : null))
    }

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
          <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING_WALLET)}>
            Go to billing
          </Button>
        ) : undefined,
      )
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
      if (selectedSandbox?.id === id && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setTransitioningSandboxes((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleStop = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    setTransitioningSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToStop = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STOPPING } : s)))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.STOPPING } : null))
    }

    try {
      await sandboxApi.stopSandbox(id, selectedOrganization?.id)
      toast.success(
        `Stopping sandbox with ID: ${id}`,
        sandboxToStop?.autoDeleteInterval !== undefined && sandboxToStop.autoDeleteInterval >= 0
          ? {
              description: `This sandbox will be deleted automatically ${sandboxToStop.autoDeleteInterval === 0 ? 'upon stopping' : `in ${formatDuration(sandboxToStop.autoDeleteInterval)} unless it is started again`}.`,
            }
          : undefined,
      )
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
      if (selectedSandbox?.id === id && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setTransitioningSandboxes((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleDelete = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToDelete = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToDelete?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.DESTROYING } : null))
    }

    try {
      await sandboxApi.deleteSandbox(id, true, selectedOrganization?.id)
      setSandboxToDelete(null)
      setShowDeleteDialog(false)

      if (selectedSandbox?.id === id) {
        setShowSandboxDetails(false)
        setSelectedSandbox(null)
      }

      toast.success(`Deleting sandbox with ID:  ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
      if (selectedSandbox?.id === id && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleBulkDelete = async (ids: string[]) => {
    setLoadingSandboxes((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))

    const selectedSandboxInBulk = selectedSandbox && ids.includes(selectedSandbox.id)

    for (const id of ids) {
      const sandboxToDelete = sandboxes.find((s) => s.id === id)
      const previousState = sandboxToDelete?.state

      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)))

      if (selectedSandbox?.id === id) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.DESTROYING } : null))
      }

      try {
        await sandboxApi.deleteSandbox(id, true, selectedOrganization?.id)
        toast.success(`Deleting sandbox with ID: ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')

        setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
        if (selectedSandbox?.id === id && previousState) {
          setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
        }

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

    if (selectedSandboxInBulk) {
      setShowSandboxDetails(false)
      setSelectedSandbox(null)
    }
  }

  const handleArchive = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToArchive = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToArchive?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.ARCHIVING } : s)))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.ARCHIVING } : null))
    }

    try {
      await sandboxApi.archiveSandbox(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
      if (selectedSandbox?.id === id && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number): Promise<string> => {
      setLoadingSandboxes((prev) => ({ ...prev, [sandboxId]: true }))
      try {
        return (await sandboxApi.getPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
      } finally {
        setLoadingSandboxes((prev) => ({ ...prev, [sandboxId]: false }))
      }
    },
    [sandboxApi, selectedOrganization],
  )

  const getVncUrl = async (sandboxId: string): Promise<string | null> => {
    try {
      const portPreviewUrl = await getPortPreviewUrl(sandboxId, 6080)
      return portPreviewUrl + '/vnc.html'
    } catch (error) {
      handleApiError(error, 'Failed to construct VNC URL')
      return null
    }
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
        const vncUrl = await getVncUrl(id)
        if (vncUrl) {
          window.open(vncUrl, '_blank')
          toast.success('Opening VNC desktop...')
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
              const vncUrl = await getVncUrl(id)

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

  const getWebTerminalUrl = useCallback(
    async (sandboxId: string): Promise<string | null> => {
      try {
        return await getPortPreviewUrl(sandboxId, 22222)
      } catch (error) {
        handleApiError(error, 'Failed to construct web terminal URL')
        return null
      }
    },
    [getPortPreviewUrl],
  )

  const handleCreateSshAccess = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    try {
      const response = await sandboxApi.createSshAccess(id, selectedOrganization?.id, sshExpiryMinutes)
      setSshToken(response.data.token)
      setSshSandboxId(id)
      setShowCreateSshDialog(true)
      toast.success('SSH access created successfully')
    } catch (error) {
      handleApiError(error, 'Failed to create SSH access')
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const openCreateSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowCreateSshDialog(true)
  }

  const handleRevokeSshAccess = async (id: string) => {
    if (!revokeSshToken.trim()) {
      toast.error('Please enter a token to revoke')
      return
    }

    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    try {
      await sandboxApi.revokeSshAccess(id, selectedOrganization?.id, revokeSshToken)
      setRevokeSshToken('')
      setSshSandboxId('')
      setShowRevokeSshDialog(false)
      toast.success('SSH access revoked successfully')
    } catch (error) {
      handleApiError(error, 'Failed to revoke SSH access')
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const openRevokeSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowRevokeSshDialog(true)
  }

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(label)
      setTimeout(() => setCopied(null), 2000)
    } catch (err) {
      console.error('Failed to copy text:', err)
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
          setLocalStorageItem(skipOnboardingKey, 'true')
        }
      } catch (error) {
        console.error('Failed to check if user needs onboarding', error)
      }
    }

    onboardIfNeeded()
  }, [navigate, user, selectedOrganization, apiKeyApi])

  return (
    <div className="flex flex-col min-h-dvh px-10 py-3">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Sandboxes</h1>
        {!loadingTable && sandboxes.length === 0 && (
          <div className="flex items-center gap-2">
            <Button variant="link" className="text-primary" onClick={() => navigate(RoutePath.ONBOARDING)}>
              Onboarding guide
            </Button>
            <Button variant="link" className="text-primary" asChild>
              <a href={DAYTONA_DOCS_URL} target="_blank" rel="noopener noreferrer" className="text-primary">
                Docs
              </a>
            </Button>
          </div>
        )}
      </div>

      <SandboxTable
        loadingSandboxes={loadingSandboxes}
        transitioningSandboxes={transitioningSandboxes}
        handleStart={handleStart}
        handleStop={handleStop}
        handleDelete={(id: string) => {
          setSandboxToDelete(id)
          setShowDeleteDialog(true)
        }}
        handleBulkDelete={handleBulkDelete}
        handleArchive={handleArchive}
        handleVnc={handleVnc}
        getWebTerminalUrl={getWebTerminalUrl}
        handleCreateSshAccess={openCreateSshDialog}
        handleRevokeSshAccess={openRevokeSshDialog}
        data={sandboxes}
        loading={loadingTable}
        snapshots={snapshots}
        loadingSnapshots={loadingSnapshots}
        onRowClick={(sandbox: Sandbox) => {
          setSelectedSandbox(sandbox)
          setShowSandboxDetails(true)
        }}
      />

      {sandboxToDelete && (
        <AlertDialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setSandboxToDelete(null)
            }
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Confirm Sandbox Deletion</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to delete this sandbox? This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                className={buttonVariants({ variant: 'destructive' })}
                onClick={() => handleDelete(sandboxToDelete)}
                disabled={loadingSandboxes[sandboxToDelete]}
              >
                {loadingSandboxes[sandboxToDelete] ? 'Deleting...' : 'Delete'}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {/* Create SSH Access Dialog */}
      <AlertDialog
        open={showCreateSshDialog}
        onOpenChange={(isOpen) => {
          setShowCreateSshDialog(isOpen)
          if (!isOpen) {
            setSshToken('')
            setSshExpiryMinutes(60)
            setSshSandboxId('')
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Create SSH Access</AlertDialogTitle>
            <AlertDialogDescription>
              {sshToken
                ? 'SSH access has been created successfully. Use the token below to connect:'
                : 'Set the expiration time for SSH access:'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-4">
            {!sshToken ? (
              <div className="space-y-3">
                <Label className="text-sm font-medium">Expiry (minutes):</Label>
                <input
                  type="number"
                  min="1"
                  max="1440"
                  value={sshExpiryMinutes}
                  onChange={(e) => setSshExpiryMinutes(Number(e.target.value))}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                />
              </div>
            ) : (
              <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                <span className="overflow-x-auto pr-2 cursor-text select-all">
                  {import.meta.env.VITE_SSH_GATEWAY_COMMAND?.replace('{{TOKEN}}', sshToken) ||
                    `ssh -p 22222 user@host -o ProxyCommand="echo ${sshToken}"`}
                </span>
                {(copied === 'SSH Command' && <Check className="w-4 h-4" />) || (
                  <Copy
                    className="w-4 h-4 cursor-pointer"
                    onClick={() =>
                      copyToClipboard(
                        import.meta.env.VITE_SSH_GATEWAY_COMMAND?.replace('{{TOKEN}}', sshToken) ||
                          `ssh -p 22222 user@host -o ProxyCommand="echo ${sshToken}"`,
                        'SSH Command',
                      )
                    }
                  />
                )}
              </div>
            )}
          </div>
          <AlertDialogFooter>
            {!sshToken ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={() => handleCreateSshAccess(sshSandboxId)}
                  disabled={!sshSandboxId}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  Create
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => setShowCreateSshDialog(false)}
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
              >
                Close
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Revoke SSH Access Dialog */}
      <AlertDialog
        open={showRevokeSshDialog}
        onOpenChange={(isOpen) => {
          setShowRevokeSshDialog(isOpen)
          if (!isOpen) {
            setRevokeSshToken('')
            setSshSandboxId('')
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke SSH Access</AlertDialogTitle>
            <AlertDialogDescription>Enter the SSH access token you want to revoke:</AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-4">
            <div className="space-y-3">
              <label className="text-sm font-medium">SSH Token:</label>
              <input
                type="text"
                value={revokeSshToken}
                onChange={(e) => setRevokeSshToken(e.target.value)}
                placeholder="Enter SSH token to revoke"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => handleRevokeSshAccess(sshSandboxId)}
              disabled={!revokeSshToken.trim() || !sshSandboxId}
              className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
            >
              Revoke Access
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <SandboxDetailsSheet
        sandbox={selectedSandbox}
        open={showSandboxDetails}
        onOpenChange={setShowSandboxDetails}
        loadingSandboxes={loadingSandboxes}
        handleStart={handleStart}
        handleStop={handleStop}
        handleDelete={(id) => {
          setSandboxToDelete(id)
          setShowDeleteDialog(true)
          setShowSandboxDetails(false)
        }}
        handleArchive={handleArchive}
        getWebTerminalUrl={getWebTerminalUrl}
        writePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
        deletePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
      />
    </div>
  )
}

export default Sandboxes
