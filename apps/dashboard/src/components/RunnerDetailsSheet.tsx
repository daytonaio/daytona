/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import QuotaLine from '@/components/QuotaLine'
import { ResourceChip } from '@/components/ResourceChip'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { Runner, RunnerState } from '@daytona/api-client'
import { Trash, X } from 'lucide-react'
import React from 'react'
import { InfoRow, InfoSection } from './sandboxes/SandboxInfoPanel'
import { Badge, type BadgeProps } from './ui/badge'
import { ScrollArea } from './ui/scroll-area'

interface RunnerDetailsSheetProps {
  runner: Runner | null
  open: boolean
  onOpenChange: (open: boolean) => unknown
  runnerIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  onDelete: (runner: Runner) => unknown
  onToggleEnabled?: (runner: Runner) => unknown
  getRegionName: (regionId: string) => string | undefined
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
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function isNestedDialogEvent(event: Event) {
  return (
    event.target instanceof HTMLElement &&
    Boolean(event.target.closest('[data-slot="dialog-content"], [data-slot="dialog-overlay"]'))
  )
}

function RelativeTimestamp({ timestamp }: { timestamp?: string }) {
  if (!timestamp) {
    return <span className="text-muted-foreground font-normal">—</span>
  }

  return (
    <TimestampTooltip timestamp={timestamp}>
      <span>{getRelativeTimeString(timestamp).relativeTimeString}</span>
    </TimestampTooltip>
  )
}

function RunnerMetric({
  label,
  value,
  className,
}: {
  label: string
  value: number | null | undefined
  className?: string
}) {
  const percentage = value ?? 0

  return (
    <div className={cn('flex flex-col gap-1.5 py-1', className)}>
      <div className="flex items-center justify-between gap-3">
        <span className="text-sm text-muted-foreground">{label}</span>
        <span className="text-sm">{value != null ? `${value.toFixed(2)}%` : 'N/A'}</span>
      </div>
      <QuotaLine current={percentage} total={100} />
    </div>
  )
}

function RunnerOverviewPanel({
  runner,
  actionDisabled,
  deletePermitted,
  writePermitted,
  onDelete,
  onToggleEnabled,
  getRegionName,
}: {
  runner: Runner
  actionDisabled: boolean
  deletePermitted: boolean
  writePermitted: boolean
  onDelete: (runner: Runner) => unknown
  onToggleEnabled?: (runner: Runner) => unknown
  getRegionName: (regionId: string) => string | undefined
}) {
  const regionName = getRegionName(runner.region) ?? runner.region

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0">
        <InfoSection title={null} className="last:border-b">
          <InfoRow label="Name" className="-mr-2">
            <div className="flex items-center gap-1 min-w-0">
              <span className="truncate">{runner.name}</span>
              <CopyButton value={runner.name} tooltipText="Copy name" size="icon-xs" />
            </div>
          </InfoRow>
          <InfoRow label="UUID" className="-mr-2">
            <div className="flex items-center gap-1 min-w-0">
              <span className="truncate">{runner.id}</span>
              <CopyButton value={runner.id} tooltipText="Copy UUID" size="icon-xs" />
            </div>
          </InfoRow>
          <div className="flex items-center justify-between gap-3 pt-3">
            <Badge variant={getStateBadgeVariant(runner.state)}>{getStateLabel(runner.state)}</Badge>
            {deletePermitted && (
              <Button variant="ghostDestructive" size="sm" onClick={() => onDelete(runner)} disabled={actionDisabled}>
                <Trash className="size-4" />
                Delete
              </Button>
            )}
          </div>
        </InfoSection>
      </div>

      <ScrollArea fade="mask" className="min-h-0 flex-1">
        <div className="flex flex-col">
          <InfoSection title={null}>
            <InfoRow label="Region" className="-mr-2">
              <div className="flex items-center gap-1 min-w-0">
                <span className="truncate">{regionName}</span>
                <CopyButton value={regionName} tooltipText="Copy region" size="icon-xs" />
              </div>
            </InfoRow>
            <InfoRow label="Version">
              {runner.appVersion ? runner.appVersion : <span className="text-muted-foreground font-normal">—</span>}
            </InfoRow>
            <InfoRow label="Schedulable">
              {onToggleEnabled ? (
                <Switch
                  checked={!runner.unschedulable}
                  onCheckedChange={() => writePermitted && !actionDisabled && onToggleEnabled(runner)}
                  disabled={!writePermitted || actionDisabled}
                />
              ) : runner.unschedulable ? (
                'No'
              ) : (
                'Yes'
              )}
            </InfoRow>
          </InfoSection>

          <InfoSection title="Health Metrics">
            <div className="space-y-3">
              <RunnerMetric label="Availability score" value={runner.availabilityScore} />
              <RunnerMetric label="CPU usage" value={runner.currentCpuUsagePercentage} />
              <RunnerMetric label="Memory usage" value={runner.currentMemoryUsagePercentage} />
              <RunnerMetric label="Disk usage" value={runner.currentDiskUsagePercentage} />
            </div>
          </InfoSection>

          <InfoSection title="Current Load">
            <InfoRow label="Active sandboxes">{runner.currentStartedSandboxes ?? 0}</InfoRow>
            <InfoRow label="Snapshots">{runner.currentSnapshotCount ?? 0}</InfoRow>
          </InfoSection>

          <InfoSection title="Total Resources">
            <div className="flex flex-wrap gap-2 py-1">
              <ResourceChip resource="cpu" value={Number(runner.cpu.toFixed(2))} />
              <ResourceChip resource="memory" value={Number(runner.memory.toFixed(2))} />
              <ResourceChip resource="disk" value={Number(runner.disk.toFixed(2))} />
              {runner.gpu !== undefined && runner.gpu > 0 && (
                <div className="flex items-center gap-1 rounded-full border border-border bg-muted/80 px-2 py-[2px] text-sm whitespace-nowrap">
                  {runner.gpu} GPU{runner.gpuType ? ` (${runner.gpuType})` : ''}
                </div>
              )}
            </div>
          </InfoSection>

          <InfoSection title="Activity">
            <InfoRow label="Last checked">
              <RelativeTimestamp timestamp={runner.lastChecked} />
            </InfoRow>
            <InfoRow label="Created">
              <RelativeTimestamp timestamp={runner.createdAt} />
            </InfoRow>
            <InfoRow label="Last updated">
              <RelativeTimestamp timestamp={runner.updatedAt} />
            </InfoRow>
          </InfoSection>
        </div>
      </ScrollArea>
    </div>
  )
}

const RunnerDetailsSheet: React.FC<RunnerDetailsSheetProps> = ({
  runner,
  open,
  onOpenChange,
  runnerIsLoading,
  writePermitted,
  deletePermitted,
  onDelete,
  onToggleEnabled,
  getRegionName,
}) => {
  if (!runner) return null

  const actionDisabled = runnerIsLoading[runner.id] || false

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[800px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="space-y-0 flex flex-row justify-between items-center p-6">
          <SheetTitle>Runner Details</SheetTitle>
          <div className="flex items-center">
            {deletePermitted && (
              <Button
                variant="outline"
                className="w-8 h-8"
                onClick={() => onDelete(runner)}
                disabled={runnerIsLoading[runner.id]}
              >
                <Trash className="w-4 h-4" />
              </Button>
            )}
            <Button
              variant="outline"
              className="w-8 h-8"
              onClick={() => onOpenChange(false)}
              disabled={runnerIsLoading[runner.id]}
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
          <Button variant="ghost" size="icon-sm" onClick={() => onOpenChange(false)} disabled={actionDisabled}>
            <X className="size-4" />
            <span className="sr-only">Close</span>
          </Button>
        </SheetHeader>

        <RunnerOverviewPanel
          runner={runner}
          actionDisabled={actionDisabled}
          deletePermitted={deletePermitted}
          writePermitted={writePermitted}
          onDelete={onDelete}
          onToggleEnabled={onToggleEnabled}
          getRegionName={getRegionName}
        />
      </SheetContent>
    </Sheet>
  )
}

export default RunnerDetailsSheet
