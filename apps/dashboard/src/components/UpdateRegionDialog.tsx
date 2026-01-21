/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useEffect, useMemo } from 'react'
import { Region, UpdateRegion, CreateRegionResponse } from '@daytonaio/api-client'
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
import { toast } from 'sonner'
import { Copy, AlertTriangle } from 'lucide-react'
import { getMaskedToken } from '@/lib/utils'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface UpdateRegionDialogProps {
  region: Region
  open: boolean
  onOpenChange: (open: boolean) => void
  onUpdateRegion: (regionId: string, data: UpdateRegion) => Promise<CreateRegionResponse | null>
  loading: boolean
}

export const UpdateRegionDialog: React.FC<UpdateRegionDialogProps> = ({
  region,
  open,
  onOpenChange,
  onUpdateRegion,
  loading,
}) => {
  const [updatedRegion, setUpdatedRegion] = useState<CreateRegionResponse | null>(null)
  const [isProxyApiKeyRevealed, setIsProxyApiKeyRevealed] = useState(false)
  const [isSshGatewayApiKeyRevealed, setIsSshGatewayApiKeyRevealed] = useState(false)
  const [isSnapshotManagerPasswordRevealed, setIsSnapshotManagerPasswordRevealed] = useState(false)

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
      setUpdatedRegion(null)
      setIsProxyApiKeyRevealed(false)
      setIsSshGatewayApiKeyRevealed(false)
      setIsSnapshotManagerPasswordRevealed(false)
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

    const response = await onUpdateRegion(region.id, updateData)
    if (response) {
      if (
        !response.proxyApiKey &&
        !response.sshGatewayApiKey &&
        !response.snapshotManagerUsername &&
        !response.snapshotManagerPassword
      ) {
        onOpenChange(false)
        setUpdatedRegion(null)
      } else {
        setUpdatedRegion(response)
      }
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  const hasCredentials =
    updatedRegion &&
    (updatedRegion.proxyApiKey ||
      updatedRegion.sshGatewayApiKey ||
      updatedRegion.snapshotManagerUsername ||
      updatedRegion.snapshotManagerPassword)

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setUpdatedRegion(null)
          setIsProxyApiKeyRevealed(false)
          setIsSshGatewayApiKeyRevealed(false)
          setIsSnapshotManagerPasswordRevealed(false)
        }
      }}
    >
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{hasCredentials ? 'Region Updated' : `Update Region: ${region.name}`}</DialogTitle>
          <DialogDescription>
            {!hasCredentials
              ? 'Modify the URLs for this region.'
              : "Save these credentials securely. You won't be able to see them again."}
          </DialogDescription>
        </DialogHeader>

        {hasCredentials ? (
          <div className="space-y-6">
            {updatedRegion.proxyApiKey && (
              <div className="space-y-3">
                <Label htmlFor="proxy-api-key">Proxy API Key</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsProxyApiKeyRevealed(true)}
                    onMouseLeave={() => setIsProxyApiKeyRevealed(false)}
                  >
                    {isProxyApiKeyRevealed ? updatedRegion.proxyApiKey : getMaskedToken(updatedRegion.proxyApiKey)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(updatedRegion.proxyApiKey!)}
                  />
                </div>
              </div>
            )}

            {updatedRegion.sshGatewayApiKey && (
              <div className="space-y-3">
                <Label htmlFor="ssh-gateway-api-key">SSH Gateway API Key</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsSshGatewayApiKeyRevealed(true)}
                    onMouseLeave={() => setIsSshGatewayApiKeyRevealed(false)}
                  >
                    {isSshGatewayApiKeyRevealed
                      ? updatedRegion.sshGatewayApiKey
                      : getMaskedToken(updatedRegion.sshGatewayApiKey)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(updatedRegion.sshGatewayApiKey!)}
                  />
                </div>
              </div>
            )}

            {updatedRegion.snapshotManagerUsername && (
              <div className="space-y-3">
                <Label htmlFor="snapshot-manager-username">Snapshot manager username</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span className="overflow-x-auto pr-2 cursor-text select-all">
                    {updatedRegion.snapshotManagerUsername}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(updatedRegion.snapshotManagerUsername!)}
                  />
                </div>
              </div>
            )}

            {updatedRegion.snapshotManagerPassword && (
              <div className="space-y-3">
                <Label htmlFor="snapshot-manager-password">Snapshot manager password</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsSnapshotManagerPasswordRevealed(true)}
                    onMouseLeave={() => setIsSnapshotManagerPasswordRevealed(false)}
                  >
                    {isSnapshotManagerPasswordRevealed
                      ? updatedRegion.snapshotManagerPassword
                      : getMaskedToken(updatedRegion.snapshotManagerPassword)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(updatedRegion.snapshotManagerPassword!)}
                  />
                </div>
              </div>
            )}
          </div>
        ) : (
          <form
            id="update-region-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleUpdate()
            }}
          >
            {hasChanges && (
              <Alert variant="warning">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  Changing URLs will regenerate credentials. Components will need to be redeployed with new credentials.
                </AlertDescription>
              </Alert>
            )}

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
        )}

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              {hasCredentials ? 'Close' : 'Cancel'}
            </Button>
          </DialogClose>
          {!hasCredentials &&
            (loading ? (
              <Button type="button" variant="default" disabled>
                Updating...
              </Button>
            ) : (
              <Button type="submit" form="update-region-form" variant="default" disabled={loading || !hasChanges}>
                Update
              </Button>
            ))}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
