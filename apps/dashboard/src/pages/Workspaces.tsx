/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState, useCallback } from 'react'
import { useApi } from '@/hooks/useApi'
import { OrganizationSuspendedError } from '@/api/errors'
import { OrganizationUserRoleEnum, Workspace, WorkspaceState } from '@daytonaio/api-client'
import { WorkspaceTable } from '@/components/WorkspaceTable'
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

const Workspaces: React.FC = () => {
  const { workspaceApi, apiKeyApi } = useApi()
  const { user } = useAuth()
  const { notificationSocket } = useNotificationSocket()

  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [loadingWorkspaces, setLoadingWorkspaces] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [workspaceToDelete, setWorkspaceToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const fetchWorkspaces = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const workspaces = (await workspaceApi.listWorkspaces(selectedOrganization.id)).data
        setWorkspaces(workspaces)
      } catch (error) {
        handleApiError(error, 'Failed to fetch sandboxes')
      } finally {
        setLoadingTable(false)
      }
    },
    [workspaceApi, selectedOrganization],
  )

  useEffect(() => {
    fetchWorkspaces()
  }, [fetchWorkspaces])

  useEffect(() => {
    const handleWorkspaceCreatedEvent = (workspace: Workspace) => {
      if (!workspaces.some((w) => w.id === workspace.id)) {
        setWorkspaces((prev) => [workspace, ...prev])
      }
    }

    const handleWorkspaceStateUpdatedEvent = (data: {
      workspace: Workspace
      oldState: WorkspaceState
      newState: WorkspaceState
    }) => {
      if (data.newState === WorkspaceState.DESTROYED) {
        setWorkspaces((prev) => prev.filter((w) => w.id !== data.workspace.id))
      } else if (!workspaces.some((w) => w.id === data.workspace.id)) {
        setWorkspaces((prev) => [data.workspace, ...prev])
      } else {
        setWorkspaces((prev) => prev.map((w) => (w.id === data.workspace.id ? data.workspace : w)))
      }
    }

    notificationSocket.on('workspace.created', handleWorkspaceCreatedEvent)
    notificationSocket.on('workspace.state.updated', handleWorkspaceStateUpdatedEvent)

    return () => {
      notificationSocket.off('workspace.created', handleWorkspaceCreatedEvent)
      notificationSocket.off('workspace.state.updated', handleWorkspaceStateUpdatedEvent)
    }
  }, [notificationSocket, workspaces])

  const handleStart = async (id: string) => {
    setLoadingWorkspaces((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const workspaceToStart = workspaces.find((w) => w.id === id)
    const previousState = workspaceToStart?.state

    // Optimistically update the workspace state
    setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: WorkspaceState.STARTING } : w)))

    try {
      await workspaceApi.startWorkspace(id, selectedOrganization?.id)
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
      setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: previousState } : w)))
    } finally {
      setLoadingWorkspaces((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleStop = async (id: string) => {
    setLoadingWorkspaces((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const workspaceToStop = workspaces.find((w) => w.id === id)
    const previousState = workspaceToStop?.state

    // Optimistically update the workspace state
    setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: WorkspaceState.STOPPING } : w)))

    try {
      await workspaceApi.stopWorkspace(id, selectedOrganization?.id)
      toast.success(`Stopping sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      // Revert the optimistic update
      setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: previousState } : w)))
    } finally {
      setLoadingWorkspaces((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleDelete = async (id: string) => {
    setLoadingWorkspaces((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const workspaceToDelete = workspaces.find((w) => w.id === id)
    const previousState = workspaceToDelete?.state

    // Optimistically update the workspace state
    setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: WorkspaceState.DESTROYING } : w)))

    try {
      await workspaceApi.deleteWorkspace(id, true, selectedOrganization?.id)
      setWorkspaceToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting sandbox with ID:  ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      // Revert the optimistic update
      setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: previousState } : w)))
    } finally {
      setLoadingWorkspaces((prev) => ({ ...prev, [id]: false }))
    }
  }

  const handleBulkDelete = async (ids: string[]) => {
    setLoadingWorkspaces((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))

    for (const id of ids) {
      // Save the current state
      const workspaceToDelete = workspaces.find((w) => w.id === id)
      const previousState = workspaceToDelete?.state

      // Optimistically update the workspace state
      setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: WorkspaceState.DESTROYING } : w)))

      try {
        await workspaceApi.deleteWorkspace(id, true, selectedOrganization?.id)
        toast.success(`Deleting sandbox with ID: ${id}`)
      } catch (error) {
        handleApiError(error, 'Failed to delete sandbox')

        // Revert the optimistic update
        setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: previousState } : w)))

        // Wait for user decision
        const shouldContinue = window.confirm(
          `Failed to delete sandbox with ID: ${id}. Do you want to continue with the remaining sandboxes?`,
        )

        if (!shouldContinue) {
          break
        }
      } finally {
        setLoadingWorkspaces((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: false }), {}) }))
      }
    }
  }

  const handleArchive = async (id: string) => {
    setLoadingWorkspaces((prev) => ({ ...prev, [id]: true }))

    // Save the current state
    const workspaceToArchive = workspaces.find((w) => w.id === id)
    const previousState = workspaceToArchive?.state

    // Optimistically update the workspace state
    setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: WorkspaceState.ARCHIVING } : w)))

    try {
      await workspaceApi.archiveWorkspace(id, selectedOrganization?.id)
      toast.success(`Archiving sandbox with ID: ${id}`)
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      // Revert the optimistic update
      setWorkspaces((prev) => prev.map((w) => (w.id === id ? { ...w, state: previousState } : w)))
    } finally {
      setLoadingWorkspaces((prev) => ({ ...prev, [id]: false }))
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
        {!loadingTable && workspaces.length === 0 && (
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
      <WorkspaceTable
        loadingWorkspaces={loadingWorkspaces}
        handleStart={handleStart}
        handleStop={handleStop}
        handleDelete={(id) => {
          setWorkspaceToDelete(id)
          setShowDeleteDialog(true)
        }}
        handleBulkDelete={handleBulkDelete}
        handleArchive={handleArchive}
        data={workspaces}
        loading={loadingTable}
      />

      {workspaceToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setWorkspaceToDelete(null)
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
                onClick={() => handleDelete(workspaceToDelete)}
                disabled={loadingWorkspaces[workspaceToDelete]}
              >
                {loadingWorkspaces[workspaceToDelete] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Workspaces
