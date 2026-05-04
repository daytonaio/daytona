/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { ResourceChip } from '@/components/ResourceChip'
import { SandboxLabel } from '@/components/SandboxLabel'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { cn, formatDuration, getRelativeTimeString } from '@/lib/utils'
import { Sandbox } from '@daytona/api-client'
import { AlertCircle, ArrowUpRight, KeyRound, Tag, UserRoundX } from 'lucide-react'
import React, { ReactNode, useMemo } from 'react'

export function InfoSection({
  title,
  children,
  className,
}: {
  title?: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <div className={cn('px-5 py-4 border-b border-border last:border-b-0', className)}>
      {title && <div className="text-xs uppercase tracking-widest text-muted-foreground mb-2">{title}</div>}
      {children}
    </div>
  )
}

export function InfoRow({
  label,
  children,
  className,
}: {
  label: React.ReactNode
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn('flex items-center justify-between gap-3 py-1', className)}>
      <span className="text-sm text-muted-foreground shrink-0">{label}</span>
      <div className="min-w-0 text-sm text-right">{children}</div>
    </div>
  )
}

interface SandboxInfoPanelProps {
  sandbox: Sandbox
  getRegionName: (id: string) => string | undefined
  actionsDisabled?: boolean
  writePermitted?: boolean
  onCreateSshAccess?: () => void
  onRevokeSshAccess?: () => void
  onScreenRecordings?: () => void
}

export function SandboxInfoPanel({
  sandbox,
  getRegionName,
  actionsDisabled = false,
  writePermitted = false,
  onCreateSshAccess,
  onRevokeSshAccess,
  onScreenRecordings,
}: SandboxInfoPanelProps) {
  const labelEntries = useMemo(() => {
    return sandbox.labels ? Object.entries(sandbox.labels) : []
  }, [sandbox.labels])
  const showRecordingsSection = !!onScreenRecordings
  const showSshSection = !!onCreateSshAccess || !!onRevokeSshAccess

  return (
    <div className="flex flex-col">
      {sandbox.errorReason && (
        <div className="px-5 pt-4">
          <Alert variant={sandbox.recoverable ? 'warning' : 'destructive'}>
            <AlertCircle />
            <AlertDescription>{sandbox.errorReason}</AlertDescription>
          </Alert>
        </div>
      )}

      <InfoSection title={null}>
        <InfoRow label="Region" className="-mr-2">
          <div className="flex items-center gap-1">
            <span className="truncate">{getRegionName(sandbox.target) ?? sandbox.target}</span>
            <CopyButton value={sandbox.target} tooltipText="Copy" size="icon-xs" />
          </div>
        </InfoRow>
        <InfoRow label="Snapshot" className="-mr-2">
          {sandbox.snapshot ? (
            <div className="flex items-center gap-1 min-w-0">
              <span className="truncate font-mono text-sm">{sandbox.snapshot}</span>
              <CopyButton value={sandbox.snapshot} tooltipText="Copy" size="icon-xs" />
            </div>
          ) : (
            <span className="text-muted-foreground font-normal">—</span>
          )}
        </InfoRow>
        <InfoRow label="Preview access">
          {sandbox.public ? 'Public' : <span className="text-muted-foreground font-normal">Private</span>}
        </InfoRow>
      </InfoSection>

      <InfoSection title="Resources">
        <div className="flex flex-wrap gap-2 py-1">
          <ResourceChip resource="cpu" value={sandbox.cpu} />
          <ResourceChip resource="memory" value={sandbox.memory} />
          <ResourceChip resource="disk" value={sandbox.disk} />
        </div>
      </InfoSection>

      <InfoSection title="Lifecycle">
        <InfoRow label="Auto-stop">
          {sandbox.autoStopInterval ? (
            formatDuration(sandbox.autoStopInterval)
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
        <InfoRow label="Auto-archive">
          {sandbox.autoArchiveInterval ? (
            formatDuration(sandbox.autoArchiveInterval)
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
        <InfoRow label="Auto-delete">
          {sandbox.autoDeleteInterval !== undefined && sandbox.autoDeleteInterval >= 0 ? (
            sandbox.autoDeleteInterval === 0 ? (
              'On stop'
            ) : (
              formatDuration(sandbox.autoDeleteInterval)
            )
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
      </InfoSection>

      {showSshSection && (
        <div className="px-5 py-3 border-b border-border">
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs uppercase tracking-widest text-muted-foreground">SSH Access</span>
            <ButtonGroup>
              {onCreateSshAccess && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onCreateSshAccess}
                  disabled={actionsDisabled || !writePermitted}
                >
                  <KeyRound className="size-4" />
                  Create
                </Button>
              )}
              {onRevokeSshAccess && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onRevokeSshAccess}
                  disabled={actionsDisabled || !writePermitted}
                >
                  <UserRoundX className="size-4" />
                  Revoke
                </Button>
              )}
            </ButtonGroup>
          </div>
        </div>
      )}

      <InfoSection title="Labels">
        {labelEntries.length > 0 ? (
          <div className="max-h-[250px] overflow-y-auto scrollbar-sm">
            <div className="flex flex-wrap gap-2 py-1">
              {labelEntries.map(([key, value]) => (
                <SandboxLabel key={key} labelKey={key} value={value} />
              ))}
            </div>
          </div>
        ) : (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Tag className="size-4" />
              </EmptyMedia>
              <EmptyDescription>No labels</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </InfoSection>

      {showRecordingsSection && (
        <div className="px-5 py-3 border-b border-border">
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs uppercase tracking-widest text-muted-foreground">Recordings</span>
            <Button
              variant="link"
              className="h-auto px-0 py-0 text-sm"
              onClick={onScreenRecordings}
              disabled={actionsDisabled}
            >
              View
              <ArrowUpRight className="size-4" />
            </Button>
          </div>
        </div>
      )}

      <InfoSection title="Activity">
        <InfoRow label="Created">
          <TimestampTooltip timestamp={sandbox.createdAt}>
            <span>{getRelativeTimeString(sandbox.createdAt).relativeTimeString}</span>
          </TimestampTooltip>
        </InfoRow>
        <InfoRow label="Last event">
          <TimestampTooltip timestamp={sandbox.updatedAt}>
            <span>{getRelativeTimeString(sandbox.updatedAt).relativeTimeString}</span>
          </TimestampTooltip>
        </InfoRow>
      </InfoSection>
    </div>
  )
}

export function InfoPanelSkeleton() {
  return (
    <div className="flex flex-col">
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-16 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-12" />
            <Skeleton className="h-4 w-20" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-20 mb-3" />
        <div className="flex gap-2">
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-16 rounded-full" />
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-18 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-22" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-14 mb-3" />
        <Skeleton className="h-4 w-full" />
      </div>
      <div className="px-5 py-4">
        <Skeleton className="h-2.5 w-24 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-24" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
      </div>
    </div>
  )
}
