/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { Region } from '@daytonaio/api-client'
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

interface SetDefaultRegionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  regions: Region[]
  loadingRegions: boolean
  onSetDefaultRegion: (defaultRegionId: string) => Promise<boolean>
}

export const SetDefaultRegionDialog: React.FC<SetDefaultRegionDialogProps> = ({
  open,
  onOpenChange,
  regions,
  loadingRegions,
  onSetDefaultRegion,
}) => {
  const [defaultRegionId, setDefaultRegionId] = useState<string | undefined>(undefined)
  const [loading, setLoading] = useState(false)

  const handleSetDefaultRegion = async () => {
    if (!defaultRegionId) {
      return
    }

    setLoading(true)
    const success = await onSetDefaultRegion(defaultRegionId)
    // TODO: Return when we fix the selected org states
    // if (success) {
    //   onOpenChange(false)
    // }
    // setLoading(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
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
            id="set-default-region-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleSetDefaultRegion()
            }}
          >
            <div className="space-y-3">
              <Label htmlFor="region-select">Region</Label>
              <Select value={defaultRegionId} onValueChange={setDefaultRegionId}>
                <SelectTrigger className="h-8" id="region-select" disabled={loadingRegions} loading={loadingRegions}>
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
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Saving...
            </Button>
          ) : (
            <Button type="submit" form="set-default-region-form" variant="default" disabled={!defaultRegionId}>
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
