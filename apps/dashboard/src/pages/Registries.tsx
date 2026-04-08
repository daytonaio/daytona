/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { UpsertRegistrySheet } from '@/components/UpsertRegistrySheet'
import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { RegistryTable } from '@/components/RegistryTable'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useDeleteRegistryMutation } from '@/hooks/mutations/useDeleteRegistryMutation'
import { useRegistriesQuery } from '@/hooks/queries/useRegistriesQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationRolePermissionsEnum, type DockerRegistry } from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import React, { useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

const Registries: React.FC = () => {
  const [registryToDelete, setRegistryToDelete] = useState<string | null>(null)
  const [registryToEdit, setRegistryToEdit] = useState<DockerRegistry | null>(null)
  const [showEditSheet, setShowEditSheet] = useState(false)
  const addRegistrySheetRef = useRef<{ open: () => void }>(null)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const { data: registries = [], isLoading: loading, error: registriesError } = useRegistriesQuery()
  const deleteRegistryMutation = useDeleteRegistryMutation()
  const deleteInProgress = deleteRegistryMutation.isPending

  useEffect(() => {
    if (registriesError) {
      handleApiError(registriesError, 'Failed to fetch registries')
    }
  }, [registriesError])

  const handleDelete = async (id: string) => {
    try {
      await deleteRegistryMutation.mutateAsync({
        registryId: id,
        organizationId: selectedOrganization?.id,
      })
      toast.success('Registry deleted successfully')
      setRegistryToDelete(null)
    } catch (error) {
      handleApiError(error, 'Failed to delete registry')
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!writePermitted) {
      return []
    }

    return [
      {
        id: 'add-registry',
        label: 'Add Registry',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => addRegistrySheetRef.current?.open(),
      },
    ]
  }, [writePermitted])

  useRegisterCommands(rootCommands, { groupId: 'registry-actions', groupLabel: 'Registry actions', groupOrder: 0 })

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Registries</PageTitle>
        {writePermitted && <UpsertRegistrySheet className="ml-auto" disabled={loading} ref={addRegistrySheetRef} />}
      </PageHeader>

      <PageContent size="full">
        <RegistryTable
          data={registries}
          loading={loading}
          onDelete={(id) => setRegistryToDelete(id)}
          onEdit={(registry) => {
            setRegistryToEdit(registry)
            setShowEditSheet(true)
          }}
        />

        <UpsertRegistrySheet
          mode="edit"
          trigger={null}
          open={showEditSheet}
          onOpenChange={(isOpen) => {
            setShowEditSheet(isOpen)
            if (!isOpen) {
              setRegistryToEdit(null)
            }
          }}
          registry={registryToEdit}
        />

        <Dialog
          open={!!registryToDelete}
          onOpenChange={(isOpen) => {
            if (isOpen) {
              return
            }

            setRegistryToDelete(null)
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Registry Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this registry? This action cannot be undone.
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
                onClick={() => {
                  if (registryToDelete) {
                    handleDelete(registryToDelete)
                  }
                }}
                disabled={deleteInProgress}
              >
                {deleteInProgress ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </PageContent>
    </PageLayout>
  )
}

export default Registries
