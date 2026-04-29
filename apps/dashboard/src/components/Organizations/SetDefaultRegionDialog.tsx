/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

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
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useSetOrganizationDefaultRegionMutation } from '@/hooks/mutations/useSetOrganizationDefaultRegionMutation'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { Ref, useId, useImperativeHandle, useState } from 'react'
import { toast } from 'sonner'
import { Spinner } from '../ui/spinner'

interface SetDefaultRegionDialogProps {
  ref?: Ref<SetDefaultRegionDialogRef>
}

export type SetDefaultRegionDialogRef = {
  open: () => void
}

export const SetDefaultRegionDialog: React.FC<SetDefaultRegionDialogProps> = ({ ref }) => {
  const { refreshOrganizations } = useOrganizations()
  const { sharedRegions: regions, loadingSharedRegions: loadingRegions } = useRegions()
  const { selectedOrganization } = useSelectedOrganization()
  const setDefaultRegionMutation = useSetOrganizationDefaultRegionMutation()
  const formId = useId()
  const regionSelectId = useId()
  const [open, setOpen] = useState(false)
  const [defaultRegionId, setDefaultRegionId] = useState<string | undefined>(undefined)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const handleSetDefaultRegion = async () => {
    if (!selectedOrganization || !defaultRegionId) {
      return
    }

    try {
      await setDefaultRegionMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        defaultRegionId,
      })
      toast.success('Default region set successfully')
      setOpen(false)
      void refreshOrganizations(selectedOrganization.id)
    } catch (error) {
      handleApiError(error, 'Failed to set default region')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setDefaultRegionId(undefined)
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Set Default Region</DialogTitle>
          <DialogDescription>
            Your organization needs a default region to create sandboxes and manage resources.
          </DialogDescription>
        </DialogHeader>
        {!loadingRegions && regions.length === 0 ? (
          <div className="p-3 rounded-md bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
            <p className="font-medium">No regions available</p>
            <p className="text-sm mt-1">Default region cannot be set because no regions are available.</p>
          </div>
        ) : (
          <form
            id={formId}
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleSetDefaultRegion()
            }}
          >
            <div className="space-y-3">
              <Label htmlFor={regionSelectId}>Region</Label>
              <Select value={defaultRegionId} onValueChange={setDefaultRegionId}>
                <SelectTrigger className="h-8" id={regionSelectId} disabled={loadingRegions} loading={loadingRegions}>
                  <SelectValue placeholder={loadingRegions ? 'Loading regions...' : 'Select a region'} />
                </SelectTrigger>
                <SelectContent>
                  {regions.map((region) => (
                    <SelectItem key={region.id} value={region.id}>
                      {region.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </form>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={setDefaultRegionMutation.isPending}>
              Cancel
            </Button>
          </DialogClose>
          <Button
            type="submit"
            form={formId}
            variant="default"
            disabled={!defaultRegionId || setDefaultRegionMutation.isPending}
          >
            {setDefaultRegionMutation.isPending && <Spinner />}
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
