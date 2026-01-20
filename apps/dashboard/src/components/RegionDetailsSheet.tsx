/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { formatTimestamp } from '@/lib/utils'
import { Region, RegionType } from '@daytonaio/api-client'
import { Copy, Info, Trash, X } from 'lucide-react'
import React from 'react'
import { toast } from 'sonner'

interface RegionDetailsSheetProps {
  region: Region | null
  open: boolean
  onOpenChange: (open: boolean) => void
  regionIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  onDelete: (region: Region) => void
  onRegenerateProxyApiKey: (region: Region) => void
  onRegenerateSshGatewayApiKey: (region: Region) => void
  onRegenerateSnapshotManagerCredentials: (region: Region) => void
}

const RegionDetailsSheet: React.FC<RegionDetailsSheetProps> = ({
  region,
  open,
  onOpenChange,
  regionIsLoading,
  writePermitted,
  deletePermitted,
  onDelete,
  onRegenerateProxyApiKey,
  onRegenerateSshGatewayApiKey,
  onRegenerateSnapshotManagerCredentials,
}) => {
  if (!region) return null

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  const isLoading = regionIsLoading[region.id] || false
  const isCustomRegion = region.regionType === RegionType.CUSTOM

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[800px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="space-y-0 flex flex-row justify-between items-center p-6">
          <SheetTitle className="text-2xl font-medium">Region Details</SheetTitle>
          <div className="flex gap-2 items-center">
            {deletePermitted && isCustomRegion && (
              <Button variant="outline" className="w-8 h-8" onClick={() => onDelete(region)} disabled={isLoading}>
                <Trash className="w-4 h-4" />
              </Button>
            )}
            <Button variant="outline" className="w-8 h-8" onClick={() => onOpenChange(false)} disabled={isLoading}>
              <X className="w-4 h-4" />
            </Button>
          </div>
        </SheetHeader>

        <div className="flex-1 p-6 space-y-10 overflow-y-auto min-h-0">
          <div className="grid grid-cols-2 gap-6">
            <div>
              <h3 className="text-sm text-muted-foreground">Name</h3>
              <div className="mt-1 flex items-center gap-2">
                <p className="text-sm font-medium truncate">{region.name}</p>
                <button
                  onClick={() => copyToClipboard(region.name)}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                  aria-label="Copy name"
                >
                  <Copy className="w-3 h-3" />
                </button>
              </div>
            </div>
            <div>
              <h3 className="text-sm text-muted-foreground">ID</h3>
              <div className="mt-1 flex items-center gap-2">
                <p className="text-sm font-medium truncate">{region.id}</p>
                <button
                  onClick={() => copyToClipboard(region.id)}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                  aria-label="Copy ID"
                >
                  <Copy className="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-6">
            <div>
              <h3 className="text-sm text-muted-foreground">Created</h3>
              <p className="mt-1 text-sm font-medium">{formatTimestamp(region.createdAt)}</p>
            </div>
            <div>
              <h3 className="text-sm text-muted-foreground">Type</h3>
              <p className="mt-1 text-sm font-medium">{region.regionType}</p>
            </div>
          </div>

          <div>
            <h3 className="text-lg font-medium">URLs</h3>
            <div className="mt-3 space-y-4">
              <div>
                <h4 className="text-sm text-muted-foreground">Proxy URL</h4>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{region.proxyUrl || '-'}</p>
                  {region.proxyUrl && (
                    <button
                      onClick={() => copyToClipboard(region.proxyUrl || '')}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                      aria-label="Copy Proxy URL"
                    >
                      <Copy className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
              <div>
                <h4 className="text-sm text-muted-foreground">SSH Gateway URL</h4>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{region.sshGatewayUrl || '-'}</p>
                  {region.sshGatewayUrl && (
                    <button
                      onClick={() => copyToClipboard(region.sshGatewayUrl || '')}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                      aria-label="Copy SSH Gateway URL"
                    >
                      <Copy className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
              <div>
                <h4 className="text-sm text-muted-foreground">Snapshot Manager URL</h4>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{region.snapshotManagerUrl || '-'}</p>
                  {region.snapshotManagerUrl && (
                    <button
                      onClick={() => copyToClipboard(region.snapshotManagerUrl || '')}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                      aria-label="Copy Snapshot Manager URL"
                    >
                      <Copy className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
            </div>
          </div>

          {isCustomRegion &&
            writePermitted &&
            (region.proxyUrl || region.sshGatewayUrl || region.snapshotManagerUrl) && (
              <div>
                <h3 className="text-lg font-medium">Credentials</h3>
                <div className="mt-3 space-y-3">
                  {region.proxyUrl && (
                    <Button
                      variant="outline"
                      onClick={() => onRegenerateProxyApiKey(region)}
                      disabled={isLoading}
                      className="w-full justify-start"
                    >
                      <Info className="w-4 h-4 mr-2" />
                      Regenerate Proxy API Key
                    </Button>
                  )}
                  {region.sshGatewayUrl && (
                    <Button
                      variant="outline"
                      onClick={() => onRegenerateSshGatewayApiKey(region)}
                      disabled={isLoading}
                      className="w-full justify-start"
                    >
                      <Info className="w-4 h-4 mr-2" />
                      Regenerate SSH Gateway API Key
                    </Button>
                  )}
                  {region.snapshotManagerUrl && (
                    <Button
                      variant="outline"
                      onClick={() => onRegenerateSnapshotManagerCredentials(region)}
                      disabled={isLoading}
                      className="w-full justify-start"
                    >
                      <Info className="w-4 h-4 mr-2" />
                      Regenerate Snapshot Manager Credentials
                    </Button>
                  )}
                </div>
              </div>
            )}
        </div>
      </SheetContent>
    </Sheet>
  )
}

export default RegionDetailsSheet
