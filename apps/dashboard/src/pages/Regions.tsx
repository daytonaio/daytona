/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageFooter, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import RegionDetailsSheet, { type RegionDetailsSheetRef } from '@/components/RegionDetailsSheet'
import { RegionTable } from '@/components/RegionTable'
import { UpsertRegionSheet } from '@/components/UpsertRegionSheet'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Spinner } from '@/components/ui/spinner'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useDeleteRegionMutation } from '@/hooks/mutations/useDeleteRegionMutation'
import { useMutatingRegions } from '@/hooks/mutations/useMutatingRegions'
import { useRegenerateRegionProxyApiKeyMutation } from '@/hooks/mutations/useRegenerateRegionProxyApiKeyMutation'
import { useRegenerateRegionSnapshotManagerCredentialsMutation } from '@/hooks/mutations/useRegenerateRegionSnapshotManagerCredentialsMutation'
import { useRegenerateRegionSshGatewayApiKeyMutation } from '@/hooks/mutations/useRegenerateRegionSshGatewayApiKeyMutation'
import { useAvailableRegionsQuery } from '@/hooks/queries/useRegionsQuery'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { EMPTY_REGIONS } from '@/lib/regions'
import { getMaskedToken } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, Region, SnapshotManagerCredentials } from '@daytona/api-client'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon, InfoIcon, PlusIcon } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

const Regions: React.FC = () => {
  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const { data: regions = EMPTY_REGIONS, isLoading: loadingRegions } = useAvailableRegionsQuery(
    selectedOrganization?.id,
  )
  const deleteRegionMutation = useDeleteRegionMutation()
  const regenerateProxyApiKeyMutation = useRegenerateRegionProxyApiKeyMutation()
  const regenerateSshGatewayApiKeyMutation = useRegenerateRegionSshGatewayApiKeyMutation()
  const regenerateSnapshotManagerCredentialsMutation = useRegenerateRegionSnapshotManagerCredentialsMutation()
  const mutatingRegionIds = useMutatingRegions()

  const regionIsLoading = useMemo<Record<string, boolean>>(() => {
    return Object.fromEntries([...mutatingRegionIds].map((regionId) => [regionId, true]))
  }, [mutatingRegionIds])

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

  // Region Details Sheet state
  const [selectedRegion, setSelectedRegion] = useState<Region | null>(null)
  const [showRegionDetails, setShowRegionDetails] = useState(false)
  const regionDetailsSheetRef = useRef<RegionDetailsSheetRef>(null)

  // Edit Region Sheet state
  const [showEditRegionSheet, setShowEditRegionSheet] = useState(false)
  const [regionToUpdate, setRegionToUpdate] = useState<Region | null>(null)
  const upsertRegionSheetRef = useRef<{ open: () => void }>(null)

  const handleDelete = async (region: Region) => {
    try {
      await deleteRegionMutation.mutateAsync({
        regionId: region.id,
        organizationId: selectedOrganization?.id,
      })
      setRegionToDelete(null)
      setDeleteRegionDialogIsOpen(false)
      toast.success(`Deleting region ${region.name}`)
    } catch (error) {
      handleApiError(error, 'Failed to delete region')
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

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!writePermitted) {
      return []
    }

    return [
      {
        id: 'create-region',
        label: 'Create Region',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => upsertRegionSheetRef.current?.open(),
      },
    ]
  }, [writePermitted])

  useRegisterCommands(rootCommands, { groupId: 'region-actions', groupLabel: 'Region actions', groupOrder: 0 })

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
    regionDetailsSheetRef.current?.open()
  }

  const handleOpenUpdateDialog = (region: Region) => {
    setRegionToUpdate(region)
    setShowEditRegionSheet(true)
    regionDetailsSheetRef.current?.close()
  }

  useEffect(() => {
    if (!selectedRegion) return

    const found = regions.find((region) => region.id === selectedRegion.id)
    if (!found) {
      regionDetailsSheetRef.current?.close()
      return
    }

    setSelectedRegion(found)
  }, [regions, selectedRegion])

  const selectedRegionIndex = useMemo(() => {
    if (!selectedRegion) {
      return -1
    }

    return regions.findIndex((region) => region.id === selectedRegion.id)
  }, [regions, selectedRegion])

  const handleNavigateRegion = useCallback(
    (direction: 'prev' | 'next') => {
      if (selectedRegionIndex === -1) {
        return
      }

      const nextIndex = direction === 'prev' ? selectedRegionIndex - 1 : selectedRegionIndex + 1
      const nextRegion = regions[nextIndex]
      if (nextRegion) {
        setSelectedRegion(nextRegion)
      }
    },
    [regions, selectedRegionIndex],
  )

  const confirmRegenerateProxyApiKey = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    try {
      const response = await regenerateProxyApiKeyMutation.mutateAsync({
        regionId: regionForRegenerate.id,
        organizationId: selectedOrganization.id,
      })
      setRegeneratedApiKey(response.apiKey)
      setShowRegenerateProxyApiKeyDialog(true)
      toast.success('Proxy API key regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate proxy API key')
      setShowRegenerateProxyApiKeyDialog(false)
      setRegionForRegenerate(null)
    }
  }

  const confirmRegenerateSshGatewayApiKey = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    try {
      const response = await regenerateSshGatewayApiKeyMutation.mutateAsync({
        regionId: regionForRegenerate.id,
        organizationId: selectedOrganization.id,
      })
      setRegeneratedApiKey(response.apiKey)
      setShowRegenerateSshGatewayApiKeyDialog(true)
      toast.success('SSH Gateway API key regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate SSH Gateway API key')
      setShowRegenerateSshGatewayApiKeyDialog(false)
      setRegionForRegenerate(null)
    }
  }

  const confirmRegenerateSnapshotManagerCredentials = async () => {
    if (!regionForRegenerate || !selectedOrganization) {
      return
    }

    try {
      const response = await regenerateSnapshotManagerCredentialsMutation.mutateAsync({
        regionId: regionForRegenerate.id,
        organizationId: selectedOrganization.id,
      })
      setRegeneratedSnapshotManagerCreds(response)
      setShowRegenerateSnapshotManagerCredsDialog(true)
      toast.success('Snapshot Manager credentials regenerated successfully')
    } catch (error) {
      handleApiError(error, 'Failed to regenerate Snapshot Manager credentials')
      setShowRegenerateSnapshotManagerCredsDialog(false)
      setRegionForRegenerate(null)
    }
  }

  return (
    <PageLayout contained>
      <PageHeader />

      <PageContent size="full" className="overflow-hidden">
        <PageIntro
          title="Regions"
          actions={
            writePermitted ? <UpsertRegionSheet disabled={loadingRegions} ref={upsertRegionSheetRef} /> : undefined
          }
        />
        <RegionTable
          data={regions}
          loading={loadingRegions}
          activeRegionId={showRegionDetails ? selectedRegion?.id : undefined}
          isLoadingRegion={(region) => regionIsLoading[region.id] || false}
          deletePermitted={deletePermitted}
          writePermitted={writePermitted}
          onDelete={(region) => {
            setRegionToDelete(region)
            setDeleteRegionDialogIsOpen(true)
          }}
          onOpenDetails={handleOpenRegionDetails}
          onUpdate={handleOpenUpdateDialog}
        />
      </PageContent>
      <PageFooter />

      <RegionDetailsSheet
        ref={regionDetailsSheetRef}
        region={selectedRegion}
        onOpenChange={setShowRegionDetails}
        regionIsLoading={regionIsLoading}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        hasPrev={selectedRegionIndex > 0}
        hasNext={selectedRegionIndex >= 0 && selectedRegionIndex < regions.length - 1}
        onNavigate={handleNavigateRegion}
        onDelete={(region) => {
          setRegionToDelete(region)
          setDeleteRegionDialogIsOpen(true)
        }}
        onUpdate={handleOpenUpdateDialog}
        onRegenerateProxyApiKey={handleRegenerateProxyApiKey}
        onRegenerateSshGatewayApiKey={handleRegenerateSshGatewayApiKey}
        onRegenerateSnapshotManagerCredentials={handleRegenerateSnapshotManagerCredentials}
      />

      {regionToUpdate && (
        <UpsertRegionSheet
          mode="edit"
          trigger={null}
          region={regionToUpdate}
          open={showEditRegionSheet}
          onOpenChange={(isOpen) => {
            setShowEditRegionSheet(isOpen)
            if (!isOpen) setRegionToUpdate(null)
          }}
        />
      )}

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
                {regionIsLoading[regionToDelete.id] && <Spinner />}
                Delete
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
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {regeneratedApiKey ? 'Proxy API Key Regenerated' : 'Regenerate Proxy API Key'}
            </AlertDialogTitle>
            {regeneratedApiKey ? (
              <AlertDialogDescription className="sr-only">
                The new API key has been generated. Copy it now as it will not be shown again.
              </AlertDialogDescription>
            ) : (
              <AlertDialogDescription>
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current proxy API key. The proxy will
                  need to be redeployed with the new API key.
                </>
              </AlertDialogDescription>
            )}
          </AlertDialogHeader>
          {regeneratedApiKey && (
            <RegeneratedApiKeyDisplay id="regenerated-proxy-api-key" apiKey={regeneratedApiKey} label="Proxy API Key" />
          )}

          <AlertDialogFooter>
            {!regeneratedApiKey ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateProxyApiKey}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] && <Spinner />}
                  Regenerate
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateProxyApiKeyDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedApiKey(null)
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
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {regeneratedApiKey ? 'SSH Gateway API Key Regenerated' : 'Regenerate SSH Gateway API Key'}
            </AlertDialogTitle>
            {regeneratedApiKey ? (
              <AlertDialogDescription className="sr-only">
                The new API key has been generated. Copy it now as it will not be shown again.
              </AlertDialogDescription>
            ) : (
              <AlertDialogDescription>
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current SSH gateway API key. The SSH
                  gateway will need to be redeployed with the new API key.
                </>
              </AlertDialogDescription>
            )}
          </AlertDialogHeader>
          {regeneratedApiKey && (
            <RegeneratedApiKeyDisplay
              id="regenerated-ssh-gateway-api-key"
              apiKey={regeneratedApiKey}
              label="SSH Gateway API Key"
            />
          )}

          <AlertDialogFooter>
            {!regeneratedApiKey ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateSshGatewayApiKey}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] && <Spinner />}
                  Regenerate
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateSshGatewayApiKeyDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedApiKey(null)
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
            {regeneratedSnapshotManagerCreds ? (
              <AlertDialogDescription className="sr-only">
                The new credentials have been generated. Copy them now as they will not be shown again.
              </AlertDialogDescription>
            ) : (
              <AlertDialogDescription>
                <>
                  <strong>Warning:</strong> This will immediately invalidate the current Snapshot Manager credentials.
                  The Snapshot Manager will need to be reconfigured with the new credentials.
                </>
              </AlertDialogDescription>
            )}
          </AlertDialogHeader>
          {regeneratedSnapshotManagerCreds && (
            <RegeneratedSnapshotManagerCredentialsDisplay credentials={regeneratedSnapshotManagerCreds} />
          )}

          <AlertDialogFooter>
            {!regeneratedSnapshotManagerCreds ? (
              <>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={confirmRegenerateSnapshotManagerCredentials}
                  disabled={!regionForRegenerate || regionIsLoading[regionForRegenerate?.id || '']}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                >
                  {regionForRegenerate && regionIsLoading[regionForRegenerate.id] && <Spinner />}
                  Regenerate
                </AlertDialogAction>
              </>
            ) : (
              <AlertDialogAction
                onClick={() => {
                  setShowRegenerateSnapshotManagerCredsDialog(false)
                  setRegionForRegenerate(null)
                  setRegeneratedSnapshotManagerCreds(null)
                }}
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
              >
                Close
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageLayout>
  )
}

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

function OneTimeSecretAlert({ children }: { children: React.ReactNode }) {
  return (
    <Alert variant="warning">
      <InfoIcon />
      <AlertDescription>{children}</AlertDescription>
    </Alert>
  )
}

function CopyableCredentialField({
  id,
  label,
  value,
  revealable = false,
  masked = false,
}: {
  id: string
  label: string
  value: string
  revealable?: boolean
  masked?: boolean
}) {
  const [copiedText, copyToClipboard] = useCopyToClipboard()
  const [revealed, setRevealed] = useState(false)
  const displayValue = masked && !revealed ? getMaskedToken(value) : value

  return (
    <Field>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>

      <InputGroup className="pr-1 flex-1">
        <InputGroupInput id={id} value={displayValue} readOnly />
        {revealable && (
          <InputGroupButton
            variant="ghost"
            size="icon-xs"
            aria-label={revealed ? `Hide ${label}` : `Show ${label}`}
            aria-pressed={revealed}
            onClick={() => setRevealed(!revealed)}
          >
            {revealed ? <EyeOffIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
          </InputGroupButton>
        )}
        <InputGroupButton
          variant="ghost"
          size="icon-xs"
          aria-label={`Copy ${label}`}
          onClick={() => copyToClipboard(value)}
        >
          <AnimatePresence initial={false} mode="wait">
            {copiedText === value ? (
              <MotionCheckIcon className="h-4 w-4" key="copied" {...iconProps} />
            ) : (
              <MotionCopyIcon className="h-4 w-4" key="copy" {...iconProps} />
            )}
          </AnimatePresence>
        </InputGroupButton>
      </InputGroup>
    </Field>
  )
}

function RegeneratedApiKeyDisplay({ id, apiKey, label }: { id: string; apiKey: string; label: string }) {
  return (
    <div className="space-y-6">
      <OneTimeSecretAlert>You can only view this key once. Store it safely.</OneTimeSecretAlert>
      <FieldGroup className="gap-4">
        <CopyableCredentialField id={id} label={label} value={apiKey} revealable masked />
      </FieldGroup>
    </div>
  )
}

function RegeneratedSnapshotManagerCredentialsDisplay({ credentials }: { credentials: SnapshotManagerCredentials }) {
  return (
    <div className="space-y-6">
      <OneTimeSecretAlert>You can only view these credentials once. Store them safely.</OneTimeSecretAlert>
      <FieldGroup className="gap-4">
        <CopyableCredentialField
          id="regenerated-snapshot-manager-username"
          label="Username"
          value={credentials.username}
        />
        <CopyableCredentialField
          id="regenerated-snapshot-manager-password"
          label="Password"
          value={credentials.password}
          revealable
          masked
        />
      </FieldGroup>
    </div>
  )
}

export default Regions
