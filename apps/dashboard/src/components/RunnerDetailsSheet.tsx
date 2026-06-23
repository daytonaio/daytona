/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { getRelativeTimeString } from '@/lib/utils'
import { Runner, RunnerState } from '@daytona/api-client'
import { ChevronDown, ChevronUp, CircleHelp, Trash2, X } from 'lucide-react'
import React, { Ref, useCallback, useImperativeHandle, useState } from 'react'
import { CopyButton } from './CopyButton'
import { ResourceChip } from './ResourceChip'
import { InfoRow, InfoSection } from './sandboxes/SandboxInfoPanel'
import { TimestampTooltip } from './TimestampTooltip'
import QuotaLine from './QuotaLine'
import { Badge, type BadgeProps } from './ui/badge'

export interface RunnerDetailsSheetRef {
  open: () => void
  close: () => void
}

interface RunnerDetailsSheetProps {
  runner: Runner | null
  ref?: Ref<RunnerDetailsSheetRef>
  onOpenChange: (open: boolean) => void
  runnerIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  hasNext: boolean
  hasPrev: boolean
  onDelete: (runner: Runner) => void
  onNavigate: (direction: 'prev' | 'next') => void
  onToggleEnabled: (runner: Runner) => void
  getRegionName: (regionId: string) => string | undefined
}

const RunnerDetailsSheet: React.FC<RunnerDetailsSheetProps> = ({
  runner,
  ref,
  onOpenChange,
  runnerIsLoading,
  writePermitted,
  deletePermitted,
  hasNext,
  hasPrev,
  onDelete,
  onNavigate,
  onToggleEnabled,
  getRegionName,
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

  if (!runner) return null

  const isLoading = runnerIsLoading[runner.id] || false

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetContent className="h-dvh max-h-dvh w-dvw sm:w-[450px] p-0 flex flex-col gap-0 overflow-hidden [&>button]:hidden">
        <SheetHeader className="flex flex-row items-start justify-between p-4 px-5 space-y-0 border-b border-border">
          <div className="min-w-0">
            <SheetTitle>Runner Details</SheetTitle>
          </div>
          <div className="flex flex-wrap items-center justify-end shrink-0">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev || isLoading} onClick={() => onNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous runner</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext || isLoading} onClick={() => onNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next runner</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => handleOpenChange(false)} disabled={isLoading}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <ScrollArea fade="mask" fadeOffset={30} className="min-h-0 flex-1">
          <div className="flex flex-col">
            <InfoSection>
              <InfoRow label="Name" className="-mr-2">
                <div className="flex items-center gap-1 min-w-0">
                  <span className="truncate">{runner.name}</span>
                  <CopyButton value={runner.name} tooltipText="Copy name" size="icon-xs" />
                </div>
              </InfoRow>
              <InfoRow label="UUID" className="-mr-2">
                <div className="flex items-center gap-1 min-w-0">
                  <span className="truncate font-mono text-sm">{runner.id}</span>
                  <CopyButton value={runner.id} tooltipText="Copy UUID" size="icon-xs" />
                </div>
              </InfoRow>
              <InfoRow label="State">
                <Badge variant={getStateBadgeVariant(runner.state)}>{getStateLabel(runner.state)}</Badge>
              </InfoRow>
              <InfoRow label="Schedulable">
                <Switch
                  checked={!runner.unschedulable}
                  onCheckedChange={() => writePermitted && !isLoading && onToggleEnabled(runner)}
                  disabled={!writePermitted || isLoading}
                  aria-label="Toggle runner schedulable status"
                />
              </InfoRow>

              {deletePermitted && (
                <div className="flex justify-end pt-3">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onDelete(runner)}
                    disabled={isLoading}
                    className="text-destructive-foreground hover:bg-destructive/10 hover:text-destructive-foreground"
                  >
                    {isLoading ? <Spinner /> : <Trash2 className="size-4" />}
                    Delete
                  </Button>
                </div>
              )}
            </InfoSection>

            <InfoSection title="Placement">
              <InfoRow label="Region" className="-mr-2">
                <div className="flex items-center gap-1 min-w-0">
                  <span className="truncate">{getRegionName(runner.region) ?? runner.region}</span>
                  <CopyButton value={runner.region} tooltipText="Copy region" size="icon-xs" />
                </div>
              </InfoRow>
              <InfoRow label="Version">
                {runner.appVersion ?? <span className="text-muted-foreground">N/A</span>}
              </InfoRow>
            </InfoSection>

            <InfoSection title="Health">
              <div className="space-y-4 py-1">
                <SchedulingScore value={runner.availabilityScore} />
                <div className="flex flex-col gap-1.5">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">CPU Usage</span>
                    <span className="text-xs">
                      {runner.currentCpuUsagePercentage != null ? runner.currentCpuUsagePercentage.toFixed(2) : 'N/A'}%
                    </span>
                  </div>
                  <QuotaLine current={runner.currentCpuUsagePercentage ?? 0} total={100} />
                </div>
                <div className="flex flex-col gap-1.5">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">Memory Usage</span>
                    <span className="text-xs">
                      {runner.currentMemoryUsagePercentage != null
                        ? runner.currentMemoryUsagePercentage.toFixed(2)
                        : 'N/A'}
                      %
                    </span>
                  </div>
                  <QuotaLine current={runner.currentMemoryUsagePercentage ?? 0} total={100} />
                </div>
                <div className="flex flex-col gap-1.5">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">Disk Usage</span>
                    <span className="text-xs">
                      {runner.currentDiskUsagePercentage != null ? runner.currentDiskUsagePercentage.toFixed(2) : 'N/A'}
                      %
                    </span>
                  </div>
                  <QuotaLine current={runner.currentDiskUsagePercentage ?? 0} total={100} />
                </div>
              </div>
            </InfoSection>

            <InfoSection title="Resources">
              <div className="flex flex-wrap gap-2 py-1">
                <ResourceChip resource="cpu" value={Number(runner.cpu.toFixed(2))} />
                <ResourceChip resource="memory" value={Number(runner.memory.toFixed(2))} />
                <ResourceChip resource="disk" value={Number(runner.disk.toFixed(2))} />
                {runner.gpu !== undefined && runner.gpu > 0 && (
                  <ResourceChip
                    resource="gpu"
                    value={runner.gpu}
                    unit={runner.gpuType ? `GPU · ${runner.gpuType}` : undefined}
                  />
                )}
              </div>
              <div className="mt-3">
                <InfoRow label="Allocated CPU">{runner.currentAllocatedCpu ?? 0}</InfoRow>
                <InfoRow label="Allocated memory">{runner.currentAllocatedMemoryGiB ?? 0} GiB</InfoRow>
                <InfoRow label="Allocated disk">{runner.currentAllocatedDiskGiB ?? 0} GiB</InfoRow>
                <InfoRow label="Active sandboxes">{runner.currentStartedSandboxes ?? 0}</InfoRow>
                <InfoRow label="Snapshots">{runner.currentSnapshotCount ?? 0}</InfoRow>
              </div>
            </InfoSection>

            <InfoSection title="Activity">
              {runner.lastChecked && (
                <InfoRow label="Last checked">
                  <RunnerTimestamp timestamp={runner.lastChecked} />
                </InfoRow>
              )}
              <InfoRow label="Created">
                <RunnerTimestamp timestamp={runner.createdAt} />
              </InfoRow>
              <InfoRow label="Last updated">
                <RunnerTimestamp timestamp={runner.updatedAt} />
              </InfoRow>
            </InfoSection>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

const getStateBadgeVariant = (state: RunnerState): BadgeProps['variant'] => {
  switch (state) {
    case RunnerState.READY:
      return 'success'
    case RunnerState.UNRESPONSIVE:
      return 'destructive'
    case RunnerState.INITIALIZING:
      return 'warning'
    case RunnerState.DISABLED:
    case RunnerState.DECOMMISSIONED:
    default:
      return 'secondary'
  }
}

const getStateLabel = (state: RunnerState) => {
  return String(state)
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function SchedulingScore({ value }: { value?: number }) {
  return (
    <div>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="text-sm font-medium tracking-tight text-foreground">Scheduling Score</span>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className="inline-flex size-5 items-center justify-center rounded-full text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
                  aria-label="Scheduling score information"
                >
                  <CircleHelp className="size-3.5" />
                </button>
              </TooltipTrigger>
              <TooltipContent className="max-w-[260px]">
                Higher is better. Daytona uses this score to choose runners for new sandboxes based on health and
                remaining capacity.
              </TooltipContent>
            </Tooltip>
          </div>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">Higher means more ready for new sandboxes.</p>
        </div>
        <span className="shrink-0 text-2xl font-semibold leading-none tracking-tight tabular-nums">
          {value != null ? `${value.toFixed(0)}%` : 'N/A'}
        </span>
      </div>
    </div>
  )
}

function RunnerTimestamp({ timestamp }: { timestamp: string }) {
  return (
    <TimestampTooltip timestamp={timestamp}>
      <span>{getRelativeTimeString(timestamp).relativeTimeString}</span>
    </TimestampTooltip>
  )
}

export default RunnerDetailsSheet
