/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationSuspendedError } from '@/api/errors'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import SandboxDetailsSheet from '@/components/SandboxDetailsSheet'
import { SandboxTable } from '@/components/SandboxTable'
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
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { SnapshotFilters, SnapshotQueryParams, useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useRegions } from '@/hooks/useRegions'
import { DEFAULT_SANDBOX_SORTING } from '@/hooks/useSandboxes'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { formatDuration, pluralize } from '@/lib/utils'
import { SandboxListProvider, useSandboxListContext } from '@/providers/SandboxListProvider'
import {
  OrganizationUserRoleEnum,
  Sandbox,
  SandboxDesiredState,
  SandboxState,
  SshAccessDto,
} from '@daytonaio/api-client'
import { Check, Copy } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi } = useApi()
  const { user } = useAuth()
  const navigate = useNavigate()
  const { notificationSocket } = useNotificationSocket()
  const config = useConfig()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const {
    sandboxes: sandboxItems,
    totalItems,
    pageCount,
    isLoading: sandboxesDataIsLoading,
    pagination,
    onPaginationChange: handlePaginationChange,
    sorting,
    onSortingChange: handleSortingChange,
    filters,
    onFiltersChange: handleFiltersChange,
    handleRefresh,
    isRefreshing: sandboxDataIsRefreshing,
    performSandboxStateOptimisticUpdate,
    revertSandboxStateOptimisticUpdate,
    cancelOutgoingRefetches,
    markAllQueriesAsStale,
  } = useSandboxListContext()

  const [sandboxIsLoading, setSandboxIsLoading] = useState<Record<string, boolean>>({})
  const [sandboxStateIsTransitioning, setSandboxStateIsTransitioning] = useState<Record<string, boolean>>({})

  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)

  useEffect(() => {
    if (!selectedSandbox || !sandboxItems) {
      return
    }

    const selectedSandboxInData = sandboxItems.find((s) => s.id === selectedSandbox.id)

    if (!selectedSandboxInData) {
      setSelectedSandbox(null)
      setShowSandboxDetails(false)
      return
    }

    if (selectedSandboxInData !== selectedSandbox) {
      setSelectedSandbox(selectedSandboxInData)
    }
  }, [sandboxItems, selectedSandbox])

  const performOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      performSandboxStateOptimisticUpdate(sandboxId, newState)

      if (selectedSandbox?.id === sandboxId) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: newState } : null))
      }
    },
    [performSandboxStateOptimisticUpdate, selectedSandbox?.id],
  )

  const revertOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      revertSandboxStateOptimisticUpdate(sandboxId, previousState)

      if (selectedSandbox?.id === sandboxId && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    },
    [revertSandboxStateOptimisticUpdate, selectedSandbox?.id],
  )

  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshAccess, setSshAccess] = useState<SshAccessDto | null>(null)
  const [sshExpiryMinutes, setSshExpiryMinutes] = useState<number>(60)
  const [revokeSshToken, setRevokeSshToken] = useState<string>('')
  const [sshSandboxId, setSshSandboxId] = useState<string>('')
  const [copied, setCopied] = useState<string | null>(null)

  const [snapshotFilters, setSnapshotFilters] = useState<SnapshotFilters>({})

  const handleSnapshotFiltersChange = useCallback((filters: Partial<SnapshotFilters>) => {
    setSnapshotFilters((prev) => ({ ...prev, ...filters }))
  }, [])

  const snapshotsQueryParams = useMemo<SnapshotQueryParams>(
    () => ({
      page: 1,
      pageSize: 100,
      filters: snapshotFilters,
    }),
    [snapshotFilters],
  )

  const {
    data: snapshotsData,
    isLoading: snapshotsDataIsLoading,
    error: snapshotsDataError,
  } = useSnapshotsQuery(snapshotsQueryParams)

  const snapshotsDataHasMore = useMemo(() => {
    return snapshotsData && snapshotsData.totalPages > 1
  }, [snapshotsData])

  useEffect(() => {
    if (snapshotsDataError) {
      handleApiError(snapshotsDataError, 'Failed to fetch snapshots')
    }
  }, [snapshotsDataError])

  const { availableRegions: regionsData, loadingAvailableRegions: regionsDataIsLoading, getRegionName } = useRegions()

  useEffect(() => {
    const handleSandboxCreatedEvent = () => {
      const isFirstPage = pagination.pageIndex === 0
      const isDefaultFilters = Object.keys(filters).length === 0
      const isDefaultSorting =
        sorting.field === DEFAULT_SANDBOX_SORTING.field && sorting.direction === DEFAULT_SANDBOX_SORTING.direction

      const shouldRefetchActiveQueries = isFirstPage && isDefaultFilters && isDefaultSorting

      markAllQueriesAsStale(shouldRefetchActiveQueries)
    }

    const handleSandboxStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldState: SandboxState
      newState: SandboxState
    }) => {
      if (data.oldState === data.newState && data.newState === SandboxState.STARTED) {
        handleSandboxCreatedEvent()
        return
      }

      let updatedState = data.newState

      if (
        data.sandbox.desiredState === SandboxDesiredState.DESTROYED &&
        (data.newState === SandboxState.ERROR || data.newState === SandboxState.BUILD_FAILED)
      ) {
        updatedState = SandboxState.DESTROYED
      }

      performOptimisticUpdate(data.sandbox.id, updatedState)

      markAllQueriesAsStale()
    }

    const handleSandboxDesiredStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }) => {
      if (data.newDesiredState !== SandboxDesiredState.DESTROYED) {
        return
      }

      if (data.sandbox.state !== SandboxState.ERROR && data.sandbox.state !== SandboxState.BUILD_FAILED) {
        return
      }

      performOptimisticUpdate(data.sandbox.id, SandboxState.DESTROYED)

      markAllQueriesAsStale()
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
  }, [
    filters,
    markAllQueriesAsStale,
    notificationSocket,
    pagination.pageIndex,
    performOptimisticUpdate,
    sorting.direction,
    sorting.field,
  ])

  const handleStart = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToStart = sandboxItems.find((s) => s.id === id)
    const previousState = sandboxToStart?.state

    await cancelOutgoingRefetches()
    performOptimisticUpdate(id, SandboxState.STARTING)

    try {
      await sandboxApi.startSandbox(id, selectedOrganization?.id)
      toast.success(`Starting sandbox with ID: ${id}`)
      await markAllQueriesAsStale()
    } catch (error) {
      handleApiError(
        error,
        'Failed to start sandbox',
        error instanceof OrganizationSuspendedError &&
          config.billingApiUrl &&
          authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER ? (
          <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING_WALLET)}>
            Go to billing
          </Button>
        ) : undefined,
      )
      revertOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleRecover = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToRecover = sandboxItems.find((s) => s.id === id)
    const previousState = sandboxToRecover?.state

    await cancelOutgoingRefetches()
    performOptimisticUpdate(id, SandboxState.STARTING)

    try {
      await sandboxApi.recoverSandbox(id, selectedOrganization?.id)
      toast.success('Sandbox recovered. Restarting...')
      await markAllQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to recover sandbox')
      revertOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleStop = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToStop = sandboxItems.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    await cancelOutgoingRefetches()
    performOptimisticUpdate(id, SandboxState.STOPPING)

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
      await markAllQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      revertOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleDelete = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToDelete = sandboxItems.find((s) => s.id === id)
    const previousState = sandboxToDelete?.state

    await cancelOutgoingRefetches()
    performOptimisticUpdate(id, SandboxState.DESTROYING)

    try {
      await sandboxApi.deleteSandbox(id, selectedOrganization?.id)
      setSandboxToDelete(null)
      setShowDeleteDialog(false)

      if (selectedSandbox?.id === id) {
        setShowSandboxDetails(false)
        setSelectedSandbox(null)
      }

      toast.success(`Deleting sandbox with ID:  ${id}`)

      await markAllQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      revertOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleArchive = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToArchive = sandboxItems.find((s) => s.id === id)
    const previousState = sandboxToArchive?.state

    await cancelOutgoingRefetches()
    performOptimisticUpdate(id, SandboxState.ARCHIVING)

    try {
      await sandboxApi.archiveSandbox(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
      await markAllQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      revertOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const executeBulkAction = useCallback(
    async ({
      ids,
      actionName,
      optimisticState,
      apiCall,
      toastMessages,
    }: {
      ids: string[]
      actionName: string
      optimisticState: SandboxState
      apiCall: (id: string) => Promise<unknown>
      toastMessages: {
        successTitle: string
        errorTitle: string
        warningTitle: string
        canceledTitle: string
      }
    }) => {
      await cancelOutgoingRefetches()

      const previousStatesById = new Map((sandboxItems ?? []).map((sandbox) => [sandbox.id, sandbox.state]))

      let isCancelled = false
      let processedCount = 0
      let successCount = 0
      let failureCount = 0

      const totalLabel = pluralize(ids.length, 'sandbox', 'sandboxes')
      const onCancel = () => {
        isCancelled = true
      }

      const bulkToast = createBulkActionToast(`${actionName} 0 of ${totalLabel}.`, {
        action: { label: 'Cancel', onClick: onCancel },
      })

      try {
        for (const id of ids) {
          if (isCancelled) break

          processedCount += 1
          bulkToast.loading(`${actionName} ${processedCount} of ${totalLabel}.`, {
            action: { label: 'Cancel', onClick: onCancel },
          })

          setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
          setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))
          performOptimisticUpdate(id, optimisticState)

          try {
            await apiCall(id)
            successCount += 1
          } catch (error) {
            failureCount += 1
            revertOptimisticUpdate(id, previousStatesById.get(id))
            console.error(`${actionName} sandbox failed`, id, error)
          } finally {
            setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
            setTimeout(() => {
              setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
            }, 2000)
          }
        }

        await markAllQueriesAsStale()
        bulkToast.result({ successCount, failureCount }, toastMessages)
      } catch (error) {
        console.error(`${actionName} sandboxes failed`, error)
        bulkToast.error(`${actionName} sandboxes failed.`)
      }

      return { successCount, failureCount }
    },
    [cancelOutgoingRefetches, sandboxItems, performOptimisticUpdate, revertOptimisticUpdate, markAllQueriesAsStale],
  )

  const handleBulkStart = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Starting',
      optimisticState: SandboxState.STARTING,
      apiCall: (id) => sandboxApi.startSandbox(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} started.`,
        errorTitle: `Failed to start ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to start some sandboxes.',
        canceledTitle: 'Start canceled.',
      },
    })

  const handleBulkStop = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Stopping',
      optimisticState: SandboxState.STOPPING,
      apiCall: (id) => sandboxApi.stopSandbox(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} stopped.`,
        errorTitle: `Failed to stop ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to stop some sandboxes.',
        canceledTitle: 'Stop canceled.',
      },
    })

  const handleBulkArchive = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Archiving',
      optimisticState: SandboxState.ARCHIVING,
      apiCall: (id) => sandboxApi.archiveSandbox(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} archived.`,
        errorTitle: `Failed to archive ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to archive some sandboxes.',
        canceledTitle: 'Archive canceled.',
      },
    })

  const handleBulkDelete = async (ids: string[]) => {
    const selectedSandboxInBulk = selectedSandbox && ids.includes(selectedSandbox.id)

    await executeBulkAction({
      ids,
      actionName: 'Deleting',
      optimisticState: SandboxState.DESTROYING,
      apiCall: (id) => sandboxApi.deleteSandbox(id, selectedOrganization?.id),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} deleted.`,
        errorTitle: `Failed to delete ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to delete some sandboxes.',
        canceledTitle: 'Delete canceled.',
      },
    })

    if (selectedSandboxInBulk) {
      setShowSandboxDetails(false)
      setSelectedSandbox(null)
    }
  }

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number): Promise<string> => {
      setSandboxIsLoading((prev) => ({ ...prev, [sandboxId]: true }))
      try {
        return (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
      } finally {
        setSandboxIsLoading((prev) => ({ ...prev, [sandboxId]: false }))
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
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    toast.info('Checking VNC desktop status...')

    try {
      const statusResponse = await toolboxApi.getComputerUseStatusDeprecated(id, selectedOrganization?.id)
      const status = statusResponse.data.status

      if (status === 'active') {
        const vncUrl = await getVncUrl(id)
        if (vncUrl) {
          window.open(vncUrl, '_blank')
          toast.success('Opening VNC desktop...')
        }
      } else {
        try {
          await toolboxApi.startComputerUseDeprecated(id, selectedOrganization?.id)
          toast.success('Starting VNC desktop...')

          await new Promise((resolve) => setTimeout(resolve, 5000))

          try {
            const newStatusResponse = await toolboxApi.getComputerUseStatusDeprecated(id, selectedOrganization?.id)
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
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
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

  const handleScreenRecordings = async (id: string) => {
    const sandbox = sandboxItems?.find((s) => s.id === id)
    if (!sandbox || sandbox.state !== SandboxState.STARTED) {
      toast.error('Sandbox must be started to access Screen Recordings')
      return
    }

    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    try {
      const portPreviewUrl = await getPortPreviewUrl(id, 33333)
      window.open(portPreviewUrl, '_blank')
      toast.success('Opening Screen Recordings dashboard...')
    } catch (error) {
      handleApiError(error, 'Failed to open Screen Recordings')
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleCreateSshAccess = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    try {
      const response = await sandboxApi.createSshAccess(id, selectedOrganization?.id, sshExpiryMinutes)
      setSshAccess(response.data)
      setSshSandboxId(id)
      setShowCreateSshDialog(true)
      toast.success('SSH access created successfully')
    } catch (error) {
      handleApiError(error, 'Failed to create SSH access')
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
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

    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    try {
      await sandboxApi.revokeSshAccess(id, selectedOrganization?.id, revokeSshToken)
      setRevokeSshToken('')
      setSshSandboxId('')
      setShowRevokeSshDialog(false)
      toast.success('SSH access revoked successfully')
    } catch (error) {
      handleApiError(error, 'Failed to revoke SSH access')
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
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
    <PageLayout>
      <PageHeader>
        <PageTitle>Sandboxes</PageTitle>
        <div className="flex items-center gap-2 ml-auto">
          {!sandboxesDataIsLoading && (!sandboxItems || sandboxItems.length === 0) && (
            <>
              <Button variant="link" className="text-primary" onClick={() => navigate(RoutePath.ONBOARDING)} size="sm">
                Onboarding guide
              </Button>
              <Button variant="link" className="text-primary" asChild size="sm">
                <a href={DAYTONA_DOCS_URL} target="_blank" rel="noopener noreferrer" className="text-primary">
                  Docs
                </a>
              </Button>
            </>
          )}
        </div>
      </PageHeader>
      <PageContent size="full" className="flex-1 max-h-[calc(100vh-65px)]">
        <SandboxTable
          sandboxIsLoading={sandboxIsLoading}
          sandboxStateIsTransitioning={sandboxStateIsTransitioning}
          handleStart={handleStart}
          handleStop={handleStop}
          handleDelete={(id: string) => {
            setSandboxToDelete(id)
            setShowDeleteDialog(true)
          }}
          handleBulkDelete={handleBulkDelete}
          handleBulkStart={handleBulkStart}
          handleBulkStop={handleBulkStop}
          handleBulkArchive={handleBulkArchive}
          handleArchive={handleArchive}
          handleVnc={handleVnc}
          getWebTerminalUrl={getWebTerminalUrl}
          handleCreateSshAccess={openCreateSshDialog}
          handleRevokeSshAccess={openRevokeSshDialog}
          handleRefresh={handleRefresh}
          isRefreshing={sandboxDataIsRefreshing}
          data={sandboxItems || []}
          loading={sandboxesDataIsLoading}
          snapshots={snapshotsData?.items || []}
          snapshotsDataIsLoading={snapshotsDataIsLoading}
          snapshotsDataHasMore={snapshotsDataHasMore}
          onChangeSnapshotSearchValue={(name?: string) => handleSnapshotFiltersChange({ name })}
          regionsData={regionsData || []}
          regionsDataIsLoading={regionsDataIsLoading}
          onRowClick={(sandbox: Sandbox) => {
            setSelectedSandbox(sandbox)
            setShowSandboxDetails(true)
          }}
          pageCount={pageCount}
          totalItems={totalItems}
          onPaginationChange={handlePaginationChange}
          pagination={pagination}
          sorting={sorting}
          onSortingChange={handleSortingChange}
          filters={filters}
          onFiltersChange={handleFiltersChange}
          handleRecover={handleRecover}
          getRegionName={getRegionName}
          handleScreenRecordings={handleScreenRecordings}
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
                  variant="destructive"
                  onClick={() => handleDelete(sandboxToDelete)}
                  disabled={sandboxIsLoading[sandboxToDelete]}
                >
                  {sandboxIsLoading[sandboxToDelete] ? 'Deleting...' : 'Delete'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}

        <AlertDialog
          open={showCreateSshDialog}
          onOpenChange={(isOpen) => {
            setShowCreateSshDialog(isOpen)
            if (!isOpen) {
              setSshAccess(null)
              setSshExpiryMinutes(60)
              setSshSandboxId('')
            }
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Create SSH Access</AlertDialogTitle>
              <AlertDialogDescription>
                {sshAccess
                  ? 'SSH access has been created successfully. Use the token below to connect:'
                  : 'Set the expiration time for SSH access:'}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <div className="space-y-4">
              {!sshAccess ? (
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
                  <span className="overflow-x-auto pr-2 cursor-text select-all">{sshAccess.sshCommand}</span>
                  {(copied === 'SSH Command' && <Check className="w-4 h-4" />) || (
                    <Copy
                      className="w-4 h-4 cursor-pointer"
                      onClick={() => copyToClipboard(sshAccess.sshCommand, 'SSH Command')}
                    />
                  )}
                </div>
              )}
            </div>
            <AlertDialogFooter>
              {!sshAccess ? (
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
          sandboxIsLoading={sandboxIsLoading}
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
          handleRecover={handleRecover}
          getRegionName={getRegionName}
        />
      </PageContent>
    </PageLayout>
  )
}

const SandboxesPage: React.FC = () => (
  <SandboxListProvider>
    <Sandboxes />
  </SandboxListProvider>
)

export default SandboxesPage
