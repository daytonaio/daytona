/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { formatTimestamp, getRelativeTimeString } from '@/lib/utils'
import { Runner, RunnerState } from '@daytonaio/api-client'
import { Copy, Trash, X, CheckCircle, AlertTriangle, Timer, Pause } from 'lucide-react'
import React from 'react'
import { toast } from 'sonner'
import { ResourceChip } from './ResourceChip'
import QuotaLine from './QuotaLine'

interface RunnerDetailsSheetProps {
  runner: Runner | null
  open: boolean
  onOpenChange: (open: boolean) => void
  runnerIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  onDelete: (runner: Runner) => void
  getRegionName: (regionId: string) => string | undefined
}

const RunnerDetailsSheet: React.FC<RunnerDetailsSheetProps> = ({
  runner,
  open,
  onOpenChange,
  runnerIsLoading,
  deletePermitted,
  onDelete,
  getRegionName,
}) => {
  if (!runner) return null

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  const getStateIcon = (state: RunnerState) => {
    switch (state) {
      case RunnerState.READY:
        return <CheckCircle className="w-4 h-4 flex-shrink-0" />
      case RunnerState.DISABLED:
      case RunnerState.DECOMMISSIONED:
        return <Pause className="w-4 h-4 flex-shrink-0" />
      case RunnerState.UNRESPONSIVE:
        return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
      default:
        return <Timer className="w-4 h-4 flex-shrink-0" />
    }
  }

  const getStateColor = (state: RunnerState) => {
    switch (state) {
      case RunnerState.READY:
        return 'text-green-500'
      case RunnerState.DISABLED:
      case RunnerState.DECOMMISSIONED:
        return 'text-gray-500 dark:text-gray-400'
      case RunnerState.UNRESPONSIVE:
        return 'text-red-500'
      default:
        return 'text-gray-600 dark:text-gray-400'
    }
  }

  const getStateLabel = (state: RunnerState) => {
    return state
      .split('_')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ')
  }

  const getLastEvent = (runner: Runner): { date: Date; relativeTimeString: string } => {
    return getRelativeTimeString(runner.updatedAt)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[800px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="space-y-0 flex flex-row justify-between items-center p-6">
          <SheetTitle className="text-2xl font-medium">Runner Details</SheetTitle>
          <div className="flex gap-2 items-center">
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
        </SheetHeader>

        <Tabs defaultValue="overview" className="flex-1 flex flex-col min-h-0">
          <TabsContent value="overview" className="flex-1 p-6 space-y-10 overflow-y-auto min-h-0">
            {/* Basic Info */}
            <div className="grid grid-cols-2 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Name</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{runner.name}</p>
                  <button
                    onClick={() => copyToClipboard(runner.name)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy name"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">UUID</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{runner.id}</p>
                  <button
                    onClick={() => copyToClipboard(runner.id)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy UUID"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>

            {/* Status */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">State</h3>
                <div className={`mt-1 flex items-center gap-2 ${getStateColor(runner.state)}`}>
                  {getStateIcon(runner.state)}
                  <span className="text-sm font-medium">{getStateLabel(runner.state)}</span>
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Schedulable</h3>
                <p className="mt-1 text-sm font-medium">{runner.unschedulable ? 'No' : 'Yes'}</p>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Region</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{getRegionName(runner.region) ?? runner.region}</p>
                  <button
                    onClick={() => copyToClipboard(runner.region)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy region"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Version</h3>
                <p className="mt-1 text-sm font-medium">{runner.appVersion ?? 'N/A'}</p>
              </div>
            </div>

            {/* Health Metrics */}
            <div>
              <h3 className="text-lg font-medium mb-4">Health Metrics</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="flex flex-col gap-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">Availability Score</span>
                    <span className="text-xs">
                      {runner.availabilityScore != null ? runner.availabilityScore.toFixed(2) : 'N/A'}%
                    </span>
                  </div>
                  <QuotaLine current={runner.availabilityScore ?? 0} total={100} />
                </div>
                <div className="flex flex-col gap-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">CPU Usage</span>
                    <span className="text-xs">
                      {runner.currentCpuUsagePercentage != null ? runner.currentCpuUsagePercentage.toFixed(2) : 'N/A'}%
                    </span>
                  </div>
                  <QuotaLine current={runner.currentCpuUsagePercentage ?? 0} total={100} />
                </div>
                <div className="flex flex-col gap-1">
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
                <div className="flex flex-col gap-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground text-xs">Disk Usage</span>
                    <span className="text-xs">
                      {runner.currentDiskUsagePercentage != null ? runner.currentDiskUsagePercentage.toFixed(2) : 'N/A'}
                      %
                    </span>
                  </div>
                  <QuotaLine current={runner.currentDiskUsagePercentage ?? 0} total={100} />
                </div>
                <div>
                  <h4 className="text-muted-foreground text-xs">Active Sandboxes</h4>
                  <p className="mt-1 text-2xl font-semibold">{runner.currentStartedSandboxes ?? 0}</p>
                </div>
                <div>
                  <h4 className="text-muted-foreground text-xs">Snapshots</h4>
                  <p className="mt-1 text-2xl font-semibold">{runner.currentSnapshotCount ?? 0}</p>
                </div>
              </div>
            </div>

            {/* Total Resources */}
            <div className="grid grid-cols-1">
              <div>
                <h3 className="text-sm text-muted-foreground">Total Resources</h3>
                <div className="mt-1 text-sm font-medium flex items-center gap-1 flex-wrap">
                  <ResourceChip resource="cpu" value={Number(runner.cpu.toFixed(2))} />
                  <ResourceChip resource="memory" value={Number(runner.memory.toFixed(2))} />
                  <ResourceChip resource="disk" value={Number(runner.disk.toFixed(2))} />
                  {runner.gpu !== undefined && runner.gpu > 0 && (
                    <div className="flex items-center gap-1 bg-purple-100 text-purple-600 dark:bg-purple-950 dark:text-purple-200 rounded-full px-2 py-[2px] text-sm whitespace-nowrap">
                      {runner.gpu} GPU{runner.gpuType ? ` (${runner.gpuType})` : ''}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Timestamps */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {runner.lastChecked && (
                <div>
                  <h3 className="text-sm text-muted-foreground">Last Checked</h3>
                  <p className="mt-1 text-sm font-medium">{formatTimestamp(runner.lastChecked)}</p>
                </div>
              )}
              <div>
                <h3 className="text-sm text-muted-foreground">Created At</h3>
                <p className="mt-1 text-sm font-medium">{formatTimestamp(runner.createdAt)}</p>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Last Updated</h3>
                <p className="mt-1 text-sm font-medium">{getLastEvent(runner).relativeTimeString}</p>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </SheetContent>
    </Sheet>
  )
}

export default RunnerDetailsSheet
