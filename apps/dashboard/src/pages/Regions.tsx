/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import {
  Region,
  OrganizationRolePermissionsEnum,
  CreateRegion,
  CreateRegionResponse,
  SnapshotManagerCredentials,
} from '@daytonaio/api-client'
import { RegionTable } from '@/components/RegionTable'
import { CreateRegionDialog } from '@/components/CreateRegionDialog'
import RegionDetailsSheet from '@/components/RegionDetailsSheet'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { useRegions } from '@/hooks/useRegions'
import { getMaskedToken } from '@/lib/utils'
import { Copy } from 'lucide-react'

const Regions: React.FC = () => {
  const { organizationsApi } = useApi()
  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const {
    availableRegions: regions,
    loadingAvailableRegions: loadingRegions,
    refreshAvailableRegions: refreshRegions,
  } = useRegions()

  const [regionIsLoading, setRegionIsLoading] = useState<Record<string, boolean>>({})

  const [regionToDelete, setRegionToDelete] = useState<Region | null>(null)
  const [deleteRegionDialogIsOpen, setDeleteRegionDialogIsOpen] = useState(false)

  // Regenerate API Key state
  const [showRegenerateProxyApiKeyDialog, setShowRegenerateProxyApiKeyDialog] = useState(false)
  const [showRegenerateSshGatewayApiKeyDialog, setShowRegenerateSshGatewayApiKeyDialog] = useState(false)
  const [showRegenerateSnapshotManagerCredsDialog, setShowRegenerateSnapshotManagerCredsDialog] = useState(false)
  const [regeneratedApiKey, setRegeneratedApiKey] = useState<string | null>(null)
  const [regeneratedSnapshotManagerCreds, setRegeneratedSnapshotManagerCreds] =
    useState<SnapshotManagerCredentials | null>(null)
  const [regionForRegenerate, setRegionForRegenerate] = useState<Region | null>(null)
  const [copied, setCopied] = useState(false)
  const [isApiKeyRevealed, setIsApiKeyRevealed] = useState(false)
  const [isSnapshotManagerPasswordRevealed, setIsSnapshotManagerPasswordRevealed] = useState(false)

  // Region Details Sheet state
  const [selectedRegion, setSelectedRegion] = useState<Region | null>(null)
  const [showRegionDetails, setShowRegionDetails] = useState(false)

  const handleCreateRegion = async (createRegionData: CreateRegion): Promise<CreateRegionResponse | null> => {
    if (!selectedOrganization) {
      return null
    }

    try {
      const response = (await organizationsApi.createRegion(createRegionData, selectedOrganization.id)).data
      toast.success(`Creating region ${createRegionData.name}`)
      await refreshRegions()
      return response
    } catch (error) {
      handleApiError(error, 'Failed to create region')
      return null
    }
  }

  const handleDelete = async (region: Region) => {
    if (!selectedOrganization) {
      return
    }

    setRegionIsLoading((prev) => ({ ...prev, [region.id]: true }))

    try {
      await organizationsApi.deleteRegion(region.id, selectedOrganization.id)
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

  const handleRegenerateProxyApiKey = async (region: Region) => {
    setRegionForRegenerate(region)
    setRegeneratedApiKey(null)
    setShowRegenerateProxyApiKeyDialog(true)
  }

  const handleRegenerateSshGatewayApiKey = async (region: Region) => {
    setRegionForRegenerate(region)
    setRegeneratedApiKey(null)
    setShowRegenerateSshGatewayApiKeyDialog(true)
  }

  const handleRegenerateSnapshotManagerCredentials = async (region: Region) => {
    setRegionForRegenerate(region)
    setRegeneratedSnapshotManagerCreds(null)
    setShowRegenerateSnapshotManagerCredsDialog(true)
  }

  const handleOpenRegionDetails = (region: Region) => {
    setSelectedRegion(region)
    setShowRegionDetails(true)
  }

  const confirmRegenerateProxyApiKey = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: true }))

    try {
      const response = await organizationsApi.regenerateProxyApiKey(regionForRegenerate.id, selectedOrganization.id)
      setRegeneratedApiKey(response.data.apiKey)
      setShowRegenerateProxyApiKeyDialog(true)
      toast.success('Proxy API key regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate proxy API key')
      setShowRegenerateProxyApiKeyDialog(false)
      setRegionForRegenerate(null)
    } finally {
      setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: false }))
    }
  }

  const confirmRegenerateSshGatewayApiKey = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: true }))

    try {
      const response = await organizationsApi.regenerateSshGatewayApiKey(
        regionForRegenerate.id,
        selectedOrganization.id,
      )
      setRegeneratedApiKey(response.data.apiKey)
      setShowRegenerateSshGatewayApiKeyDialog(true)
      toast.success('SSH Gateway API key regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate SSH Gateway API key')
      setShowRegenerateSshGatewayApiKeyDialog(false)
      setRegionForRegenerate(null)
    } finally {
      setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: false }))
    }
  }

  const confirmRegenerateSnapshotManagerCredentials = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: true }))

    try {
      const response = await organizationsApi.regenerateSnapshotManagerCredentials(
        regionForRegenerate.id,
        selectedOrganization.id,
      )
      setRegeneratedSnapshotManagerCreds(response.data)
      setShowRegenerateSnapshotManagerCredsDialog(true)
      toast.success('Snapshot Manager credentials regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate Snapshot Manager credentials')
      setShowRegenerateSnapshotManagerCredsDialog(false)
      setRegionForRegenerate(null)
    } finally {
      setRegionIsLoading((prev) => ({ ...prev, [regionForRegenerate.id]: false }))
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Regions</h1>
        <CreateRegionDialog
          onCreateRegion={handleCreateRegion}
          writePermitted={writePermitted}
          loadingData={loadingRegions}
        />
      </div>

      <RegionTable
        data={regions}
        loading={loadingRegions}
        isLoadingRegion={(region) => regionIsLoading[region.id] || false}
        deletePermitted={deletePermitted}
        writePermitted={writePermitted}
        onDelete={(region) => {
          setRegionToDelete(region)
          setDeleteRegionDialogIsOpen(true)
        }}
        onOpenDetails={handleOpenRegionDetails}
      />

      <RegionDetailsSheet
        region={selectedRegion}
        open={showRegionDetails}
        onOpenChange={(open) => {
          setShowRegionDetails(open)
          if (!open) {
            setSelectedRegion(null)
          }
        }}
        regionIsLoading={regionIsLoading}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        onDelete={(region) => {
          setRegionToDelete(region)
          setDeleteRegionDialogIsOpen(true)
          setShowRegionDetails(false)
        }}
        onRegenerateProxyApiKey={handleRegenerateProxyApiKey}
        onRegenerateSshGatewayApiKey={handleRegenerateSshGatewayApiKey}
        onRegenerateSnapshotManagerCredentials={handleRegenerateSnapshotManagerCredentials}
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

      {/* Regenerate Proxy API Key Dialog */}
      <AlertDialog
        open={showRegenerateProxyApiKeyDialog}
        onOpenChange={(isOpen) => {
          setShowRegenerateProxyApiKeyDialog(isOpen)
          if (!isOpen) {
            setRegionForRegenerate(null)
            setRegeneratedApiKey(null)
            setCopied(false)
            setIsApiKeyRevealed(false)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {regeneratedApiKey ? 'Proxy API Key Regenerated' : 'Regenerate Proxy API Key'}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {regeneratedApiKey ? (
                'The new API key has been generated. Copy it now as it will not be shown again.'
              ) : (
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current proxy API key. The proxy will
                  need to be redeployed with the new API key.
                </>
              )}
              {regeneratedApiKey && (
                <div className="space-y-4 mt-4">
                  <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                    <span
                      className="overflow-x-auto pr-2 cursor-text select-all"
                      onMouseEnter={() => setIsApiKeyRevealed(true)}
                      onMouseLeave={() => setIsApiKeyRevealed(false)}
                    >
                      {isApiKeyRevealed ? regeneratedApiKey : getMaskedToken(regeneratedApiKey)}
                    </span>
                    <Copy
                      className="w-4 h-4 cursor-pointer flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
                      onClick={() => copyToClipboard(regeneratedApiKey)}
                    />
                  </div>
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogFooter>
            {!regeneratedApiKey ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateProxyApiKey}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] ? 'Regenerating...' : 'Regenerate'}
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateProxyApiKeyDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedApiKey(null)
                  setCopied(false)
                  setIsApiKeyRevealed(false)
                }}
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
              >
                Close
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Regenerate SSH Gateway API Key Dialog */}
      <AlertDialog
        open={showRegenerateSshGatewayApiKeyDialog}
        onOpenChange={(isOpen) => {
          setShowRegenerateSshGatewayApiKeyDialog(isOpen)
          if (!isOpen) {
            setRegionForRegenerate(null)
            setRegeneratedApiKey(null)
            setCopied(false)
            setIsApiKeyRevealed(false)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {regeneratedApiKey ? 'SSH Gateway API Key Regenerated' : 'Regenerate SSH Gateway API Key'}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {regeneratedApiKey ? (
                'The new API key has been generated. Copy it now as it will not be shown again.'
              ) : (
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current SSH gateway API key. The SSH
                  gateway will need to be redeployed with the new API key.
                </>
              )}
              {regeneratedApiKey && (
                <div className="space-y-4 mt-4">
                  <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                    <span
                      className="overflow-x-auto pr-2 cursor-text select-all"
                      onMouseEnter={() => setIsApiKeyRevealed(true)}
                      onMouseLeave={() => setIsApiKeyRevealed(false)}
                    >
                      {isApiKeyRevealed ? regeneratedApiKey : getMaskedToken(regeneratedApiKey)}
                    </span>
                    <Copy
                      className="w-4 h-4 cursor-pointer flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
                      onClick={() => copyToClipboard(regeneratedApiKey)}
                    />
                  </div>
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogFooter>
            {!regeneratedApiKey ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateSshGatewayApiKey}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] ? 'Regenerating...' : 'Regenerate'}
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateSshGatewayApiKeyDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedApiKey(null)
                  setCopied(false)
                  setIsApiKeyRevealed(false)
                }}
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
              >
                Close
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Regenerate Snapshot Manager Credentials Dialog */}
      <AlertDialog
        open={showRegenerateSnapshotManagerCredsDialog}
        onOpenChange={(isOpen) => {
          setShowRegenerateSnapshotManagerCredsDialog(isOpen)
          if (!isOpen) {
            setRegionForRegenerate(null)
            setRegeneratedSnapshotManagerCreds(null)
            setCopied(false)
            setIsSnapshotManagerPasswordRevealed(false)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {regeneratedSnapshotManagerCreds
                ? 'Snapshot Manager Credentials Regenerated'
                : 'Regenerate Snapshot Manager Credentials'}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {regeneratedSnapshotManagerCreds ? (
                'The new credentials have been generated. Copy them now as they will not be shown again.'
              ) : (
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current Snapshot Manager credentials.
                  The Snapshot Manager will need to be reconfigured with the new credentials.
                </>
              )}
              {regeneratedSnapshotManagerCreds && (
                <div className="space-y-4 mt-4">
                  <div>
                    <span className="text-xs text-muted-foreground">Username</span>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span className="overflow-x-auto pr-2 cursor-text select-all">
                        {regeneratedSnapshotManagerCreds.username}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(regeneratedSnapshotManagerCreds.username)}
                      />
                    </div>
                  </div>
                  <div>
                    <span className="text-xs text-muted-foreground">Password</span>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span
                        className="overflow-x-auto pr-2 cursor-text select-all"
                        onMouseEnter={() => setIsSnapshotManagerPasswordRevealed(true)}
                        onMouseLeave={() => setIsSnapshotManagerPasswordRevealed(false)}
                      >
                        {isSnapshotManagerPasswordRevealed
                          ? regeneratedSnapshotManagerCreds.password
                          : getMaskedToken(regeneratedSnapshotManagerCreds.password)}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(regeneratedSnapshotManagerCreds.password)}
                      />
                    </div>
                  </div>
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogFooter>
            {!regeneratedSnapshotManagerCreds ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateSnapshotManagerCredentials}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] ? 'Regenerating...' : 'Regenerate'}
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateSnapshotManagerCredsDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedSnapshotManagerCreds(null)
                  setCopied(false)
                  setIsSnapshotManagerPasswordRevealed(false)
                }}
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
              >
                Close
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

export default Regions
