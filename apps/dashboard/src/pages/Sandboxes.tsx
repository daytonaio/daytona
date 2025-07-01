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
  SandboxState,
  OrganizationRolePermissionsEnum,
  SnapshotDto,
} from '@daytonaio/api-client'
import { SandboxTable } from '@/components/SandboxTable/index'
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

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, snapshotApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()

  const [sandboxes, setSandboxes] = useState<Sandbox[]>([])
  const [snapshots, setSnapshots] = useState<SnapshotDto[]>([])
  const [loadingSandboxes, setLoadingSandboxes] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [loadingSnapshots, setLoadingSnapshots] = useState(true)
  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedSandbox, setSelectedSandbox] = useState<Sandbox | null>(null)
  const [showSandboxDetails, setShowSandboxDetails] = useState(false)

  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const fetchSnapshots = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setLoadingSnapshots(true)
    try {
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

    notificationSocket.on('sandbox.created', handleSandboxCreatedEvent)
    notificationSocket.on('sandbox.state.updated', handleSandboxStateUpdatedEvent)

    return () => {
      notificationSocket.off('sandbox.created', handleSandboxCreatedEvent)
      notificationSocket.off('sandbox.state.updated', handleSandboxStateUpdatedEvent)
    }
  }, [notificationSocket, sandboxes])

  const handleStart = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

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
          <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING)}>
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
    }
  }

  const handleStop = async (id: string) => {
    setLoadingSandboxes((prev) => ({ ...prev, [id]: true }))

    const sandboxToStop = sandboxes.find((s) => s.id === id)
    const previousState = sandboxToStop?.state

    setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: SandboxState.STOPPING } : s)))

    if (selectedSandbox?.id === id) {
      setSelectedSandbox((prev) => (prev ? { ...prev, state: SandboxState.STOPPING } : null))
    }

    try {
      await sandboxApi.stopSandbox(id, selectedOrganization?.id)
      toast.success(`Stopping sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      setSandboxes((prev) => prev.map((s) => (s.id === id ? { ...s, state: previousState } : s)))
      if (selectedSandbox?.id === id && previousState) {
        setSelectedSandbox((prev) => (prev ? { ...prev, state: previousState } : null))
      }
    } finally {
      setLoadingSandboxes((prev) => ({ ...prev, [id]: false }))
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
    <div className="flex flex-col h-dvh px-6 py-2 ">
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
        handleStart={handleStart}
        handleStop={handleStop}
        handleDelete={(id: string) => {
          setSandboxToDelete(id)
          setShowDeleteDialog(true)
        }}
        handleBulkDelete={handleBulkDelete}
        handleArchive={handleArchive}
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
        writePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
        deletePermitted={authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER}
      />
    </div>
  )
}

export default Sandboxes
