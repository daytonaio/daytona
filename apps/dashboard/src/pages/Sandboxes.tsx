/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationSuspendedError } from '@/api/errors'
import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { CreateSandboxSheet } from '@/components/Sandbox/CreateSandboxSheet'
import SandboxDetailsSheet from '@/components/SandboxDetailsSheet'
import { SandboxTable } from '@/components/SandboxTable'
import { CreateSshAccessSheet } from '@/components/sandboxes/CreateSshAccessSheet'
import { RevokeSshAccessDialog } from '@/components/sandboxes/RevokeSshAccessDialog'
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
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { mutationKeys, type SandboxMutationVariables } from '@/hooks/mutations/mutationKeys'
import { SnapshotFilters, SnapshotQueryParams, useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { usePendingMutationKeys } from '@/hooks/usePendingMutationKeys'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { formatDuration, pluralize } from '@/lib/utils'
import { SandboxListProvider, useSandboxListContext } from '@/providers/SandboxListProvider'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum, Sandbox, SandboxState } from '@daytona/api-client'
import { AlertCircle, PlusIcon, RefreshCw } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

const sandboxPendingMutationSelectors = [
  {
    mutationKey: mutationKeys.sandboxes.start,
    getKey: (variables: SandboxMutationVariables | undefined) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.recover,
    getKey: (variables: SandboxMutationVariables | undefined) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.stop,
    getKey: (variables: SandboxMutationVariables | undefined) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.archive,
    getKey: (variables: SandboxMutationVariables | undefined) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.delete,
    getKey: (variables: SandboxMutationVariables | undefined) => variables?.sandboxId,
  },
] as const

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi } = useApi()
  const { user } = useAuth()
  const navigate = useNavigate()
  const config = useConfig()
  const { selectedOrganization, authenticatedUserOrganizationMember, authenticatedUserHasPermission } =
    useSelectedOrganization()

  const {
    sandboxes: sandboxItems,
    totalItems,
    pageCount,
    isLoading: sandboxesDataIsLoading,
    isRefetching: sandboxesDataIsRefetching,
    error: sandboxesDataError,
    pagination,
    onPaginationChange: handlePaginationChange,
    sorting,
    onSortingChange: handleSortingChange,
    filters,
    onFiltersChange: handleFiltersChange,
    handleRefresh,
    isRefreshing: sandboxDataIsRefreshing,
    startSandbox,
    recoverSandbox,
    stopSandbox,
    archiveSandbox,
    deleteSandbox,
  } = useSandboxListContext()

  const pendingSandboxMutationIds = usePendingMutationKeys(sandboxPendingMutationSelectors)
  const [sandboxAsyncOperationIsLoading, setSandboxAsyncOperationIsLoading] = useState<Record<string, boolean>>({})
  const [sandboxStateIsTransitioning, setSandboxStateIsTransitioning] = useState<Record<string, boolean>>({})
  const sandboxTransitionTimeoutsRef = useRef<Record<string, number>>({})

  const sandboxIsLoading = useMemo(() => {
    const combined = { ...sandboxAsyncOperationIsLoading }

    for (const sandboxId of pendingSandboxMutationIds) {
      combined[sandboxId] = true
    }

    return combined
  }, [sandboxAsyncOperationIsLoading, pendingSandboxMutationIds])

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

  const clearSandboxTransitionTimeout = useCallback((sandboxId: string) => {
    const timeoutId = sandboxTransitionTimeoutsRef.current[sandboxId]
    if (timeoutId === undefined) {
      return
    }

    window.clearTimeout(timeoutId)
    delete sandboxTransitionTimeoutsRef.current[sandboxId]
  }, [])

  useEffect(() => {
    return () => {
      Object.values(sandboxTransitionTimeoutsRef.current).forEach((timeoutId) => {
        window.clearTimeout(timeoutId)
      })
    }
  }, [])

  const startSandboxTransition = useCallback(
    (sandboxId: string) => {
      clearSandboxTransitionTimeout(sandboxId)
      setSandboxStateIsTransitioning((prev) => ({ ...prev, [sandboxId]: true }))
    },
    [clearSandboxTransitionTimeout],
  )

  const stopSandboxTransition = useCallback(
    (sandboxId: string) => {
      clearSandboxTransitionTimeout(sandboxId)

      sandboxTransitionTimeoutsRef.current[sandboxId] = window.setTimeout(() => {
        delete sandboxTransitionTimeoutsRef.current[sandboxId]
        setSandboxStateIsTransitioning((prev) => ({ ...prev, [sandboxId]: false }))
      }, 2000)
    },
    [clearSandboxTransitionTimeout],
  )

  const runSandboxTransition = useCallback(
    async (sandboxId: string, action: () => Promise<void>) => {
      startSandboxTransition(sandboxId)
      try {
        await action()
      } finally {
        stopSandboxTransition(sandboxId)
      }
    },
    [startSandboxTransition, stopSandboxTransition],
  )

  const setSandboxAsyncLoading = useCallback((sandboxId: string, isLoading: boolean) => {
    setSandboxAsyncOperationIsLoading((prev) => ({ ...prev, [sandboxId]: isLoading }))
  }, [])

  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshSandboxId, setSshSandboxId] = useState<string>('')
  const createSandboxSheetRef = useRef<{ open: () => void }>(null)

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

  const handleStart = async (id: string) => {
    await runSandboxTransition(id, async () => {
      try {
        await startSandbox(id)
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
            ) : undefined,
        })
      }
    })
  }

  const handleRecover = async (id: string) => {
    await runSandboxTransition(id, async () => {
      try {
        await recoverSandbox(id)
        toast.success('Sandbox recovered. Restarting...')
      } catch (error) {
        handleApiError(error, 'Failed to recover sandbox')
      }
    })
  }

  const handleStop = async (id: string) => {
    const sandboxToStop = sandboxItems.find((s) => s.id === id)

    await runSandboxTransition(id, async () => {
      try {
        await stopSandbox(id)
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
      }
    })
  }

  const handleDelete = async (id: string) => {
    await runSandboxTransition(id, async () => {
      try {
        await deleteSandbox(id)
        setSandboxToDelete(null)
        setShowDeleteDialog(false)

        if (selectedSandbox?.id === id) {
          setShowSandboxDetails(false)
          setSelectedSandbox(null)
        }

        toast.success(`Deleting sandbox with ID:  ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')
      }
    })
  }

  const handleArchive = async (id: string) => {
    await runSandboxTransition(id, async () => {
      try {
        await archiveSandbox(id)
        toast.success(`Archiving sandbox with ID: ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to archive sandbox')
      }
    })
  }

  const executeBulkAction = useCallback(
    async ({
      ids,
      actionName,
      apiCall,
      toastMessages,
    }: {
      ids: string[]
      actionName: string
      apiCall: (id: string) => Promise<unknown>
      toastMessages: {
        successTitle: string
        errorTitle: string
        warningTitle: string
        canceledTitle: string
      }
    }) => {
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

          startSandboxTransition(id)

          try {
            await apiCall(id)
            successCount += 1
          } catch (error) {
            failureCount += 1
            console.error(`${actionName} sandbox failed`, id, error)
          } finally {
            stopSandboxTransition(id)
          }
        }

        bulkToast.result({ successCount, failureCount }, toastMessages)
      } catch (error) {
        console.error(`${actionName} sandboxes failed`, error)
        bulkToast.error(`${actionName} sandboxes failed.`)
      }

      return { successCount, failureCount }
    },
    [startSandboxTransition, stopSandboxTransition],
  )

  const handleBulkStart = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Starting',
      apiCall: (id) => startSandbox(id),
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
      apiCall: (id) => stopSandbox(id),
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
      apiCall: (id) => archiveSandbox(id),
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
      apiCall: (id) => deleteSandbox(id),
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
      setSandboxAsyncLoading(sandboxId, true)
      try {
        return (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
      } finally {
        setSandboxAsyncLoading(sandboxId, false)
      }
    },
    [sandboxApi, selectedOrganization, setSandboxAsyncLoading],
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
    setSandboxAsyncLoading(id, true)
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
      setSandboxAsyncLoading(id, false)
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

    setSandboxAsyncLoading(id, true)
    try {
      const portPreviewUrl = await getPortPreviewUrl(id, 33333)
      window.open(portPreviewUrl, '_blank')
      toast.success('Opening Screen Recordings dashboard...')
    } catch (error) {
      handleApiError(error, 'Failed to open Screen Recordings')
    } finally {
      setSandboxAsyncLoading(id, false)
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
          {canCreateSandbox && <CreateSandboxSheet ref={createSandboxSheetRef} />}
        </div>
      </PageHeader>
      <PageContent size="full" className="flex-1 max-h-[calc(100vh-65px)]">
        {sandboxesDataError ? (
          <Empty className="py-12">
            <EmptyHeader>
              <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                <AlertCircle />
              </EmptyMedia>
              <EmptyTitle className="text-destructive">Failed to load sandboxes</EmptyTitle>
              <EmptyDescription>Something went wrong while fetching sandboxes. Please try again.</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button variant="secondary" size="sm" onClick={() => handleRefresh()} disabled={sandboxDataIsRefreshing}>
                <RefreshCw />
                Retry
              </Button>
            </EmptyContent>
          </Empty>
        ) : (
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
            isRefetching={sandboxesDataIsRefetching}
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
        )}

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
