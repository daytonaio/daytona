/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import { ImageDto, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { ImageTable } from '@/components/ImageTable'
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
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'

const Images: React.FC = () => {
  const { imageApi } = useApi()
  const { notificationSocket } = useNotificationSocket()
  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const [images, setImages] = useState<ImageDto[]>([])
  const [loadingImages, setLoadingImages] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)

  const [imageToDelete, setImageToDelete] = useState<ImageDto | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [imageName, setImageName] = useState('')
  const [entrypoint, setEntrypoint] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)

  const fetchImages = useCallback(
    async (showTableLoading = true) => {
      if (!selectedOrganization) return
      if (showTableLoading) setLoadingTable(true)
      try {
        const response = (await imageApi.getAllImages(selectedOrganization.id)).data
        setImages(response.items)
      } catch (error) {
        handleApiError(error, 'Failed to fetch images')
      } finally {
        setLoadingTable(false)
      }
    },
    [imageApi, selectedOrganization],
  )

  useEffect(() => {
    fetchImages()
  }, [fetchImages])

  useEffect(() => {
    const onCreate = (image: ImageDto) => setImages((prev) => [image, ...prev])
    const onUpdate = (data: { image: ImageDto }) =>
      setImages((prev) => prev.map((i) => (i.id === data.image.id ? data.image : i)))
    const onRemove = (id: string) => setImages((prev) => prev.filter((i) => i.id !== id))

    notificationSocket.on('image.created', onCreate)
    notificationSocket.on('image.state.updated', onUpdate)
    notificationSocket.on('image.enabled.toggled', onUpdate)
    notificationSocket.on('image.removed', onRemove)

    return () => {
      notificationSocket.off('image.created', onCreate)
      notificationSocket.off('image.state.updated', onUpdate)
      notificationSocket.off('image.enabled.toggled', onUpdate)
      notificationSocket.off('image.removed', onRemove)
    }
  }, [notificationSocket])

  const validateImageName = (name: string): string | null => {
    const regex = /^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:\/[a-z0-9]+(?:[._-][a-z0-9]+)*)*:[a-z0-9]+(?:[._-][a-z0-9]+)*$/
    if (!name.includes(':') || name.endsWith(':')) return 'Image name must include a tag (e.g., ubuntu:22.04)'
    if (name.endsWith(':latest')) return 'Images with tag ":latest" are not allowed'
    if (!regex.test(name)) return 'Invalid image format. Must be lowercase with tag.'
    return null
  }

  const handleCreate = useCallback(async () => {
    const error = validateImageName(imageName)
    if (error) {
      toast.warning(error)
      return
    }

    setLoadingCreate(true)
    try {
      await imageApi.createImage(
        {
          name: imageName,
          entrypoint: entrypoint.trim() ? entrypoint.trim().split(' ') : undefined,
        },
        selectedOrganization?.id,
      )
      toast.success(`Creating image ${imageName}`)
      setShowCreateDialog(false)
      setImageName('')
      setEntrypoint('')
    } catch (error) {
      handleApiError(error, 'Failed to create image')
    } finally {
      setLoadingCreate(false)
    }
  }, [imageApi, imageName, entrypoint, selectedOrganization])

  const handleDelete = useCallback(
    async (image: ImageDto) => {
      setLoadingImages((prev) => ({ ...prev, [image.id]: true }))
      try {
        await imageApi.removeImage(image.id, selectedOrganization?.id)
        toast.success(`Deleted image ${image.name}`)
        setImageToDelete(null)
        setShowDeleteDialog(false)
      } catch (error) {
        handleApiError(error, 'Failed to delete image')
      } finally {
        setLoadingImages((prev) => ({ ...prev, [image.id]: false }))
      }
    },
    [imageApi, selectedOrganization],
  )

  const handleBulkDelete = async (ids: string[]) => {
    setLoadingImages((prev) => ({ ...prev, ...ids.reduce((acc, id) => ({ ...acc, [id]: true }), {}) }))
    for (const id of ids) {
      try {
        await imageApi.removeImage(id, selectedOrganization?.id)
        toast.success(`Deleted image with ID: ${id}`)
      } catch (error) {
        handleApiError(error, `Failed to delete image with ID: ${id}`)
        const shouldContinue = window.confirm(
          `Failed to delete image with ID: ${id}. Do you want to continue with the remaining images?`,
        )
        if (!shouldContinue) break
      } finally {
        setLoadingImages((prev) => ({ ...prev, [id]: false }))
      }
    }
  }

  const handleToggleEnabled = async (image: ImageDto, enabled: boolean) => {
    setLoadingImages((prev) => ({ ...prev, [image.id]: true }))
    try {
      await imageApi.toggleImageState(image.id, { enabled }, selectedOrganization?.id)
      toast.success(`${enabled ? 'Enabled' : 'Disabled'} image ${image.name}`)
    } catch (error) {
      handleApiError(error, enabled ? 'Failed to enable image' : 'Failed to disable image')
    } finally {
      setLoadingImages((prev) => ({ ...prev, [image.id]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_IMAGES),
    [authenticatedUserHasPermission],
  )

  return (
    <div className="p-6">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-2xl font-bold">Images</h1>
        {writePermitted && (
          <Button
            variant="default"
            size="icon"
            disabled={loadingTable}
            className="w-auto px-4"
            title="Create Image"
            onClick={() => setShowCreateDialog(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Create Image
          </Button>
        )}
      </div>

      <ImageTable
        data={images}
        loading={loadingTable}
        loadingImages={loadingImages}
        onDelete={(image) => {
          setImageToDelete(image)
          setShowDeleteDialog(true)
        }}
        onBulkDelete={handleBulkDelete}
        onToggleEnabled={handleToggleEnabled}
      />

      {/* CREATE DIALOG */}
      <Dialog
        open={showCreateDialog}
        onOpenChange={(v) => {
          setShowCreateDialog(v)
          if (!v) {
            setImageName('')
            setEntrypoint('')
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Image</DialogTitle>
            <DialogDescription>Register a new image to be used for sandboxes.</DialogDescription>
          </DialogHeader>
          <form
            id="create-image-form"
            className="space-y-6"
            onSubmit={(e) => {
              e.preventDefault()
              handleCreate()
            }}
          >
            <div>
              <Label htmlFor="image-name">Image Name</Label>
              <Input
                id="image-name"
                value={imageName}
                onChange={(e) => setImageName(e.target.value)}
                placeholder="ubuntu:22.04"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                Must include a tag (e.g., ubuntu:22.04). Tag "latest" is not allowed.
              </p>
            </div>
            <div>
              <Label htmlFor="entrypoint">Entrypoint (optional)</Label>
              <Input
                id="entrypoint"
                value={entrypoint}
                onChange={(e) => setEntrypoint(e.target.value)}
                placeholder="sleep infinity"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">Will default to 'sleep infinity' if left blank.</p>
            </div>
          </form>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="secondary">Cancel</Button>
            </DialogClose>
            <Button
              type="submit"
              form="create-image-form"
              variant="default"
              disabled={!imageName.trim() || !!validateImageName(imageName) || loadingCreate}
            >
              {loadingCreate ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* DELETE DIALOG */}
      {imageToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(v) => {
            setShowDeleteDialog(v)
            if (!v) setImageToDelete(null)
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Image Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete <strong>{imageToDelete.name}</strong>? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="secondary">Cancel</Button>
              </DialogClose>
              <Button
                variant="destructive"
                onClick={() => handleDelete(imageToDelete)}
                disabled={loadingImages[imageToDelete.id]}
              >
                {loadingImages[imageToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Images
