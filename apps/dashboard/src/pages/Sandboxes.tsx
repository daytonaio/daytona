/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState, useCallback, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { OrganizationSuspendedError } from '@/api/errors'
import { OrganizationUserRoleEnum, Sandbox, SandboxDesiredState, SandboxState } from '@daytonaio/api-client'
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
import { Label } from '@/components/ui/label'
import { Check, Copy } from 'lucide-react'
import { useConfig } from '@/hooks/useConfig'
import { QueryKey, useQueryClient } from '@tanstack/react-query'
import {
  getSandboxesQueryKey,
  SandboxFilters,
  SandboxSorting,
  SandboxQueryParams,
  useSandboxes,
  DEFAULT_SANDBOX_SORTING,
} from '@/hooks/useSandboxes'
import { getRegionsQueryKey, useRegions } from '@/hooks/useRegions'
import { getSnapshotsQueryKey, SnapshotFilters, SnapshotQueryParams, useSnapshots } from '@/hooks/useSnapshots'

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi } = useApi()
  const { user } = useAuth()
  const navigate = useNavigate()
  const { notificationSocket } = useNotificationSocket()
  const config = useConfig()
  const queryClient = useQueryClient()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  // Pagination

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  // Filters

  const [filters, setFilters] = useState<SandboxFilters>({})

  const handleFiltersChange = useCallback((filters: SandboxFilters) => {
    setFilters(filters)
    setPaginationParams((prev) => ({ ...prev, pageIndex: 0 }))
  }, [])

  // Sorting

  const [sorting, setSorting] = useState<SandboxSorting>(DEFAULT_SANDBOX_SORTING)

  const handleSortingChange = useCallback((sorting: SandboxSorting) => {
    setSorting(sorting)
    setPaginationParams((prev) => ({ ...prev, pageIndex: 0 }))
  }, [])

  // Sandboxes Data

  const queryParams = useMemo<SandboxQueryParams>(
    () => ({
      page: paginationParams.pageIndex + 1, // 1-indexed
      pageSize: paginationParams.pageSize,
      filters: filters,
      sorting: sorting,
    }),
    [paginationParams, filters, sorting],
  )

  const baseQueryKey = useMemo<QueryKey>(
    () => getSandboxesQueryKey(selectedOrganization?.id),
    [selectedOrganization?.id],
  )

  const queryKey = useMemo<QueryKey>(
    () => getSandboxesQueryKey(selectedOrganization?.id, queryParams),
    [selectedOrganization?.id, queryParams],
  )

  const {
    data: sandboxesData,
    isLoading: sandboxesDataIsLoading,
    error: sandboxesDataError,
    refetch: refetchSandboxesData,
  } = useSandboxes(queryKey, queryParams)

  useEffect(() => {
    if (sandboxesDataError) {
      handleApiError(sandboxesDataError, 'Failed to fetch sandboxes')
    }
  }, [sandboxesDataError])

  const updateSandboxInCache = useCallback(
    (sandboxId: string, updates: Partial<Sandbox>) => {
      queryClient.setQueryData(queryKey, (oldData: any) => {
        if (!oldData) return oldData
        return {
          ...oldData,
          items: oldData.items.map((sandbox: Sandbox) =>
            sandbox.id === sandboxId ? { ...sandbox, ...updates } : sandbox,
          ),
        }
      })
    },
    [queryClient, queryKey],
  )

  /**
   * Marks all sandbox queries for this organization as stale.
   *
   * Useful when sandbox event occurs and we don't have a good way of knowing for which combination of query parameters the sandbox would be shown.
   *
   * @param shouldRefetchActiveQueries If true, only active queries will be refetched. Otherwise, no queries will be refetched.
   */
  const markAllSandboxQueriesAsStale = useCallback(
    async (shouldRefetchActiveQueries = false) => {
      queryClient.invalidateQueries({
        queryKey: baseQueryKey,
        refetchType: shouldRefetchActiveQueries ? 'active' : 'none',
      })
    },
    [queryClient, baseQueryKey],
  )

  /**
   * Aborts all outgoing refetches for the provided key.
   *
   * Useful for preventing refetches from overwriting optimistic updates.
   *
   * @param queryKey
   */
  const cancelQueryRefetches = useCallback(
    async (queryKey: QueryKey) => {
      queryClient.cancelQueries({ queryKey })
    },
    [queryClient],
  )

  // Go to previous page if there are no items on the current page

  useEffect(() => {
    if (sandboxesData?.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [sandboxesData?.items.length, paginationParams.pageIndex])

  // Ephemeral Sandbox States

  const [sandboxIsLoading, setSandboxIsLoading] = useState<Record<string, boolean>>({})
  const [sandboxStateIsTransitioning, setSandboxStateIsTransitioning] = useState<Record<string, boolean>>({}) // display transition animation

  // Manual Refreshing

  const [sandboxDataIsRefreshing, setSandboxDataIsRefreshing] = useState(false)

  const handleRefresh = useCallback(async () => {
    setSandboxDataIsRefreshing(true)
    try {
      await refetchSandboxesData()
    } catch (error) {
      handleApiError(error, 'Failed to refresh sandboxes')
    } finally {
      setSandboxDataIsRefreshing(false)
    }
  }, [refetchSandboxesData])

  // Delete Sandbox Dialog

  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  // Sandbox Details Drawer

  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)

  useEffect(() => {
    if (!selectedSandbox || !sandboxesData?.items) {
      return
    }

    const selectedSandboxInData = sandboxesData.items.find((s) => s.id === selectedSandbox.id)

    if (!selectedSandboxInData) {
      setSelectedSandbox(null)
      setShowSandboxDetails(false)
      return
    }

    if (selectedSandboxInData !== selectedSandbox) {
      setSelectedSandbox(selectedSandboxInData)
    }
  }, [sandboxesData?.items, selectedSandbox])

  const performSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      updateSandboxInCache(sandboxId, { state: newState })

      if (selectedSandbox?.id === sandboxId) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: newState } : null))
      }
    },
    [updateSandboxInCache, selectedSandbox?.id],
  )

  const revertSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      if (!previousState) {
        return
      }

      updateSandboxInCache(sandboxId, { state: previousState })

      if (selectedSandbox?.id === sandboxId) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    },
    [updateSandboxInCache, selectedSandbox?.id],
  )

  // SSH Access Dialogs

  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshToken, setSshToken] = useState<string>('')
  const [sshExpiryMinutes, setSshExpiryMinutes] = useState<number>(60)
  const [revokeSshToken, setRevokeSshToken] = useState<string>('')
  const [sshSandboxId, setSshSandboxId] = useState<string>('')
  const [copied, setCopied] = useState<string | null>(null)

  // Snapshot Filter

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

  const snapshotsQueryKey = useMemo<QueryKey>(
    () => getSnapshotsQueryKey(selectedOrganization?.id, snapshotsQueryParams),
    [selectedOrganization?.id, snapshotsQueryParams],
  )

  const {
    data: snapshotsData,
    isLoading: snapshotsDataIsLoading,
    error: snapshotsDataError,
  } = useSnapshots(snapshotsQueryKey, snapshotsQueryParams)

  const snapshotsDataHasMore = useMemo(() => {
    return snapshotsData && snapshotsData.totalPages > 1
  }, [snapshotsData])

  useEffect(() => {
    if (snapshotsDataError) {
      handleApiError(snapshotsDataError, 'Failed to fetch snapshots')
    }
  }, [snapshotsDataError])

  // Region Filter

  const regionsQueryKey = getRegionsQueryKey(selectedOrganization?.id)

  const { data: regionsData, isLoading: regionsDataIsLoading, error: regionsDataError } = useRegions(regionsQueryKey)

  useEffect(() => {
    if (regionsDataError) {
      handleApiError(regionsDataError, 'Failed to fetch sandboxes regions')
    }
  }, [regionsDataError])

  /**
   * Marks all regions queries for this organization as stale.
   *
   * Useful when a sandbox is created, potentially in a completely new region.
   *
   */
  const markAllRegionsQueriesAsStale = useCallback(async () => {
    queryClient.invalidateQueries({
      queryKey: regionsQueryKey,
    })
  }, [queryClient, regionsQueryKey])

  // Subscribe to Sandbox Events

  useEffect(() => {
    const handleSandboxCreatedEvent = (sandbox: Sandbox) => {
      const isFirstPage = paginationParams.pageIndex === 0
      const isDefaultFilters = Object.keys(filters).length === 0
      const isDefaultSorting =
        sorting.field === DEFAULT_SANDBOX_SORTING.field && sorting.direction === DEFAULT_SANDBOX_SORTING.direction

      const shouldRefetchActiveQueries = isFirstPage && isDefaultFilters && isDefaultSorting

      markAllSandboxQueriesAsStale(shouldRefetchActiveQueries)
      markAllRegionsQueriesAsStale()
    }

    const handleSandboxStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldState: SandboxState
      newState: SandboxState
    }) => {
      // warm pool sandboxes
      if (data.oldState === data.newState && data.newState === SandboxState.STARTED) {
        handleSandboxCreatedEvent(data.sandbox)
        return
      }

      let updatedState = data.newState

      // error,build_failed | destroyed should be displayed as destroyed in the UI
      if (
        data.sandbox.desiredState === SandboxDesiredState.DESTROYED &&
        (data.newState === SandboxState.ERROR || data.newState === SandboxState.BUILD_FAILED)
      ) {
        updatedState = SandboxState.DESTROYED
      }

      performSandboxStateOptimisticUpdate(data.sandbox.id, updatedState)

      markAllSandboxQueriesAsStale()
    }

    const handleSandboxDesiredStateUpdatedEvent = (data: {
      sandbox: Sandbox
      oldDesiredState: SandboxDesiredState
      newDesiredState: SandboxDesiredState
    }) => {
      // error,build_failed | destroyed should be displayed as destroyed in the UI

      if (data.newDesiredState !== SandboxDesiredState.DESTROYED) {
        return
      }

      if (data.sandbox.state !== SandboxState.ERROR && data.sandbox.state !== SandboxState.BUILD_FAILED) {
        return
      }

      performSandboxStateOptimisticUpdate(data.sandbox.id, SandboxState.DESTROYED)

      markAllSandboxQueriesAsStale()
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
    markAllRegionsQueriesAsStale,
    markAllSandboxQueriesAsStale,
    notificationSocket,
    paginationParams.pageIndex,
    performSandboxStateOptimisticUpdate,
    sorting.direction,
    sorting.field,
  ])

  // Sandbox Action Handlers

  const handleStart = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToStart = sandboxesData?.items.find((s) => s.id === id)
    const previousState = sandboxToStart?.state

    await cancelQueryRefetches(queryKey)
    performSandboxStateOptimisticUpdate(id, SandboxState.STARTING)

    try {
      await sandboxApi.startSandbox(id, selectedOrganization?.id)
      toast.success(`Starting sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
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
      revertSandboxStateOptimisticUpdate(id, previousState)
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

    const sandboxToStop = sandboxesData?.items.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    await cancelQueryRefetches(queryKey)
    performSandboxStateOptimisticUpdate(id, SandboxState.STOPPING)

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
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
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

    const sandboxToDelete = sandboxesData?.items.find((s) => s.id === id)
    const previousState = sandboxToDelete?.state

    await cancelQueryRefetches(queryKey)
    performSandboxStateOptimisticUpdate(id, SandboxState.DESTROYING)

    try {
      await sandboxApi.deleteSandbox(id, selectedOrganization?.id)
      setSandboxToDelete(null)
      setShowDeleteDialog(false)

      if (selectedSandbox?.id === id) {
        setShowSandboxDetails(false)
        setSelectedSandbox(null)
      }

      toast.success(`Deleting sandbox with ID:  ${id}`)

      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const handleBulkDelete = async (ids: string[]) => {
    setSandboxIsLoading((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))

    await cancelQueryRefetches(queryKey)

    const selectedSandboxInBulk = selectedSandbox && ids.includes(selectedSandbox.id)

    for (const id of ids) {
      const sandboxToDelete = sandboxesData?.items.find((s) => s.id === id)
      const previousState = sandboxToDelete?.state

      performSandboxStateOptimisticUpdate(id, SandboxState.DESTROYING)

      try {
        await sandboxApi.deleteSandbox(id, selectedOrganization?.id)
        toast.success(`Deleting sandbox with ID: ${id}`)
        await markAllSandboxQueriesAsStale()
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')

        revertSandboxStateOptimisticUpdate(id, previousState)

        const shouldContinue = window.confirm(
          `Failed to delete sandbox with ID: ${id}. Do you want to continue with the remaining sandboxes?`,
        )

        if (!shouldContinue) {
          break
        }
      } finally {
        setSandboxIsLoading((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: false }), {}) }))
        setTimeout(() => {
          setSandboxStateIsTransitioning((prev) => ({
            ...prev,
            ...ids.reduce((acc, id) => ({ ...acc, [id]: false }), {}),
          }))
        }, 2000)
      }
    }

    if (selectedSandboxInBulk) {
      setShowSandboxDetails(false)
      setSelectedSandbox(null)
    }
  }

  const handleArchive = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))

    const sandboxToArchive = sandboxesData?.items.find((s) => s.id === id)
    const previousState = sandboxToArchive?.state

    await cancelQueryRefetches(queryKey)
    performSandboxStateOptimisticUpdate(id, SandboxState.ARCHIVING)

    try {
      await sandboxApi.archiveSandbox(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    } finally {
      setSandboxIsLoading((prev) => ({ ...prev, [id]: false }))
      setTimeout(() => {
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: false }))
      }, 2000)
    }
  }

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number): Promise<string> => {
      setSandboxIsLoading((prev) => ({ ...prev, [sandboxId]: true }))
      try {
        return (await sandboxApi.getPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
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

    // Notify user immediately that we're checking VNC status
    toast.info('Checking VNC desktop status...')

    try {
      // First, check if computer use is already started
      const statusResponse = await toolboxApi.getComputerUseStatusDeprecated(id, selectedOrganization?.id)
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
          await toolboxApi.startComputerUseDeprecated(id, selectedOrganization?.id)
          toast.success('Starting VNC desktop...')

          // Wait a moment for processes to start, then open VNC
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

  const handleCreateSshAccess = async (id: string) => {
    setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
    try {
      const response = await sandboxApi.createSshAccess(id, selectedOrganization?.id, sshExpiryMinutes)
      setSshToken(response.data.token)
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
        {!sandboxesDataIsLoading && (!sandboxesData?.items || sandboxesData.items.length === 0) && (
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
        sandboxIsLoading={sandboxIsLoading}
        sandboxStateIsTransitioning={sandboxStateIsTransitioning}
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
        handleRefresh={handleRefresh}
        isRefreshing={sandboxDataIsRefreshing}
        data={sandboxesData?.items || []}
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
        pageCount={sandboxesData?.totalPages || 0}
        totalItems={sandboxesData?.total || 0}
        onPaginationChange={handlePaginationChange}
        pagination={{
          pageIndex: paginationParams.pageIndex,
          pageSize: paginationParams.pageSize,
        }}
        sorting={sorting}
        onSortingChange={handleSortingChange}
        filters={filters}
        onFiltersChange={handleFiltersChange}
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
                disabled={sandboxIsLoading[sandboxToDelete]}
              >
                {sandboxIsLoading[sandboxToDelete] ? 'Deleting...' : 'Delete'}
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
                  {config.sshGatewayCommand?.replace('{{TOKEN}}', sshToken) ||
                    `ssh -p 22222 user@host -o ProxyCommand="echo ${sshToken}"`}
                </span>
                {(copied === 'SSH Command' && <Check className="w-4 h-4" />) || (
                  <Copy
                    className="w-4 h-4 cursor-pointer"
                    onClick={() =>
                      copyToClipboard(
                        config.sshGatewayCommand?.replace('{{TOKEN}}', sshToken) ||
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
      />
    </div>
  )
}

export default Sandboxes
