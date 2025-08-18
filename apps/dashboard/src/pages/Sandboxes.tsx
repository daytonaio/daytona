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
  PaginatedSandboxes,
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
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
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

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi, snapshotApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()

  const [sandboxesData, setSandboxesData] = useState<PaginatedSandboxes>({
    items: [],
    total: 0,
    page: 1,
    totalPages: 0,
  })
  const [snapshots, setSnapshots] = useState<SnapshotDto[]>([])
  const [loadingSandboxes, setLoadingSandboxes] = useState<Record<string, boolean>>({})
  const [transitioningSandboxes, setTransitioningSandboxes] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [loadingSnapshots, setLoadingSnapshots] = useState(true)
  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

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

  // TODO: search/filters/sort params?
  const fetchSandboxes = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const response = (
          await sandboxApi.listSandboxes(
            selectedOrganization.id,
            undefined, // verbose
            undefined, // labels
            undefined, // includeErroredDeleted
            paginationParams.pageSize,
            paginationParams.pageIndex + 1,
          )
        ).data
        setSandboxesData(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch sandboxes')
      } finally {
        setLoadingTable(false)
      }
    },
    [sandboxApi, selectedOrganization, paginationParams.pageIndex, paginationParams.pageSize],
  )

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  useEffect(() => {
    fetchSandboxes()
    fetchSnapshots()
  }, [fetchSandboxes, fetchSnapshots])

  useEffect(() => {
    if (selectedSandbox) {
      const updatedSandbox = sandboxesData.items.find((s) => s.id === selectedSandbox.id)
      if (updatedSandbox && updatedSandbox !== selectedSandbox) {
        setSelectedSandbox(updatedSandbox)
      }
    }
  }, [sandboxesData.items, selectedSandbox])

  useEffect(() => {
    if (selectedSandbox && !sandboxesData.items.some((s) => s.id === selectedSandbox.id)) {
      setSelectedSandbox(null)
      setShowSandboxDetails(false)
    }
  }, [sandboxesData.items, selectedSandbox])

  useEffect(() => {
    const handleSandboxCreatedEvent = (sandbox: Sandbox) => {
      if (paginationParams.pageIndex === 0) {
        setSandboxesData((prev) => {
          if (prev.items.some((s) => s.id === sandbox.id)) {
            return prev
          }

          const newSandboxes = [sandbox, ...prev.items]
          const newTotal = prev.total + 1
          return {
            ...prev,
            items: newSandboxes.slice(0, paginationParams.pageSize),
            total: newTotal,
            totalPages: Math.ceil(newTotal / paginationParams.pageSize),
          }
        })
      }
    }

    const handleSandboxStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldState: SandboxState
      newState: SandboxState
    }) => {
      if (data.newState === SandboxState.DESTROYED) {
        setSandboxesData((prev) => {
          const newTotal = Math.max(0, prev.total - 1)
          const newItems = prev.items.filter((s) => s.id !== data.sandbox.id)

          return {
            ...prev,
            items: newItems,
            total: newTotal,
            totalPages: Math.ceil(newTotal / paginationParams.pageSize),
          }
        })
      } else {
        setSandboxesData((prev) => ({
          ...prev,
          items: prev.items.some((s) => s.id === data.sandbox.id)
            ? prev.items.map((s) => (s.id === data.sandbox.id ? data.sandbox : s))
            : paginationParams.pageIndex === 0
              ? [data.sandbox, ...prev.items.slice(0, paginationParams.pageSize - 1)]
              : prev.items,
        }))
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
        setSandboxesData((prev) => {
          const newTotal = Math.max(0, prev.total - 1)
          const newItems = prev.items.filter((s) => s.id !== data.sandbox.id)

          return {
            ...prev,
            items: newItems,
            total: newTotal,
            totalPages: Math.ceil(newTotal / paginationParams.pageSize),
          }
        })
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
  }, [notificationSocket, paginationParams.pageIndex, paginationParams.pageSize])

  useEffect(() => {
    if (sandboxesData.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [sandboxesData.items.length, paginationParams.pageIndex])

  const handleStart = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    setTransitioningSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToStart = sandboxesData.items.find((s) => s.id === id)
    const previousState = sandboxToStart?.state

    setSandboxesData((prev) => ({
      ...prev,
      items: prev.items.map((s) => (s.id === id ? { ...s, state: SandboxState.STARTING } : s)),
    }))

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
      setSandboxesData((prev) => ({
        ...prev,
        items: prev.items.map((s) => (s.id === id ? { ...s, state: previousState } : s)),
      }))
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

    const sandboxToStop = sandboxesData.items.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    setSandboxesData((prev) => ({
      ...prev,
      items: prev.items.map((s) => (s.id === id ? { ...s, state: SandboxState.STOPPING } : s)),
    }))

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
      setSandboxesData((prev) => ({
        ...prev,
        items: prev.items.map((s) => (s.id === id ? { ...s, state: previousState } : s)),
      }))
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

    const sandboxToDelete = sandboxesData.items.find((s) => s.id === id)
    const previousState = sandboxToDelete?.state

    setSandboxesData((prev) => ({
      ...prev,
      items: prev.items.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)),
    }))

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
      setSandboxesData((prev) => ({
        ...prev,
        items: prev.items.map((s) => (s.id === id ? { ...s, state: previousState } : s)),
      }))
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
      const sandboxToDelete = sandboxesData.items.find((s) => s.id === id)
      const previousState = sandboxToDelete?.state

      setSandboxesData((prev) => ({
        ...prev,
        items: prev.items.map((s) => (s.id === id ? { ...s, state: SandboxState.DESTROYING } : s)),
      }))

      if (selectedSandbox?.id === id) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.DESTROYING } : null))
      }

      try {
        await sandboxApi.deleteSandbox(id, true, selectedOrganization?.id)
        toast.success(`Deleting sandbox with ID: ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')

        setSandboxesData((prev) => ({
          ...prev,
          items: prev.items.map((s) => (s.id === id ? { ...s, state: previousState } : s)),
        }))
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

    const sandboxToArchive = sandboxesData.items.find((s) => s.id === id)
    const previousState = sandboxToArchive?.state

    setSandboxesData((prev) => ({
      ...prev,
      items: prev.items.map((s) => (s.id === id ? { ...s, state: SandboxState.ARCHIVING } : s)),
    }))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.ARCHIVING } : null))
    }

    try {
      await sandboxApi.archiveSandbox(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      setSandboxesData((prev) => ({
        ...prev,
        items: prev.items.map((s) => (s.id === id ? { ...s, state: previousState } : s)),
      }))
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
        {!loadingTable && sandboxesData.items.length === 0 && (
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
        data={sandboxesData.items}
        loading={loadingTable}
        snapshots={snapshots}
        loadingSnapshots={loadingSnapshots}
        onRowClick={(sandbox: Sandbox) => {
          setSelectedSandbox(sandbox)
          setShowSandboxDetails(true)
        }}
        pageCount={sandboxesData.totalPages}
        onPaginationChange={handlePaginationChange}
        pagination={{
          pageIndex: paginationParams.pageIndex,
          pageSize: paginationParams.pageSize,
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
