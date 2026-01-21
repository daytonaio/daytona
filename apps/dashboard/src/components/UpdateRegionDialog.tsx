/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useEffect, useMemo } from 'react'
import { Region, UpdateRegion } from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface UpdateRegionDialogProps {
  region: Region
  open: boolean
  onOpenChange: (open: boolean) => void
  onUpdateRegion: (regionId: string, data: UpdateRegion) => Promise<boolean>
  loading: boolean
}

export const UpdateRegionDialog: React.FC<UpdateRegionDialogProps> = ({
  region,
  open,
  onOpenChange,
  onUpdateRegion,
  loading,
}) => {
  const [formData, setFormData] = useState({
    proxyUrl: region.proxyUrl || '',
    sshGatewayUrl: region.sshGatewayUrl || '',
    snapshotManagerUrl: region.snapshotManagerUrl || '',
  })

  // Reset form when dialog opens with new region
  useEffect(() => {
    if (open) {
      setFormData({
        proxyUrl: region.proxyUrl || '',
        sshGatewayUrl: region.sshGatewayUrl || '',
        snapshotManagerUrl: region.snapshotManagerUrl || '',
      })
    }
  }, [open, region])

  const hasChanges = useMemo(() => {
    const proxyChanged = (formData.proxyUrl.trim() || null) !== (region.proxyUrl || null)
    const sshGatewayChanged = (formData.sshGatewayUrl.trim() || null) !== (region.sshGatewayUrl || null)
    const snapshotManagerChanged = (formData.snapshotManagerUrl.trim() || null) !== (region.snapshotManagerUrl || null)
    return proxyChanged || sshGatewayChanged || snapshotManagerChanged
  }, [formData, region])

  const handleUpdate = async () => {
    // Only include changed fields
    const updateData: UpdateRegion = {}

    const proxyUrlValue = formData.proxyUrl.trim() || null
    const sshGatewayUrlValue = formData.sshGatewayUrl.trim() || null
    const snapshotManagerUrlValue = formData.snapshotManagerUrl.trim() || null

    if (proxyUrlValue !== (region.proxyUrl || null)) {
      updateData.proxyUrl = proxyUrlValue
    }
    if (sshGatewayUrlValue !== (region.sshGatewayUrl || null)) {
      updateData.sshGatewayUrl = sshGatewayUrlValue
    }
    if (snapshotManagerUrlValue !== (region.snapshotManagerUrl || null)) {
      updateData.snapshotManagerUrl = snapshotManagerUrlValue
    }

    const success = await onUpdateRegion(region.id, updateData)
    if (success) {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Update Region: {region.name}</DialogTitle>
          <DialogDescription>Modify the URLs for this region.</DialogDescription>
        </DialogHeader>

        <form
          id="update-region-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleUpdate()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="proxy-url">Proxy URL</Label>
            <Input
              id="proxy-url"
              value={formData.proxyUrl}
              onChange={(e) => {
                setFormData((prev) => ({ ...prev, proxyUrl: e.target.value }))
              }}
              placeholder="https://proxy.example.com"
            />
            <p className="text-sm text-muted-foreground mt-1 pl-1">
              (Optional) URL of the custom proxy for this region
            </p>
          </div>

          <div className="space-y-3">
            <Label htmlFor="ssh-gateway-url">SSH gateway URL</Label>
            <Input
              id="ssh-gateway-url"
              value={formData.sshGatewayUrl}
              onChange={(e) => {
                setFormData((prev) => ({ ...prev, sshGatewayUrl: e.target.value }))
              }}
              placeholder="https://ssh-gateway.example.com"
            />
            <p className="text-sm text-muted-foreground mt-1 pl-1">
              (Optional) URL of the custom SSH gateway for this region
            </p>
          </div>

          <div className="space-y-3">
            <Label htmlFor="snapshot-manager-url">Snapshot manager URL</Label>
            <Input
              id="snapshot-manager-url"
              value={formData.snapshotManagerUrl}
              onChange={(e) => {
                setFormData((prev) => ({ ...prev, snapshotManagerUrl: e.target.value }))
              }}
              placeholder="https://snapshot-manager.example.com"
            />
            <p className="text-sm text-muted-foreground mt-1 pl-1">
              (Optional) URL of the custom snapshot manager for this region. Cannot be changed if snapshots exist in
              this region.
            </p>
          </div>
        </form>

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Updating...
            </Button>
          ) : (
            <Button type="submit" form="update-region-form" variant="default" disabled={loading || !hasChanges}>
              Update
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
