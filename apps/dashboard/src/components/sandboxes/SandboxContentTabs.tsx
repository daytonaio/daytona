/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useRegions } from '@/hooks/useRegions'
import { Sandbox } from '@daytona/api-client'
import { SandboxFileSystemTab } from './SandboxFileSystemTab'
import { SandboxInfoPanel } from './SandboxInfoPanel'
import { SandboxLogsTab } from './SandboxLogsTab'
import { SandboxMetricsTab } from './SandboxMetricsTab'
import { SandboxSpendingTab } from './SandboxSpendingTab'
import { SandboxTerminalTab } from './SandboxTerminalTab'
import { SandboxTracesTab } from './SandboxTracesTab'
import { SandboxVncTab } from './SandboxVncTab'
import { TabValue } from './SearchParams'

interface SandboxContentTabsProps {
  sandbox: Sandbox | undefined
  isLoading: boolean
  spendingTabAvailable: boolean
  filesystemEnabled: boolean | undefined
  tab: TabValue
  onTabChange: (tab: TabValue) => void
}

export function SandboxContentTabs({
  sandbox,
  isLoading,
  spendingTabAvailable,
  filesystemEnabled,
  tab,
  onTabChange,
}: SandboxContentTabsProps) {
  const { getRegionName } = useRegions()

  if (isLoading) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-0 border-b border-border h-[41px] px-4 shrink-0">
          <Skeleton className="h-4 w-16 lg:hidden" />
          <Skeleton className="h-4 w-10 ml-4 lg:ml-0" />
          <Skeleton className="h-4 w-12 ml-4" />
          <Skeleton className="h-4 w-14 ml-4" />
          <Skeleton className="h-4 w-16 ml-4" />
          <Skeleton className="h-4 w-10 ml-4" />
        </div>
        <div className="flex-1 flex items-center justify-center text-muted-foreground">
          <Spinner className="size-5" />
        </div>
      </div>
    )
  }

  if (!sandbox) return null

  return (
    <Tabs value={tab} onValueChange={(v) => onTabChange(v as TabValue)} className="flex flex-col h-full gap-0">
      <ScrollArea
        fade="mask"
        horizontal
        vertical={false}
        fadeOffset={36}
        className="h-[41px] shrink-0 border-b border-border"
      >
        <TabsList variant="underline" className="h-[41px] w-max min-w-full border-b-0">
          <TabsTrigger value="overview" className="lg:hidden">
            Overview
          </TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="traces">Traces</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          {spendingTabAvailable && <TabsTrigger value="spending">Spending</TabsTrigger>}
          <TabsTrigger value="terminal">Terminal</TabsTrigger>
          {filesystemEnabled && <TabsTrigger value="filesystem">Filesystem</TabsTrigger>}
          <TabsTrigger value="vnc">VNC</TabsTrigger>
        </TabsList>
      </ScrollArea>

      <TabsContent value="overview" className="flex-1 min-h-0 m-0 overflow-y-auto scrollbar-sm lg:hidden">
        <SandboxInfoPanel sandbox={sandbox} getRegionName={getRegionName} />
      </TabsContent>
      <TabsContent value="logs" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxLogsTab sandboxId={sandbox.id} />
      </TabsContent>
      <TabsContent value="traces" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxTracesTab sandboxId={sandbox.id} />
      </TabsContent>
      <TabsContent value="metrics" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxMetricsTab sandboxId={sandbox.id} />
      </TabsContent>
      {spendingTabAvailable && (
        <TabsContent value="spending" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
          <SandboxSpendingTab sandboxId={sandbox.id} />
        </TabsContent>
      )}
      <TabsContent value="terminal" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxTerminalTab sandbox={sandbox} />
      </TabsContent>
      {filesystemEnabled && (
        <TabsContent
          value="filesystem"
          className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
        >
          <SandboxFileSystemTab sandbox={sandbox} />
        </TabsContent>
      )}
      <TabsContent value="vnc" className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxVncTab sandbox={sandbox} />
      </TabsContent>
    </Tabs>
  )
}
