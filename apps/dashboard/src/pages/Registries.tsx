/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import {
  DockerRegistryRegistryTypeEnum,
  OrganizationRolePermissionsEnum,
  type DockerRegistry,
} from '@daytonaio/api-client'
import { RegistryTable } from '@/components/RegistryTable'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'
import { toast } from 'sonner'

const Registries: React.FC = () => {
  const { dockerRegistryApi } = useApi()
  const [registries, setRegistries] = useState<DockerRegistry[]>([])
  const [loading, setLoading] = useState(true)
  const [registryToDelete, setRegistryToDelete] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    username: '',
    password: '',
    project: '',
  })
  const [registryToEdit, setRegistryToEdit] = useState<DockerRegistry | null>(null)
  const [showCreateOrEditDialog, setShowCreateOrEditDialog] = useState(false)
  const [actionInProgress, setActionInProgress] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchRegistries = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoading(true)
      }
      try {
        const response = await dockerRegistryApi.listRegistries(selectedOrganization.id)
        setRegistries(response.data)
      } catch (error) {
        handleApiError(error, 'Failed to fetch registries')
      } finally {
        setLoading(false)
      }
    },
    [dockerRegistryApi, selectedOrganization],
  )

  useEffect(() => {
    fetchRegistries()
  }, [fetchRegistries])

  const handleCreate = async () => {
    setActionInProgress(true)
    try {
      await dockerRegistryApi.createRegistry(
        { ...formData, isDefault: false, registryType: DockerRegistryRegistryTypeEnum.ORGANIZATION },
        selectedOrganization?.id,
      )
      toast.success('Registry created successfully')
      await fetchRegistries(false)
      setShowCreateOrEditDialog(false)
      setFormData({
        name: '',
        url: '',
        username: '',
        password: '',
        project: '',
      })
    } catch (error) {
      handleApiError(error, 'Failed to create registry')
    } finally {
      setActionInProgress(false)
    }
  }

  const handleEdit = async () => {
    if (!registryToEdit) return

    setActionInProgress(true)
    try {
      await dockerRegistryApi.updateRegistry(registryToEdit.id, formData, selectedOrganization?.id)
      toast.success('Registry edited successfully')
      await fetchRegistries(false)
      setShowCreateOrEditDialog(false)
      setRegistryToEdit(null)
      setFormData({
        name: '',
        url: '',
        username: '',
        password: '',
        project: '',
      })
    } catch (error) {
      handleApiError(error, 'Failed to edit registry')
    } finally {
      setActionInProgress(false)
    }
  }

  const handleDelete = async (id: string) => {
    setActionInProgress(true)
    try {
      await dockerRegistryApi.deleteRegistry(id, selectedOrganization?.id)
      toast.success('Registry deleted successfully')
      await fetchRegistries(false)
      setRegistryToDelete(null)
    } catch (error) {
      handleApiError(error, 'Failed to delete registry')
    } finally {
      setActionInProgress(false)
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const dialogContent = (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{registryToEdit ? 'Edit Registry' : 'Add Registry'}</DialogTitle>
        <DialogDescription>
          Registry details must be provided for images that are not publicly available.
        </DialogDescription>
      </DialogHeader>
      <form
        id="registry-form"
        className="space-y-6 overflow-y-auto px-1 pb-1"
        onSubmit={async (e) => {
          e.preventDefault()
          if (registryToEdit) {
            await handleEdit()
          } else {
            await handleCreate()
          }
        }}
      >
        <div className="space-y-3">
          <Label htmlFor="name">Registry Name</Label>
          <Input
            id="name"
            value={formData.name}
            onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
            placeholder="My Registry"
          />
        </div>
        <div className="space-y-3">
          <Label htmlFor="url">Registry URL</Label>
          <Input
            id="url"
            value={formData.url}
            onChange={(e) => setFormData((prev) => ({ ...prev, url: e.target.value }))}
            placeholder="https://registry.example.com"
          />
        </div>
        <div className="space-y-3">
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            value={formData.username}
            onChange={(e) => setFormData((prev) => ({ ...prev, username: e.target.value }))}
          />
        </div>
        <div className="space-y-3">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            value={formData.password}
            onChange={(e) => setFormData((prev) => ({ ...prev, password: e.target.value }))}
          />
          {registryToEdit && (
            <p className="text-sm text-gray-500">Leave blank to use the same password as the registry.</p>
          )}
        </div>
        <div className="space-y-3">
          <Label htmlFor="project">Project</Label>
          <Input
            id="project"
            value={formData.project}
            onChange={(e) => setFormData((prev) => ({ ...prev, project: e.target.value }))}
            placeholder="my-project"
          />
        </div>
      </form>
      <DialogFooter>
        <DialogClose asChild>
          <Button type="button" variant="secondary">
            Cancel
          </Button>
        </DialogClose>
        {actionInProgress ? (
          <Button type="button" variant="default" disabled>
            {registryToEdit ? 'Editing...' : 'Adding...'}
          </Button>
        ) : (
          <Button
            type="submit"
            form="registry-form"
            variant="default"
            disabled={
              !formData.name ||
              !formData.url ||
              !formData.username ||
              !formData.project ||
              (!registryToEdit && !formData.password)
            }
          >
            {registryToEdit ? 'Edit' : 'Add'}
          </Button>
        )}
      </DialogFooter>
    </DialogContent>
  )

  return (
    <div className="px-6 py-2">
      <Dialog
        open={showCreateOrEditDialog}
        onOpenChange={(isOpen) => {
          setShowCreateOrEditDialog(isOpen)
          if (isOpen) {
            return
          }

          setRegistryToDelete(null)
          setRegistryToEdit(null)
          setFormData({
            name: '',
            url: '',
            username: '',
            password: '',
            project: '',
          })
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Container Registries</h1>
          {writePermitted && (
            <DialogTrigger asChild disabled={loading}>
              <Button
                variant="default"
                size="sm"
                disabled={loading}
                className="w-auto px-4"
                title="Add Registry"
                onClick={() =>
                  setFormData({
                    name: '',
                    url: '',
                    username: '',
                    password: '',
                    project: '',
                  })
                }
              >
                <Plus className="w-4 h-4" />
                Add Registry
              </Button>
            </DialogTrigger>
          )}
          {dialogContent}
        </div>

        <RegistryTable
          data={registries}
          loading={loading}
          onDelete={(id) => setRegistryToDelete(id)}
          onEdit={(registry) => {
            setFormData({
              name: registry.name,
              url: registry.url,
              username: registry.username,
              password: '',
              project: registry.project,
            })
            setRegistryToEdit(registry)
          }}
        />
      </Dialog>

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
            <Button variant="destructive" onClick={() => handleDelete(registryToDelete!)} disabled={actionInProgress}>
              {actionInProgress ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default Registries
