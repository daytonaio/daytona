/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { CreateRegion, CreateRegionResponse } from '@daytonaio/api-client'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { toast } from 'sonner'
import { Plus, Copy } from 'lucide-react'
import { getMaskedToken } from '@/lib/utils'

const DEFAULT_FORM_DATA = {
  name: '',
  proxyUrl: '',
  sshGatewayUrl: '',
  snapshotManagerUrl: '',
}

interface CreateRegionDialogProps {
  onCreateRegion: (data: CreateRegion) => Promise<CreateRegionResponse | null>
  writePermitted: boolean
  loadingData: boolean
}

export const CreateRegionDialog: React.FC<CreateRegionDialogProps> = ({
  onCreateRegion,
  writePermitted,
  loadingData,
}) => {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const [createdRegion, setCreatedRegion] = useState<CreateRegionResponse | null>(null)
  const [isProxyApiKeyRevealed, setIsProxyApiKeyRevealed] = useState(false)
  const [isSshGatewayApiKeyRevealed, setIsSshGatewayApiKeyRevealed] = useState(false)
  const [isSnapshotManagerApiKeyRevealed, setIsSnapshotManagerApiKeyRevealed] = useState(false)

  const [formData, setFormData] = useState(DEFAULT_FORM_DATA)

  const handleCreate = async () => {
    setLoading(true)
    try {
      const createRegionData: CreateRegion = {
        name: formData.name,
        proxyUrl: formData.proxyUrl.trim() || null,
        sshGatewayUrl: formData.sshGatewayUrl.trim() || null,
        snapshotManagerUrl: formData.snapshotManagerUrl.trim() || null,
      }

      const region = await onCreateRegion(createRegionData)
      if (region) {
        if (!region.proxyApiKey && !region.sshGatewayApiKey && !region.snapshotManagerApiKey) {
          setOpen(false)
          setCreatedRegion(null)
        } else {
          setCreatedRegion(region)
        }
        setFormData(DEFAULT_FORM_DATA)
      }
    } finally {
      setLoading(false)
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

  if (!writePermitted) {
    return null
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setCreatedRegion(null)
          setFormData(DEFAULT_FORM_DATA)
          setIsProxyApiKeyRevealed(false)
          setIsSshGatewayApiKeyRevealed(false)
          setIsSnapshotManagerApiKeyRevealed(false)
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" disabled={loadingData} className="w-auto px-4" title="Create Region">
          <Plus className="w-4 h-4" />
          Create Region
        </Button>
      </DialogTrigger>

      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create New Region</DialogTitle>
          <DialogDescription>Add a new region for grouping runners and sandboxes.</DialogDescription>
        </DialogHeader>

        {createdRegion &&
        (createdRegion.proxyApiKey || createdRegion.sshGatewayApiKey || createdRegion.snapshotManagerApiKey) ? (
          <div className="space-y-6">
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Region created successfully.{' '}
                {(createdRegion.proxyApiKey || createdRegion.sshGatewayApiKey || createdRegion.snapshotManagerApiKey) &&
                  "Save these API keys securely. You won't be able to see them again."}
              </p>
            </div>

            {createdRegion.proxyApiKey && (
              <div className="space-y-3">
                <Label htmlFor="proxy-api-key">Proxy API Key</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsProxyApiKeyRevealed(true)}
                    onMouseLeave={() => setIsProxyApiKeyRevealed(false)}
                  >
                    {isProxyApiKeyRevealed ? createdRegion.proxyApiKey : getMaskedToken(createdRegion.proxyApiKey)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(createdRegion.proxyApiKey!)}
                  />
                </div>
              </div>
            )}

            {createdRegion.sshGatewayApiKey && (
              <div className="space-y-3">
                <Label htmlFor="ssh-gateway-api-key">SSH Gateway API Key</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsSshGatewayApiKeyRevealed(true)}
                    onMouseLeave={() => setIsSshGatewayApiKeyRevealed(false)}
                  >
                    {isSshGatewayApiKeyRevealed
                      ? createdRegion.sshGatewayApiKey
                      : getMaskedToken(createdRegion.sshGatewayApiKey)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(createdRegion.sshGatewayApiKey!)}
                  />
                </div>
              </div>
            )}

            {createdRegion.snapshotManagerApiKey && (
              <div className="space-y-3">
                <Label htmlFor="snapshot-manager-api-key">Snapshot Manager API Key</Label>
                <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                  <span
                    className="overflow-x-auto pr-2 cursor-text select-all"
                    onMouseEnter={() => setIsSnapshotManagerApiKeyRevealed(true)}
                    onMouseLeave={() => setIsSnapshotManagerApiKeyRevealed(false)}
                  >
                    {isSnapshotManagerApiKeyRevealed
                      ? createdRegion.snapshotManagerApiKey
                      : getMaskedToken(createdRegion.snapshotManagerApiKey)}
                  </span>
                  <Copy
                    className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => copyToClipboard(createdRegion.snapshotManagerApiKey!)}
                  />
                </div>
              </div>
            )}
          </div>
        ) : (
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
                value={formData.name}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, name: e.target.value }))
                }}
                placeholder="us-east-1"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                Region name must contain only letters, numbers, underscores, periods, and hyphens.
              </p>
            </div>

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
              <Label htmlFor="ssh-gateway-url">SSH Gateway URL</Label>
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
              <Label htmlFor="snapshot-manager-url">Snapshot Manager URL</Label>
              <Input
                id="snapshot-manager-url"
                value={formData.snapshotManagerUrl}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, snapshotManagerUrl: e.target.value }))
                }}
                placeholder="https://snapshot-manager.example.com"
              />
              <p className="text-sm text-muted-foreground mt-1 pl-1">
                (Optional) URL of the custom snapshot manager for this region
              </p>
            </div>
          </form>
        )}

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              {createdRegion ? 'Close' : 'Cancel'}
            </Button>
          </DialogClose>
          {!createdRegion &&
            (loading ? (
              <Button type="button" variant="default" disabled>
                Creating...
              </Button>
            ) : (
              <Button type="submit" form="create-region-form" variant="default" disabled={loading}>
                Create
              </Button>
            ))}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
