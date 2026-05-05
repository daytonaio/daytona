/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { ResourceChip } from '@/components/ResourceChip'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Badge, type BadgeProps } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { getSnapshotQueryErrorStatus, useSnapshotQuery } from '@/hooks/queries/useSnapshotsQuery'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { SnapshotDto, SnapshotState } from '@daytona/api-client'
import { MagnifyingGlassIcon } from '@phosphor-icons/react'
import { ChevronDown, ChevronUp, CircleAlert, Pause, Play, Trash2, X } from 'lucide-react'
import React from 'react'

export interface SnapshotSheetProps {
  snapshotId?: string | null
  snapshot: SnapshotDto | null
  open: boolean
  onOpenChange: (open: boolean) => void
  getRegionName: (regionId: string) => string | undefined
  onNavigate: (direction: 'prev' | 'next') => void
  hasPrev: boolean
  hasNext: boolean
  actionsDisabled?: boolean
  writePermitted: boolean
  deletePermitted: boolean
  onActivate: (snapshot: SnapshotDto) => void
  onDeactivate: (snapshot: SnapshotDto) => void
  onDelete: (snapshot: SnapshotDto) => void
}

function InfoSection({
  title,
  children,
  className,
}: {
  title?: React.ReactNode
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn('px-5 py-4 border-b border-border last:border-b-0', className)}>
      {title && <div className="text-xs uppercase tracking-widest text-muted-foreground mb-2">{title}</div>}
      {children}
    </div>
  )
}

function InfoRow({
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

function CopyValue({
  value,
  tooltipText = 'Copy',
  className,
}: {
  value: string
  tooltipText?: string
  className?: string
}) {
  return (
    <div className="flex items-center gap-1 min-w-0">
      <span className={cn('truncate', className)}>{value}</span>
      <CopyButton value={value} tooltipText={tooltipText} size="icon-xs" />
    </div>
  )
}

function EmptyValue() {
  return <span className="text-muted-foreground">-</span>
}

function getStateBadgeVariant(state: SnapshotState): BadgeProps['variant'] {
  switch (state) {
    case SnapshotState.ACTIVE:
      return 'success'
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return 'destructive'
    default:
      return 'secondary'
  }
}

function getStateLabel(state: SnapshotState) {
  if (state === SnapshotState.REMOVING) {
    return 'Deleting'
  }

  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function SnapshotStateBadge({ snapshot }: { snapshot: SnapshotDto }) {
  return <Badge variant={getStateBadgeVariant(snapshot.state)}>{getStateLabel(snapshot.state)}</Badge>
}

function TimestampRow({ label, value }: { label: string; value: Date | null | undefined }) {
  if (!value) {
    return (
      <InfoRow label={label}>
        <span className="text-muted-foreground">-</span>
      </InfoRow>
    )
  }

  const timestamp = getRelativeTimeString(value)

  return (
    <InfoRow label={label}>
      <TimestampTooltip timestamp={value.toString()}>{timestamp.relativeTimeString}</TimestampTooltip>
    </InfoRow>
  )
}

function SnapshotSheetSkeleton() {
  const overviewRows = ['name', 'image', 'entrypoint', 'state']
  const timestampRows = ['created', 'updated', 'last-used']

  return (
    <>
      <InfoSection>
        {overviewRows.map((row) => (
          <div key={row} className="flex items-center justify-between gap-3 py-1">
            <Skeleton className="h-4 w-20 shrink-0" />
            <Skeleton className="h-4 w-32" />
          </div>
        ))}
      </InfoSection>
      <InfoSection title="Resources">
        <div className="flex flex-wrap gap-2 py-1">
          <Skeleton className="h-7 w-16" />
          <Skeleton className="h-7 w-20" />
          <Skeleton className="h-7 w-16" />
        </div>
      </InfoSection>
      <InfoSection title="Regions">
        <div className="flex flex-wrap gap-2">
          <Skeleton className="h-6 w-24" />
          <Skeleton className="h-6 w-20" />
        </div>
      </InfoSection>
      <InfoSection title="Timestamps">
        {timestampRows.map((row) => (
          <div key={row} className="flex items-center justify-between gap-3 py-1">
            <Skeleton className="h-4 w-20 shrink-0" />
            <Skeleton className="h-4 w-28" />
          </div>
        ))}
      </InfoSection>
    </>
  )
}

function SnapshotSheetEmptyState({ error }: { error: boolean }) {
  const Icon = error ? CircleAlert : MagnifyingGlassIcon

  return (
    <Empty variant={error ? 'destructive' : 'neutral'} className="min-h-64 border-0 bg-transparent px-5 py-6">
      <EmptyHeader>
        <EmptyMedia
          variant="icon"
          className={cn('[&_svg]:size-4', {
            'bg-destructive-background text-destructive': error,
          })}
        >
          <Icon className="size-4" />
        </EmptyMedia>
        <EmptyTitle>{error ? 'Failed to load snapshot' : 'Snapshot not found'}</EmptyTitle>
        <EmptyDescription>
          {error
            ? 'Something went wrong while fetching this snapshot.'
            : 'This snapshot may have been deleted or you may not have access to it.'}
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function SnapshotSheet({
  snapshotId,
  snapshot,
  open,
  onOpenChange,
  getRegionName,
  onNavigate,
  hasPrev,
  hasNext,
  actionsDisabled = false,
  writePermitted,
  deletePermitted,
  onActivate,
  onDeactivate,
  onDelete,
}: SnapshotSheetProps) {
  const {
    data: fetchedSnapshot,
    isLoading: snapshotIsLoading,
    isFetching: snapshotIsFetching,
    isError: snapshotIsError,
    error: snapshotError,
  } = useSnapshotQuery(snapshotId, {
    enabled: open && !snapshot && !!snapshotId,
  })

  const activeSnapshot = snapshot ?? fetchedSnapshot
  const loadingSnapshot = !activeSnapshot && (snapshotIsLoading || snapshotIsFetching)
  const snapshotNotFound = snapshotIsError && getSnapshotQueryErrorStatus(snapshotError) === 404
  const regionNames = activeSnapshot?.regionIds?.map((id) => getRegionName(id) ?? id) ?? []
  const showActions = !!activeSnapshot && !activeSnapshot.general && (writePermitted || deletePermitted)
  const showActivate = !!activeSnapshot && writePermitted && activeSnapshot.state === SnapshotState.INACTIVE
  const showDeactivate = !!activeSnapshot && writePermitted && activeSnapshot.state === SnapshotState.ACTIVE
  const showDelete = !!activeSnapshot && deletePermitted && activeSnapshot.state !== SnapshotState.REMOVING

  if (!snapshotId && !activeSnapshot) return null

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        showCloseButton={false}
        className="w-dvw p-0 flex flex-col gap-0 sm:w-[450px] [&>button]:hidden"
      >
        <SheetHeader className="flex flex-row items-center justify-between p-4 px-5 space-y-0">
          <div className="min-w-0">
            <SheetTitle className="text-lg font-medium">Snapshot Details</SheetTitle>
          </div>
          <div className="flex items-center justify-end shrink-0">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev} onClick={() => onNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous snapshot</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext} onClick={() => onNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next snapshot</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => onOpenChange(false)}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <Separator />

        <ScrollArea fade="mask" className="min-h-0 flex-1">
          {loadingSnapshot ? (
            <SnapshotSheetSkeleton />
          ) : !activeSnapshot ? (
            <SnapshotSheetEmptyState error={snapshotIsError && !snapshotNotFound} />
          ) : (
            <>
              <InfoSection>
                <InfoRow label="Name" className="-mr-2">
                  <CopyValue value={activeSnapshot.name} tooltipText="Copy name" />
                </InfoRow>
                <InfoRow label="Image" className="-mr-2">
                  {activeSnapshot.imageName ? (
                    <CopyValue value={activeSnapshot.imageName} tooltipText="Copy image" className="font-mono" />
                  ) : activeSnapshot.buildInfo ? (
                    <Badge variant="secondary" className="rounded-sm px-1 font-medium">
                      DECLARATIVE BUILD
                    </Badge>
                  ) : (
                    <EmptyValue />
                  )}
                </InfoRow>
                {activeSnapshot.entrypoint?.length ? (
                  <InfoRow label="Entrypoint" className="items-start -mr-2">
                    <CopyValue
                      value={activeSnapshot.entrypoint.join(' ')}
                      tooltipText="Copy entrypoint"
                      className="font-mono"
                    />
                  </InfoRow>
                ) : null}
                <InfoRow label="State">
                  <SnapshotStateBadge snapshot={activeSnapshot} />
                </InfoRow>
                {activeSnapshot.general && (
                  <InfoRow label="Type">
                    <Badge variant="secondary">System</Badge>
                  </InfoRow>
                )}
                {showActions && (
                  <div className="flex items-center justify-end pt-3">
                    <ButtonGroup>
                      {showActivate && (
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={actionsDisabled}
                          onClick={() => onActivate(activeSnapshot)}
                        >
                          <Play className="size-4" />
                          Activate
                        </Button>
                      )}
                      {showDeactivate && (
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={actionsDisabled}
                          onClick={() => onDeactivate(activeSnapshot)}
                        >
                          <Pause className="size-4" />
                          Deactivate
                        </Button>
                      )}
                      {showDelete && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="outline"
                              size="icon-sm"
                              disabled={actionsDisabled}
                              onClick={() => onDelete(activeSnapshot)}
                              aria-label="Delete snapshot"
                              className="text-destructive-foreground hover:bg-destructive/10 hover:text-destructive-foreground"
                            >
                              <Trash2 className="size-4" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>Delete</TooltipContent>
                        </Tooltip>
                      )}
                    </ButtonGroup>
                  </div>
                )}
              </InfoSection>

              {activeSnapshot.errorReason && (
                <InfoSection title="Error">
                  <p className="text-sm text-destructive-foreground break-words">{activeSnapshot.errorReason}</p>
                </InfoSection>
              )}

              <InfoSection title="Resources">
                <div className="flex flex-wrap gap-2 py-1">
                  <ResourceChip resource="cpu" value={activeSnapshot.cpu} />
                  <ResourceChip resource="memory" value={activeSnapshot.mem} />
                  <ResourceChip resource="disk" value={activeSnapshot.disk} />
                </div>
              </InfoSection>

              <InfoSection title="Regions">
                {regionNames.length ? (
                  <div className="flex flex-wrap gap-2">
                    {regionNames.map((regionName) => (
                      <Badge key={regionName} variant="secondary">
                        {regionName}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <span className="text-sm text-muted-foreground">-</span>
                )}
              </InfoSection>

              {activeSnapshot.buildInfo && (
                <InfoSection title="Build">
                  <TimestampRow label="Created" value={activeSnapshot.buildInfo.createdAt} />
                  <TimestampRow label="Updated" value={activeSnapshot.buildInfo.updatedAt} />
                  {!!activeSnapshot.buildInfo.contextHashes?.length && (
                    <InfoRow label="Context hashes" className="items-start -mr-2">
                      <div className="flex min-w-0 flex-col items-end gap-1">
                        {activeSnapshot.buildInfo.contextHashes.map((hash) => (
                          <CopyValue key={hash} value={hash} tooltipText="Copy hash" className="font-mono" />
                        ))}
                      </div>
                    </InfoRow>
                  )}
                  {activeSnapshot.buildInfo.dockerfileContent && (
                    <div className="mt-3">
                      <div className="mb-1 text-sm text-muted-foreground">Dockerfile</div>
                      <pre className="max-h-64 overflow-auto rounded-md border border-border bg-muted/40 p-3 text-left text-xs">
                        <code>{activeSnapshot.buildInfo.dockerfileContent}</code>
                      </pre>
                    </div>
                  )}
                </InfoSection>
              )}

              <InfoSection title="Timestamps">
                <TimestampRow label="Created" value={activeSnapshot.createdAt} />
                <TimestampRow label="Updated" value={activeSnapshot.updatedAt} />
                <TimestampRow label="Last used" value={activeSnapshot.lastUsedAt} />
              </InfoSection>
            </>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
