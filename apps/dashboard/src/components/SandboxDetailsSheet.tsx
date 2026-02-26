/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { formatDuration, formatTimestamp, getRelativeTimeString } from '@/lib/utils'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { Archive, Play, Tag, Trash, Wrench, X } from 'lucide-react'
import React, { useState } from 'react'
import { Link, generatePath } from 'react-router-dom'
import { RoutePath } from '@/enums/RoutePath'
import { CopyButton } from './CopyButton'
import { ResourceChip } from './ResourceChip'
import { SandboxState as SandboxStateComponent } from './SandboxTable/SandboxState'
import { TimestampTooltip } from './TimestampTooltip'
import { LogsTab, TracesTab, MetricsTab } from './telemetry'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { FeatureFlags } from '@/enums/FeatureFlags'

interface SandboxDetailsSheetProps {
  sandbox: Sandbox | null
  open: boolean
  onOpenChange: (open: boolean) => void
  sandboxIsLoading: Record<string, boolean>
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  getRegionName: (regionId: string) => string | undefined
  writePermitted: boolean
  deletePermitted: boolean
  handleRecover: (id: string) => void
}

const SandboxDetailsSheet: React.FC<SandboxDetailsSheetProps> = ({
  sandbox,
  open,
  onOpenChange,
  sandboxIsLoading,
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  getWebTerminalUrl,
  getRegionName,
  writePermitted,
  deletePermitted,
  handleRecover,
}) => {
  const [terminalUrl, setTerminalUrl] = useState<string | null>(null)
  const experimentsEnabled = useFeatureFlagEnabled(FeatureFlags.ORGANIZATION_EXPERIMENTS)

  // TODO: uncomment when we enable the terminal tab
  // useEffect(() => {
  //   const getTerminalUrl = async () => {
  //     if (!sandbox?.id) {
  //       setTerminalUrl(null)
  //       return
  //     }

  //     const url = await getWebTerminalUrl(sandbox.id)
  //     setTerminalUrl(url)
  //   }

  //   getTerminalUrl()
  // }, [sandbox?.id, getWebTerminalUrl])

  if (!sandbox) return null

  const getLastEvent = (sandbox: Sandbox): { date: Date; relativeTimeString: string } => {
    return getRelativeTimeString(sandbox.updatedAt)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[800px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="space-y-0 flex flex-row justify-between items-center  p-4 px-5 border-b border-border">
          <SheetTitle className="text-2xl font-medium">Sandbox Details</SheetTitle>
          <div className="flex gap-2 items-center">
            <Button variant="link" asChild>
              <Link to={generatePath(RoutePath.SANDBOX_DETAILS, { sandboxId: sandbox.id })}>View</Link>
            </Button>
            {writePermitted && (
              <>
                {sandbox.state === SandboxState.STARTED && (
                  <Button
                    variant="outline"
                    onClick={() => handleStop(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    Stop
                  </Button>
                )}
                {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) &&
                  !sandbox.recoverable && (
                    <Button
                      variant="outline"
                      onClick={() => handleStart(sandbox.id)}
                      disabled={sandboxIsLoading[sandbox.id]}
                    >
                      <Play className="w-4 h-4" />
                      Start
                    </Button>
                  )}
                {sandbox.state === SandboxState.ERROR && sandbox.recoverable && (
                  <Button
                    variant="outline"
                    onClick={() => handleRecover(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Wrench className="w-4 h-4" />
                    Recover
                  </Button>
                )}
                {/* {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                  <Button
                    variant="outline"
                    onClick={() => handleFork(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <GitFork className="w-4 h-4" />
                    Fork
                  </Button>
                )}
                {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                  <Button
                    variant="outline"
                    onClick={() => handleSnapshot(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Camera className="w-4 h-4" />
                    Snapshot
                  </Button>
                )} */}
                {sandbox.state === SandboxState.STOPPED && (
                  <Button
                    variant="outline"
                    className="w-8 h-8"
                    onClick={() => handleArchive(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Archive className="w-4 h-4" />
                  </Button>
                )}
              </>
            )}
            {deletePermitted && (
              <Button
                variant="outline"
                className="w-8 h-8"
                onClick={() => handleDelete(sandbox.id)}
                disabled={sandboxIsLoading[sandbox.id]}
              >
                <Trash className="w-4 h-4" />
              </Button>
            )}
            <Button
              variant="outline"
              className="w-8 h-8"
              onClick={() => onOpenChange(false)}
              disabled={sandboxIsLoading[sandbox.id]}
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
        </SheetHeader>

        <Tabs defaultValue="overview" className="flex-1 flex flex-col min-h-0">
          {experimentsEnabled && (
            <TabsList className="mx-4 w-fit flex-shrink-0 bg-transparent border-b border-border rounded-none h-auto p-0 gap-0 mt-2">
              <TabsTrigger
                value="overview"
                className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2"
              >
                Overview
              </TabsTrigger>
              <TabsTrigger
                value="logs"
                className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2"
              >
                Logs
              </TabsTrigger>
              <TabsTrigger
                value="traces"
                className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2"
              >
                Traces
              </TabsTrigger>
              <TabsTrigger
                value="metrics"
                className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2"
              >
                Metrics
              </TabsTrigger>
            </TabsList>
          )}

          <TabsContent value="overview" className="flex-1 p-6 space-y-10 overflow-y-auto min-h-0">
            <div className="grid grid-cols-2 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Name</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.name}</p>
                  <CopyButton value={sandbox.name} tooltipText="Copy name" size="icon-xs" />
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">UUID</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.id}</p>
                  <CopyButton value={sandbox.id} tooltipText="Copy UUID" size="icon-xs" />
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">State</h3>
                <div className="mt-1 text-sm">
                  <SandboxStateComponent
                    state={sandbox.state}
                    errorReason={sandbox.errorReason}
                    recoverable={sandbox.recoverable}
                  />
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Snapshot</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.snapshot || '-'}</p>
                  {sandbox.snapshot && (
                    <CopyButton value={sandbox.snapshot} tooltipText="Copy snapshot" size="icon-xs" />
                  )}
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Region</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{getRegionName(sandbox.target) ?? sandbox.target}</p>
                  <CopyButton value={sandbox.target} tooltipText="Copy region" size="icon-xs" />
                </div>
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Last event</h3>
                <p className="mt-1 text-sm font-medium">
                  <TimestampTooltip timestamp={sandbox.updatedAt}>
                    {getLastEvent(sandbox).relativeTimeString}
                  </TimestampTooltip>
                </p>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Created at</h3>
                <p className="mt-1 text-sm font-medium">
                  <TimestampTooltip timestamp={sandbox.createdAt}>
                    {formatTimestamp(sandbox.createdAt)}
                  </TimestampTooltip>
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Auto-stop</h3>
                <p className="mt-1 text-sm font-medium">
                  {sandbox.autoStopInterval ? formatDuration(sandbox.autoStopInterval) : 'Disabled'}
                </p>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Auto-archive</h3>
                <p className="mt-1 text-sm font-medium">
                  {sandbox.autoArchiveInterval ? formatDuration(sandbox.autoArchiveInterval) : 'Disabled'}
                </p>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Auto-delete</h3>
                <p className="mt-1 text-sm font-medium">
                  {sandbox.autoDeleteInterval !== undefined && sandbox.autoDeleteInterval >= 0
                    ? sandbox.autoDeleteInterval === 0
                      ? 'On stop'
                      : formatDuration(sandbox.autoDeleteInterval)
                    : 'Disabled'}
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1">
              <div>
                <h3 className="text-sm text-muted-foreground">Resources</h3>
                <div className="mt-1 text-sm font-medium flex items-center gap-1 flex-wrap">
                  <ResourceChip resource="cpu" value={sandbox.cpu} />
                  <ResourceChip resource="memory" value={sandbox.memory} />
                  <ResourceChip resource="disk" value={sandbox.disk} />
                </div>
              </div>
            </div>
            <div>
              <h3 className="text-lg font-medium">Labels</h3>
              <div className="mt-3 space-y-4">
                {Object.entries(sandbox.labels ?? {}).length > 0 ? (
                  Object.entries(sandbox.labels ?? {}).map(([key, value]) => (
                    <div key={key} className="text-sm">
                      <div>{key}</div>
                      <div className="font-medium p-2 bg-muted rounded-md mt-1 border border-border">{value}</div>
                    </div>
                  ))
                ) : (
                  <div className="flex flex-col border border-border rounded-md items-center justify-center gap-2 text-muted-foreground w-full min-h-40">
                    <Tag className="w-4 h-4" />
                    <span className="text-sm">No labels found</span>
                  </div>
                )}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="terminal" className="p-4">
            <iframe title="Terminal" src={terminalUrl || undefined} className="w-full h-full"></iframe>
          </TabsContent>

          <TabsContent value="logs" className="flex-1 min-h-0 overflow-hidden">
            <LogsTab sandboxId={sandbox.id} />
          </TabsContent>

          <TabsContent value="traces" className="flex-1 min-h-0 overflow-hidden">
            <TracesTab sandboxId={sandbox.id} />
          </TabsContent>

          <TabsContent value="metrics" className="flex-1 min-h-0 overflow-hidden">
            <MetricsTab sandboxId={sandbox.id} />
          </TabsContent>
        </Tabs>
      </SheetContent>
    </Sheet>
  )
}

export default SandboxDetailsSheet
