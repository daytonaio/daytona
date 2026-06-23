/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { getRelativeTimeString } from '@/lib/utils'
import { Region, RegionType } from '@daytona/api-client'
import { ChevronDown, ChevronUp, KeyRound, Pencil, RefreshCw, Trash2, X } from 'lucide-react'
import React, { Ref, useCallback, useImperativeHandle, useState } from 'react'
import { CopyButton } from './CopyButton'
import { InfoRow, InfoSection } from './sandboxes/SandboxInfoPanel'
import { TimestampTooltip } from './TimestampTooltip'

export interface RegionDetailsSheetRef {
  open: () => void
  close: () => void
}

interface RegionDetailsSheetProps {
  region: Region | null
  ref?: Ref<RegionDetailsSheetRef>
  onOpenChange: (open: boolean) => void
  regionIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  hasNext: boolean
  hasPrev: boolean
  onDelete: (region: Region) => void
  onNavigate: (direction: 'prev' | 'next') => void
  onUpdate: (region: Region) => void
  onRegenerateProxyApiKey: (region: Region) => void
  onRegenerateSshGatewayApiKey: (region: Region) => void
  onRegenerateSnapshotManagerCredentials: (region: Region) => void
}

const RegionDetailsSheet: React.FC<RegionDetailsSheetProps> = ({
  region,
  ref,
  onOpenChange,
  regionIsLoading,
  writePermitted,
  deletePermitted,
  hasNext,
  hasPrev,
  onDelete,
  onNavigate,
  onUpdate,
  onRegenerateProxyApiKey,
  onRegenerateSshGatewayApiKey,
  onRegenerateSnapshotManagerCredentials,
}) => {
  const [open, setOpen] = useState(false)

  const handleOpenChange = useCallback(
    (isOpen: boolean) => {
      setOpen(isOpen)
      onOpenChange(isOpen)
    },
    [onOpenChange],
  )

  useImperativeHandle(ref, () => ({
    open: () => handleOpenChange(true),
    close: () => handleOpenChange(false),
  }))

  if (!region) return null

  const isLoading = regionIsLoading[region.id] || false
  const isCustomRegion = region.regionType === RegionType.CUSTOM
  const hasCredentialActions = Boolean(
    isCustomRegion && writePermitted && (region.proxyUrl || region.sshGatewayUrl || region.snapshotManagerUrl),
  )

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetContent className="w-dvw sm:w-[450px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="flex flex-row items-start justify-between p-4 px-5 space-y-0 border-b border-border">
          <div className="min-w-0">
            <SheetTitle>Region Details</SheetTitle>
          </div>
          <div className="flex flex-wrap items-center justify-end shrink-0">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev || isLoading} onClick={() => onNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous region</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext || isLoading} onClick={() => onNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next region</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => handleOpenChange(false)} disabled={isLoading}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="flex flex-col">
            <InfoSection>
              <InfoRow label="Name" className="-mr-2">
                <div className="flex items-center gap-1 min-w-0">
                  <span className="truncate">{region.name}</span>
                  <CopyButton value={region.name} tooltipText="Copy name" size="icon-xs" />
                </div>
              </InfoRow>
              <InfoRow label="UUID" className="-mr-2">
                <div className="flex items-center gap-1 min-w-0">
                  <span className="truncate font-mono text-sm">{region.id}</span>
                  <CopyButton value={region.id} tooltipText="Copy UUID" size="icon-xs" />
                </div>
              </InfoRow>
              <InfoRow label="Type">{isCustomRegion ? 'Custom' : <Badge variant="secondary">Shared</Badge>}</InfoRow>

              {isCustomRegion && (writePermitted || deletePermitted) && (
                <div className="flex justify-end pt-3">
                  <ButtonGroup>
                    {writePermitted && (
                      <Button variant="outline" size="sm" onClick={() => onUpdate(region)} disabled={isLoading}>
                        {isLoading ? <Spinner /> : <Pencil className="size-4" />}
                        Edit
                      </Button>
                    )}
                    {deletePermitted && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onDelete(region)}
                        disabled={isLoading}
                        className="text-destructive-foreground hover:bg-destructive/10 hover:text-destructive-foreground"
                      >
                        {isLoading ? <Spinner /> : <Trash2 className="size-4" />}
                        Delete
                      </Button>
                    )}
                  </ButtonGroup>
                </div>
              )}
            </InfoSection>

            <InfoSection title="Endpoints">
              <EndpointRow label="Proxy URL" value={region.proxyUrl} />
              <EndpointRow label="SSH Gateway URL" value={region.sshGatewayUrl} />
              <EndpointRow label="Snapshot Manager URL" value={region.snapshotManagerUrl} />
            </InfoSection>

            {hasCredentialActions && (
              <InfoSection title="Credentials">
                <div>
                  {region.proxyUrl && (
                    <CredentialActionRow onClick={() => onRegenerateProxyApiKey(region)} disabled={isLoading}>
                      Proxy API Key
                    </CredentialActionRow>
                  )}
                  {region.sshGatewayUrl && (
                    <CredentialActionRow onClick={() => onRegenerateSshGatewayApiKey(region)} disabled={isLoading}>
                      SSH Gateway API Key
                    </CredentialActionRow>
                  )}
                  {region.snapshotManagerUrl && (
                    <CredentialActionRow
                      onClick={() => onRegenerateSnapshotManagerCredentials(region)}
                      disabled={isLoading}
                    >
                      Snapshot Manager Credentials
                    </CredentialActionRow>
                  )}
                </div>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">
                  Regenerating keys will immediately invalidate existing connections using the old credentials.
                </p>
              </InfoSection>
            )}

            <InfoSection title="Activity">
              <InfoRow label="Created">
                <RegionTimestamp timestamp={region.createdAt} />
              </InfoRow>
            </InfoSection>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

function CredentialActionRow({
  children,
  onClick,
  disabled,
}: {
  children: React.ReactNode
  onClick: () => void
  disabled?: boolean
}) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border py-3 last:border-b-0">
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg text-muted-foreground [background:color-mix(in_srgb,currentColor,transparent_90%)]">
          <KeyRound className="size-3.5" />
        </div>
        <span className="min-w-0 break-words text-sm font-medium leading-tight">{children}</span>
      </div>
      <Button variant="secondary" size="sm" onClick={onClick} disabled={disabled} className="shrink-0">
        {disabled ? <Spinner /> : <RefreshCw className="size-4" />}
        Regenerate
      </Button>
    </div>
  )
}

function EndpointRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <InfoRow label={label} className="-mr-2">
      {value ? (
        <div className="flex items-center gap-1 min-w-0">
          <code className="block min-w-0 truncate rounded bg-muted px-1 py-1 font-mono text-[13px] text-foreground">
            {value}
          </code>
          <CopyButton value={value} tooltipText={`Copy ${label}`} size="icon-xs" />
        </div>
      ) : (
        <span className="text-muted-foreground">N/A</span>
      )}
    </InfoRow>
  )
}

function RegionTimestamp({ timestamp }: { timestamp?: string }) {
  return (
    <TimestampTooltip timestamp={timestamp}>
      <span>{getRelativeTimeString(timestamp).relativeTimeString}</span>
    </TimestampTooltip>
  )
}

export default RegionDetailsSheet
