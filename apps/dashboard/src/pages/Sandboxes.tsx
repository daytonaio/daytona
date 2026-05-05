/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationSuspendedError } from '@/api/errors'
import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { ForkTreeDialog } from '@/components/ForkTreeDialog'
import { PageContent, PageFooter, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { RecursiveDeleteDialog } from '@/components/RecursiveDeleteDialog'
import { CreateSandboxSheet } from '@/components/Sandbox/CreateSandboxSheet'
import { tabParser } from '@/components/sandboxes/SearchParams'
import { CreateSshAccessSheet } from '@/components/sandboxes/CreateSshAccessSheet'
import { RevokeSshAccessDialog } from '@/components/sandboxes/RevokeSshAccessDialog'
import SandboxDetailsSheet, { type SandboxDetailsSheetTabValue } from '@/components/sandboxes/SandboxDetailsSheet'
import { SandboxTable } from '@/components/SandboxTable'
import type { SandboxTableRef } from '@/components/SandboxTable/types'
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
import { Input } from '@/components/ui/input'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { formatDuration, pluralize } from '@/lib/utils'
import {
  OrganizationRolePermissionsEnum,
  OrganizationUserRoleEnum,
  Sandbox,
  SandboxDesiredState,
  SandboxState,
  SnapshotDto,
} from '@daytona/api-client'
import type { Sandbox as CreatedSandbox } from '@daytona/sdk'
import { PlusIcon } from 'lucide-react'
import { parseAsString, useQueryState } from 'nuqs'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi, snapshotApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()
  const config = useConfig()

  const [sandboxes, setSandboxes] = useState<Sandbox[]>([])
  const [snapshots, setSnapshots] = useState<SnapshotDto[]>([])
  const [loadingSandboxes, setLoadingSandboxes] = useState<Record<string, boolean>>({})
  const [transitioningSandboxes, setTransitioningSandboxes] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [loadingSnapshots, setLoadingSnapshots] = useState(true)
  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const [forkTreeSandboxId, setForkTreeSandboxId] = useState<string | null>(null)
  const [recursiveDeleteSandboxId, setRecursiveDeleteSandboxId] = useState<string | null>(null)

  // Create Snapshot Dialog

  const [sandboxToSnapshot, setSandboxToSnapshot] = useState<string | null>(null)
  const [snapshotName, setSnapshotName] = useState('')
  const [snapshotIsLoading, setSnapshotIsLoading] = useState(false)

  const handleCreateSnapshot = (id: string) => {
    setSandboxToSnapshot(id)
    setSnapshotName('')
  }

  const handleFork = async (id: string) => {
    try {
      await sandboxApi.forkSandbox(id, {}, selectedOrganization?.id)
      toast.success('Fork started')
      await fetchSandboxes()
    } catch {
      toast.error('Failed to fork sandbox')
    }
  }

  const handleViewForks = (id: string) => {
    setForkTreeSandboxId(id)
  }

  const openDeleteDialog = async (id: string) => {
    try {
      const forksRes = await sandboxApi.getSandboxForks(id, selectedOrganization?.id)
      if (forksRes.data.length > 0) {
        setRecursiveDeleteSandboxId(id)
        return
      }
    } catch {
      // Fall through to normal delete if fork check fails
    }
    setSandboxToDelete(id)
    setShowDeleteDialog(true)
  }

  // Sandbox Details Drawer

  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [orderedSandboxItems, setOrderedSandboxItems] = useState<Sandbox[] | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)
  const [sandboxDetailsInitialTab, setSandboxDetailsInitialTab] = useState<SandboxDetailsSheetTabValue>('overview')
  const [sandboxIdParam, setSandboxIdParam] = useQueryState('sandboxId', parseAsString)
  const [sandboxTabParam, setSandboxTabParam] = useQueryState('tab', tabParser)
  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshSandboxId, setSshSandboxId] = useState<string>('')
  const createSandboxSheetRef = useRef<{ open: () => void }>(null)
  const sandboxTableRef = useRef<SandboxTableRef>(null)

  // Region Filter

  const { availableRegions: regionsData, loadingAvailableRegions: regionsDataIsLoading, getRegionName } = useRegions()

  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember, authenticatedUserHasPermission } =
    useSelectedOrganization()

  const fetchSnapshots = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setLoadingSnapshots(true)
    try {
      // TODO: Implement snapshot search by input
      // e.g. "Search to load more results"
      const response = await snapshotApi.getAllSnapshots(selectedOrganization.id)
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
    if (!sandboxIdParam) {
      setShowSandboxDetails(false)
      return
    }

    const sandboxFromUrl = sandboxes.find((sandbox) => sandbox.id === sandboxIdParam)
    if (!sandboxFromUrl) {
      return
    }

    setSelectedSandbox(sandboxFromUrl)
    if (!showSandboxDetails || selectedSandbox?.id !== sandboxFromUrl.id) {
      setSandboxDetailsInitialTab(sandboxTabParam as SandboxDetailsSheetTabValue)
    }
    setShowSandboxDetails(true)
  }, [sandboxIdParam, sandboxTabParam, sandboxes, selectedSandbox?.id, showSandboxDetails])

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
      setSandboxIdParam(null)
      setSandboxTabParam(null)
    }
  }, [sandboxes, selectedSandbox, setSandboxIdParam, setSandboxTabParam])

  useEffect(() => {
    const handleSandboxCreatedEvent = (sandbox: Sandbox) => {
      setSandboxes((prev) =>
        prev.some((existingSandbox) => existingSandbox.id === sandbox.id) ? prev : [sandbox, ...prev],
      )
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
      handleApiError(error, 'Failed to start sandbox', {
        action:
          error instanceof OrganizationSuspendedError &&
          config.billingApiUrl &&
          authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER ? (
            <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING_WALLET)}>
              Go to billing
            </Button>
          ) : null,
      })
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

  const handleRecover = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    setTransitioningSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToRecover = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToRecover?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STARTING } : s)))

    try {
      await sandboxApi.recoverSandbox(id, selectedOrganization?.id)
      toast.success('Sandbox recovered. Restarting...')
    } catch (error) {
      handleApiError(error, 'Failed to recover sandbox')
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
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
      await sandboxApi.deleteSandbox(id, selectedOrganization?.id)
      setSandboxToDelete(null)
      setShowDeleteDialog(false)

      if (selectedSandbox?.id === id) {
        setShowSandboxDetails(false)
        setSelectedSandbox(null)
        setSandboxIdParam(null)
        setSandboxTabParam(null)
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

  const performSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      if (selectedSandbox?.id === sandboxId) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: newState } : null))
      }
    },
    [selectedSandbox?.id],
  )

  const revertSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      if (!previousState) {
        return
      }

      if (selectedSandbox?.id === sandboxId) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    },
    [selectedSandbox?.id],
  )

  // todo(rpavlini): we should refactor this and move to react-query mutations
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
      const previousStatesById = new Map((sandboxes ?? []).map((sandbox) => [sandbox.id, sandbox.state]))

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

          setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
          // setSandboxIsLoading((prev) => ({ ...prev, [id]: true }))
          setTransitioningSandboxes((prev) => ({ ...prev, [id]: true }))
          // setSandboxStateIsTransitioning((prev) => ({ ...prev, [id]: true }))
          performSandboxStateOptimisticUpdate(id, optimisticState)

          try {
            await apiCall(id)
            successCount += 1
          } catch (error) {
            failureCount += 1
            revertSandboxStateOptimisticUpdate(id, previousStatesById.get(id))
            console.error(`${actionName} sandbox failed`, id, error)
          } finally {
            setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
            setTimeout(() => {
              setTransitioningSandboxes((prev) => ({ ...prev, [id]: false }))
            }, 2000)
          }
        }

        bulkToast.result({ successCount, failureCount }, toastMessages)
      } catch (error) {
        console.error(`${actionName} sandboxes failed`, error)
        bulkToast.error(`${actionName} sandboxes failed.`)
      }

      return { successCount, failureCount }
    },
    [sandboxes, performSandboxStateOptimisticUpdate, revertSandboxStateOptimisticUpdate],
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
      setLoadingSandboxes((prev) => ({ ...prev, [sandboxId]: true }))
      try {
        return (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
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
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleScreenRecordings = async (id: string) => {
    // Check if sandbox is started
    const sandbox = sandboxes.find((s) => s.id === id)
    if (!sandbox || sandbox.state !== SandboxState.STARTED) {
      toast.error('Sandbox must be started to access Screen Recordings')
      return
    }

    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))
    try {
      const portPreviewUrl = await getPortPreviewUrl(id, 33333)
      window.open(portPreviewUrl, '_blank')
      toast.success('Opening Screen Recordings dashboard...')
    } catch (error) {
      handleApiError(error, 'Failed to open Screen Recordings')
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
    }
  }

  const openCreateSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowCreateSshDialog(true)
  }

  const openRevokeSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowRevokeSshDialog(true)
  }

  const sandboxItems = useMemo(() => orderedSandboxItems ?? sandboxes ?? [], [orderedSandboxItems, sandboxes])
  const selectedSandboxIndex = useMemo(
    () => sandboxItems.findIndex((sandbox) => sandbox.id === selectedSandbox?.id),
    [sandboxItems, selectedSandbox?.id],
  )

  const handleSandboxSheetNavigate = (direction: 'prev' | 'next') => {
    if (selectedSandboxIndex < 0) {
      return
    }

    const nextIndex = direction === 'prev' ? selectedSandboxIndex - 1 : selectedSandboxIndex + 1
    const nextSandbox = sandboxItems[nextIndex]

    if (nextSandbox) {
      setSelectedSandbox(nextSandbox)
      setSandboxIdParam(nextSandbox.id)
    }
  }

  const handleSandboxDetailsOpenChange = (isOpen: boolean) => {
    setShowSandboxDetails(isOpen)

    if (!isOpen) {
      setSandboxIdParam(null)
      setSandboxTabParam(null)
    }
  }

  const openSandboxDetails = (sandbox: Sandbox, initialTab: SandboxDetailsSheetTabValue = 'overview') => {
    const orderedSandboxes =
      sandboxTableRef.current?.table.getPrePaginationRowModel().rows.map((row) => row.original) ?? []
    setOrderedSandboxItems(orderedSandboxes.some((item) => item.id === sandbox.id) ? orderedSandboxes : null)
    setSelectedSandbox(sandbox)
    setSandboxDetailsInitialTab(initialTab)
    setSandboxIdParam(sandbox.id)
    setSandboxTabParam(initialTab)
    setShowSandboxDetails(true)
  }

  const handleSandboxRowClick = (sandbox: Sandbox) => {
    openSandboxDetails(sandbox)
  }

  const handleOpenTerminal = (sandbox: Sandbox) => {
    openSandboxDetails(sandbox, 'terminal')
  }

  const handleSandboxCreated = (sandbox: CreatedSandbox) => {
    const createdSandbox = sandbox as unknown as Sandbox

    setSandboxes((prev) =>
      prev.some((existingSandbox) => existingSandbox.id === createdSandbox.id)
        ? prev.map((existingSandbox) => (existingSandbox.id === createdSandbox.id ? createdSandbox : existingSandbox))
        : [createdSandbox, ...prev],
    )
    openSandboxDetails(createdSandbox)
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES),
    [authenticatedUserHasPermission],
  )
  const canCreateSandbox = writePermitted && !selectedOrganization?.suspended

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!canCreateSandbox) {
      return []
    }

    return [
      {
        id: 'create-sandbox',
        label: 'Create Sandbox',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => createSandboxSheetRef.current?.open(),
      },
    ]
  }, [canCreateSandbox])

  useRegisterCommands(rootCommands, { groupId: 'sandbox-actions', groupLabel: 'Sandbox actions', groupOrder: 0 })

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
    <PageLayout contained>
      <PageHeader>
        <PageTitle>Sandboxes</PageTitle>
        <div className="flex items-center gap-2 ml-auto">
          {!loadingTable && sandboxes.length === 0 && (
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
          {canCreateSandbox && (
            <CreateSandboxSheet ref={createSandboxSheetRef} onSandboxCreated={handleSandboxCreated} />
          )}
        </div>
      </PageHeader>
      <PageContent size="full" className="overflow-hidden">
        <SandboxTable
          ref={sandboxTableRef}
          sandboxIsLoading={loadingSandboxes}
          sandboxStateIsTransitioning={transitioningSandboxes}
          activeSandboxId={showSandboxDetails ? selectedSandbox?.id : undefined}
          handleStart={handleStart}
          handleStop={handleStop}
          handleDelete={openDeleteDialog}
          handleBulkDelete={handleBulkDelete}
          handleBulkStart={handleBulkStart}
          handleBulkStop={handleBulkStop}
          handleBulkArchive={handleBulkArchive}
          handleArchive={handleArchive}
          handleVnc={handleVnc}
          handleCreateSshAccess={openCreateSshDialog}
          handleRevokeSshAccess={openRevokeSshDialog}
          data={sandboxes}
          loading={loadingTable}
          snapshots={snapshots}
          onRowClick={handleSandboxRowClick}
          loadingSnapshots={loadingSnapshots}
          regionsData={regionsData}
          regionsDataIsLoading={regionsDataIsLoading}
          handleRecover={handleRecover}
          getRegionName={getRegionName}
          handleScreenRecordings={handleScreenRecordings}
          handleCreateSnapshot={handleCreateSnapshot}
          handleFork={handleFork}
          handleViewForks={handleViewForks}
          handleOpenTerminal={handleOpenTerminal}
        />

        {sandboxToDelete && (
          <AlertDialog
            open={showDeleteDialog}
            onOpenChange={(isOpen) => {
              if (!isOpen && loadingSandboxes[sandboxToDelete]) {
                return
              }

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
                <AlertDialogCancel disabled={loadingSandboxes[sandboxToDelete]}>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  onClick={(event) => {
                    event.preventDefault()
                    handleDelete(sandboxToDelete)
                  }}
                  disabled={loadingSandboxes[sandboxToDelete]}
                >
                  {loadingSandboxes[sandboxToDelete] ? 'Deleting...' : 'Delete'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}

        {sandboxToSnapshot && (
          <AlertDialog
            open={!!sandboxToSnapshot}
            onOpenChange={(isOpen) => {
              if (!isOpen) {
                setSandboxToSnapshot(null)
                setSnapshotName('')
              }
            }}
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Create Snapshot</AlertDialogTitle>
                <AlertDialogDescription>Enter a name for the new snapshot.</AlertDialogDescription>
              </AlertDialogHeader>
              <Input
                value={snapshotName}
                onChange={(e) => setSnapshotName(e.target.value)}
                placeholder="Snapshot name"
                disabled={snapshotIsLoading}
              />
              <AlertDialogFooter>
                <AlertDialogCancel disabled={snapshotIsLoading}>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  disabled={!snapshotName.trim() || snapshotIsLoading}
                  onClick={async (e) => {
                    e.preventDefault()
                    if (!sandboxToSnapshot || !snapshotName.trim()) return
                    setSnapshotIsLoading(true)
                    try {
                      await sandboxApi.createSandboxSnapshot(
                        sandboxToSnapshot,
                        { name: snapshotName.trim() },
                        selectedOrganization?.id,
                      )
                      toast.success('Snapshot creation started')
                      setSandboxToSnapshot(null)
                      setSnapshotName('')
                    } catch (error) {
                      handleApiError(error, 'Failed to create snapshot')
                    } finally {
                      setSnapshotIsLoading(false)
                    }
                  }}
                >
                  {snapshotIsLoading ? 'Creating...' : 'Create'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}

        <CreateSshAccessSheet
          sandboxId={sshSandboxId}
          open={showCreateSshDialog}
          onOpenChange={(isOpen) => {
            setShowCreateSshDialog(isOpen)
            if (!isOpen) {
              setSshSandboxId('')
            }
          }}
        />

        <RevokeSshAccessDialog
          sandboxId={sshSandboxId}
          open={showRevokeSshDialog}
          onOpenChange={(isOpen) => {
            setShowRevokeSshDialog(isOpen)
            if (!isOpen) {
              setSshSandboxId('')
            }
          }}
        />

        <SandboxDetailsSheet
          sandbox={selectedSandbox}
          open={showSandboxDetails}
          onOpenChange={handleSandboxDetailsOpenChange}
          sandboxIsLoading={loadingSandboxes}
          handleStart={handleStart}
          handleStop={handleStop}
          handleDelete={async (id) => {
            await openDeleteDialog(id)
          }}
          handleArchive={handleArchive}
          writePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
          deletePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
          handleRecover={handleRecover}
          getRegionName={getRegionName}
          onCreateSshAccess={openCreateSshDialog}
          onRevokeSshAccess={openRevokeSshDialog}
          onScreenRecordings={handleScreenRecordings}
          onNavigate={handleSandboxSheetNavigate}
          hasPrev={selectedSandboxIndex > 0}
          hasNext={selectedSandboxIndex >= 0 && selectedSandboxIndex < sandboxItems.length - 1}
          initialTab={sandboxDetailsInitialTab}
          activeTab={sandboxTabParam as SandboxDetailsSheetTabValue}
          onTabChange={setSandboxTabParam}
        />

        {forkTreeSandboxId && (
          <ForkTreeDialog
            sandboxId={forkTreeSandboxId}
            open={!!forkTreeSandboxId}
            onClose={() => setForkTreeSandboxId(null)}
          />
        )}

        {recursiveDeleteSandboxId && (
          <RecursiveDeleteDialog
            sandboxId={recursiveDeleteSandboxId}
            open={!!recursiveDeleteSandboxId}
            onClose={() => setRecursiveDeleteSandboxId(null)}
            onDeleted={async () => {
              await fetchSandboxes()
            }}
          />
        )}
      </PageContent>
      <PageFooter />
    </PageLayout>
  )
}

export default Sandboxes
