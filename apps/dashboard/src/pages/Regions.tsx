/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import { Region, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { RegionTable } from '@/components/RegionTable'
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
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'
import { useRegions } from '@/hooks/useRegions'

const Regions: React.FC = () => {
  const { regionsApi } = useApi()
  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const { organizationRegions, loadingRegions, refreshRegions } = useRegions()

  const [regionIsLoading, setRegionIsLoading] = useState<Record<string, boolean>>({})

  const [createRegionDialogIsOpen, setCreateRegionDialogIsOpen] = useState(false)
  const [newRegionName, setNewRegionName] = useState('')
  const [loadingCreateRegion, setLoadingCreateDialog] = useState(false)

  const [regionToDelete, setRegionToDelete] = useState<Region | null>(null)
  const [deleteRegionDialogIsOpen, setDeleteRegionDialogIsOpen] = useState(false)

  const handleCreate = async () => {
    setLoadingCreateDialog(true)
    try {
      await regionsApi.createRegion(
        {
          name: newRegionName,
        },
        selectedOrganization?.id,
      )
      setCreateRegionDialogIsOpen(false)
      setNewRegionName('')
      toast.success(`Creating region ${newRegionName}`)
      await refreshRegions()
    } catch (error) {
      handleApiError(error, 'Failed to create region')
    } finally {
      setLoadingCreateDialog(false)
    }
  }

  const handleDelete = async (region: Region) => {
    setRegionIsLoading((prev) => ({ ...prev, [region.id]: true }))

    try {
      await regionsApi.deleteRegion(region.id, selectedOrganization?.id)
      setRegionToDelete(null)
      setDeleteRegionDialogIsOpen(false)
      toast.success(`Deleting region ${region.name}`)
      await refreshRegions()
    } catch (error) {
      handleApiError(error, 'Failed to delete region')
    } finally {
      setRegionIsLoading((prev) => ({ ...prev, [region.id]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGIONS),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_REGIONS),
    [authenticatedUserHasPermission],
  )

  return (
    <div className="px-6 py-2">
      <Dialog
        open={createRegionDialogIsOpen}
        onOpenChange={(isOpen) => {
          setCreateRegionDialogIsOpen(isOpen)
          if (isOpen) {
            return
          }
          setNewRegionName('')
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Regions</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button
                variant="default"
                size="sm"
                disabled={loadingRegions}
                className="w-auto px-4"
                title="Create Region"
              >
                <Plus className="w-4 h-4" />
                Create Region
              </Button>
            </DialogTrigger>
          )}
        </div>

        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Region</DialogTitle>
            <DialogDescription>Add a new region for grouping runners and sandboxes.</DialogDescription>
          </DialogHeader>
          <form
            id="create-region-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleCreate()
            }}
          >
            <div className="space-y-3">
              <Label htmlFor="name">Region Name</Label>
              <Input
                id="name"
                value={newRegionName}
                onChange={(e) => {
                  setNewRegionName(e.target.value)
                }}
                placeholder="us-east-1"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                Region name must contain only letters, numbers, underscores, periods, and hyphens.
              </p>
            </div>
          </form>
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </DialogClose>
            {loadingCreateRegion ? (
              <Button type="button" variant="default" disabled>
                Creating...
              </Button>
            ) : (
              <Button
                type="submit"
                form="create-region-form"
                variant="default"
                disabled={!newRegionName.trim() || loadingCreateRegion}
              >
                Create
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <RegionTable
        data={organizationRegions}
        loading={loadingRegions}
        isLoadingRegion={(region) => regionIsLoading[region.id] || false}
        deletePermitted={deletePermitted}
        onDelete={(region) => {
          setRegionToDelete(region)
          setDeleteRegionDialogIsOpen(true)
        }}
      />

      {regionToDelete && (
        <Dialog
          open={deleteRegionDialogIsOpen}
          onOpenChange={(isOpen) => {
            setDeleteRegionDialogIsOpen(isOpen)
            if (!isOpen) {
              setRegionToDelete(null)
            }
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Region Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this region? This action cannot be undone.
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
                onClick={() => handleDelete(regionToDelete)}
                disabled={regionIsLoading[regionToDelete.id]}
              >
                {regionIsLoading[regionToDelete.id] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Regions
