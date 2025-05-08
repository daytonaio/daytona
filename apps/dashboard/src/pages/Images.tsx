/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import { ImageDto, ImageState, OrganizationRolePermissionsEnum, PaginatedImagesDto } from '@daytonaio/api-client'
import { ImageTable } from '@/components/ImageTable'
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
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'

const Images: React.FC = () => {
  const { notificationSocket } = useNotificationSocket()

  const { imageApi } = useApi()
  const [imagesData, setImagesData] = useState<PaginatedImagesDto>({
    items: [],
    total: 0,
    page: 1,
    totalPages: 0,
  })
  const [loadingImages, setLoadingImages] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [imageToDelete, setImageToDelete] = useState<ImageDto | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newImageName, setNewImageName] = useState('')
  const [newEntrypoint, setNewEntrypoint] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: 10,
  })

  const fetchImages = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const response = (
          await imageApi.getAllImages(
            selectedOrganization.id,
            paginationParams.pageSize,
            paginationParams.pageIndex + 1,
          )
        ).data
        setImagesData(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch images')
      } finally {
        setLoadingTable(false)
      }
    },
    [imageApi, selectedOrganization, paginationParams.pageIndex, paginationParams.pageSize],
  )

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  useEffect(() => {
    fetchImages()
  }, [fetchImages])

  useEffect(() => {
    const handleImageCreatedEvent = (image: ImageDto) => {
      if (paginationParams.pageIndex === 0) {
        setImagesData((prev) => {
          if (prev.items.some((i) => i.id === image.id)) {
            return prev
          }
          const newImages = [image, ...prev.items]
          const newTotal = prev.total + 1
          return {
            ...prev,
            items: newImages.slice(0, paginationParams.pageSize),
            total: newTotal,
            totalPages: Math.ceil(newTotal / paginationParams.pageSize),
          }
        })
      }
    }

    const handleImageStateUpdatedEvent = (data: { image: ImageDto; oldState: ImageState; newState: ImageState }) => {
      setImagesData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === data.image.id ? data.image : i)),
      }))
    }

    const handleImageEnabledToggledEvent = (image: ImageDto) => {
      setImagesData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === image.id ? image : i)),
      }))
    }

    const handleImageRemovedEvent = (imageId: string) => {
      setImagesData((prev) => {
        const newTotal = Math.max(0, prev.total - 1)
        const newItems = prev.items.filter((i) => i.id !== imageId)

        return {
          ...prev,
          items: newItems,
          total: newTotal,
          totalPages: Math.ceil(newTotal / paginationParams.pageSize),
        }
      })
    }

    notificationSocket.on('image.created', handleImageCreatedEvent)
    notificationSocket.on('image.state.updated', handleImageStateUpdatedEvent)
    notificationSocket.on('image.enabled.toggled', handleImageEnabledToggledEvent)
    notificationSocket.on('image.removed', handleImageRemovedEvent)

    return () => {
      notificationSocket.off('image.created', handleImageCreatedEvent)
      notificationSocket.off('image.state.updated', handleImageStateUpdatedEvent)
      notificationSocket.off('image.enabled.toggled', handleImageEnabledToggledEvent)
      notificationSocket.off('image.removed', handleImageRemovedEvent)
    }
  }, [notificationSocket, paginationParams.pageIndex, paginationParams.pageSize])

  useEffect(() => {
    if (imagesData.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [imagesData.items.length, paginationParams.pageIndex])

  const validateImageName = (name: string): string | null => {
    // Basic format check
    const imageNameRegex =
      /^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:\/[a-z0-9]+(?:[._-][a-z0-9]+)*)*:[a-z0-9]+(?:[._-][a-z0-9]+)*$/

    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Image name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Images with tag ":latest" are not allowed'
    }

    if (!imageNameRegex.test(name)) {
      return 'Invalid image name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  const handleCreate = async () => {
    const validationError = validateImageName(newImageName)
    if (validationError) {
      toast.warning(validationError)
      return
    }

    setLoadingCreate(true)
    try {
      await imageApi.createImage(
        {
          name: newImageName,
          entrypoint: newEntrypoint.trim() ? newEntrypoint.trim().split(' ') : undefined,
        },
        selectedOrganization?.id,
      )
      setShowCreateDialog(false)
      setNewImageName('')
      setNewEntrypoint('')
      toast.success(`Creating image ${newImageName}`)

      if (paginationParams.pageIndex !== 0) {
        setPaginationParams((prev) => ({
          ...prev,
          pageIndex: 0,
        }))
      }
    } catch (error) {
      handleApiError(error, 'Failed to create image')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDelete = async (image: ImageDto) => {
    setLoadingImages((prev) => ({ ...prev, [image.id]: true }))
    try {
      await imageApi.removeImage(image.id, selectedOrganization?.id)
      setImageToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting image ${image.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete image')
    } finally {
      setLoadingImages((prev) => ({ ...prev, [image.id]: false }))
    }
  }

  const handleToggleEnabled = async (image: ImageDto, enabled: boolean) => {
    setLoadingImages((prev) => ({ ...prev, [image.id]: true }))
    try {
      await imageApi.toggleImageState(image.id, { enabled }, selectedOrganization?.id)
      toast.success(`${enabled ? 'Enabling' : 'Disabling'} image ${image.name}`)
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
      <Dialog
        open={showCreateDialog}
        onOpenChange={(isOpen) => {
          setShowCreateDialog(isOpen)
          if (isOpen) {
            return
          }
          setNewImageName('')
          setNewEntrypoint('')
        }}
      >
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Images</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button
                variant="default"
                size="icon"
                disabled={loadingTable}
                className="w-auto px-4"
                title="Create Image"
              >
                <Plus className="w-4 h-4" />
                Create Image
              </Button>
            </DialogTrigger>
          )}
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Image</DialogTitle>
              <DialogDescription>
                Register a new image to be used for spinning up sandboxes in your organization.
              </DialogDescription>
            </DialogHeader>
            <form
              id="create-image-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleCreate()
              }}
            >
              <div className="space-y-3">
                <Label htmlFor="name">Image Name</Label>
                <Input
                  id="name"
                  value={newImageName}
                  onChange={(e) => setNewImageName(e.target.value)}
                  placeholder="ubuntu:22.04"
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Must include a tag (e.g., ubuntu:22.04). The tag "latest" is not allowed.
                </p>
              </div>
              <div className="space-y-3">
                <Label htmlFor="entrypoint">Entrypoint (optional)</Label>
                <Input
                  id="entrypoint"
                  value={newEntrypoint}
                  onChange={(e) => setNewEntrypoint(e.target.value)}
                  placeholder="sleep infinity"
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Ensure that the entrypoint is a long running command. If not provided, or if the image does not have
                  an entrypoint, 'sleep infinity' will be used as the default.
                </p>
              </div>
            </form>
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </DialogClose>
              {loadingCreate ? (
                <Button type="button" variant="default" disabled>
                  Creating...
                </Button>
              ) : (
                <Button
                  type="submit"
                  form="create-image-form"
                  variant="default"
                  disabled={!newImageName.trim() || validateImageName(newImageName) !== null}
                >
                  Create
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </div>

        <ImageTable
          data={imagesData.items}
          loading={loadingTable}
          loadingImages={loadingImages}
          onDelete={(image) => {
            setImageToDelete(image)
            setShowDeleteDialog(true)
          }}
          onToggleEnabled={handleToggleEnabled}
          pageCount={imagesData.totalPages}
          onPaginationChange={handlePaginationChange}
          pagination={{
            pageIndex: paginationParams.pageIndex,
            pageSize: paginationParams.pageSize,
          }}
        />
      </Dialog>

      {imageToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
            if (!isOpen) {
              setImageToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Image Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this image? This action cannot be undone.
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
