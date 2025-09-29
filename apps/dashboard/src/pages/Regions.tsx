/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Plus } from 'lucide-react'
import { RegionDto, CreateRegion, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
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

const REGION_NAME_REGEX = /^[a-zA-Z0-9_.-]+$/

const Regions: React.FC = () => {
  const { regionsApi } = useApi()
  const [regions, setRegions] = useState<RegionDto[]>([])
  const [loadingRegions, setLoadingRegions] = useState<Record<string, boolean>>({})
  const [loadingTable, setLoadingTable] = useState(true)
  const [regionToDelete, setRegionToDelete] = useState<RegionDto | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newRegionName, setNewRegionName] = useState('')
  const [newDockerRegistryId, setNewDockerRegistryId] = useState('')
  const [loadingCreate, setLoadingCreate] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchRegions = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingTable(true)
      }
      try {
        const response = (await regionsApi.listRegions(selectedOrganization.id)).data
        setRegions(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch regions')
      } finally {
        setLoadingTable(false)
      }
    },
    [regionsApi, selectedOrganization],
  )

  useEffect(() => {
    fetchRegions()
  }, [fetchRegions])

  const validateRegionName = (name: string): string | null => {
    if (name.includes(' ')) {
      return 'Spaces are not allowed in region names'
    }

    if (!REGION_NAME_REGEX.test(name)) {
      return 'Invalid region name format. May contain letters, digits, dots, underscores and dashes'
    }

    return null
  }

  const handleCreate = async () => {
    const nameValidationError = validateRegionName(newRegionName)
    if (nameValidationError) {
      toast.warning(nameValidationError)
      return
    }

    setLoadingCreate(true)
    try {
      const createRegionData: CreateRegion = {
        name: newRegionName,
        ...(newDockerRegistryId.trim() ? { dockerRegistryId: newDockerRegistryId.trim() } : {}),
      }

      await regionsApi.createRegion(createRegionData, selectedOrganization?.id)
      setShowCreateDialog(false)
      setNewRegionName('')
      setNewDockerRegistryId('')
      toast.success(`Creating region ${newRegionName}`)
      fetchRegions(false)
    } catch (error) {
      handleApiError(error, 'Failed to create region')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDelete = async (region: RegionDto) => {
    setLoadingRegions((prev) => ({ ...prev, [region.name]: true }))

    try {
      await regionsApi.deleteRegion(region.name, selectedOrganization?.id)
      setRegionToDelete(null)
      setShowDeleteDialog(false)
      toast.success(`Deleting region ${region.name}`)
      fetchRegions(false)
    } catch (error) {
      handleApiError(error, 'Failed to delete region')
    } finally {
      setLoadingRegions((prev) => ({ ...prev, [region.name]: false }))
    }
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGIONS),
    [authenticatedUserHasPermission],
  )

  return (
    <div className="px-6 py-2">
      <Dialog
        open={showCreateDialog}
        onOpenChange={(isOpen) => {
          setShowCreateDialog(isOpen)
          if (isOpen) {
            return
          }
          setNewRegionName('')
          setNewDockerRegistryId('')
        }}
      >
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Regions</h1>
          {writePermitted && (
            <DialogTrigger asChild>
              <Button variant="default" size="sm" disabled={loadingTable} className="w-auto px-4" title="Create Region">
                <Plus className="w-4 h-4" />
                Create Region
              </Button>
            </DialogTrigger>
          )}
        </div>

        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Region</DialogTitle>
            <DialogDescription>
              Create a new region for your organization to manage sandbox deployments.
            </DialogDescription>
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
                onChange={(e) => setNewRegionName(e.target.value)}
                placeholder="us-east-1"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                The name you will use to reference this region in your client app (SDK, CLI).
              </p>
            </div>
            <div className="space-y-3">
              <Label htmlFor="dockerRegistryId">Docker Registry ID (optional)</Label>
              <Input
                id="dockerRegistryId"
                value={newDockerRegistryId}
                onChange={(e) => setNewDockerRegistryId(e.target.value)}
                placeholder="registry-uuid"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                Optional Docker registry ID to associate with this region.
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
                form="create-region-form"
                variant="default"
                disabled={!newRegionName.trim() || validateRegionName(newRegionName) !== null}
              >
                Create
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <RegionTable
        data={regions}
        loading={loadingTable}
        loadingRegions={loadingRegions}
        onDelete={(region) => {
          setRegionToDelete(region)
          setShowDeleteDialog(true)
        }}
        writePermitted={writePermitted}
      />

      {regionToDelete && (
        <Dialog
          open={showDeleteDialog}
          onOpenChange={(isOpen) => {
            setShowDeleteDialog(isOpen)
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
                disabled={loadingRegions[regionToDelete.name]}
              >
                {loadingRegions[regionToDelete.name] ? 'Deleting...' : 'Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default Regions
