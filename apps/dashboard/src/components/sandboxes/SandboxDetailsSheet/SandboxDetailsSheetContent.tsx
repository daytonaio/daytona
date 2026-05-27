/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { Sandbox } from '@daytona/api-client'
import { ChevronsRight } from 'lucide-react'
import type { ReactNode } from 'react'
import { SandboxActionsSegmented } from '../SandboxActionsSegmented'
import { SandboxFileSystemTab } from '../SandboxFileSystemTab'
import { SandboxInfoPanel } from '../SandboxInfoPanel'
import { SandboxLogsTab } from '../SandboxLogsTab'
import { SandboxMetricsTab } from '../SandboxMetricsTab'
import { SandboxSpendingTab } from '../SandboxSpendingTab'
import { SandboxTerminalTab } from '../SandboxTerminalTab'
import { SandboxTracesTab } from '../SandboxTracesTab'
import { SandboxVncTab } from '../SandboxVncTab'
import { FadeTabList } from './FadeTabList'
import { SandboxDetailsHeader } from './SandboxDetailsHeader'
import type { SandboxDetailsSheetTabValue } from './SandboxDetailsSheet'

interface SandboxDetailsSheetContentProps {
  sandbox: Sandbox
  activeTab: SandboxDetailsSheetTabValue
  onTabChange: (value: string) => void
  isDesktop: boolean
  spendingTabAvailable: boolean
  actionDisabled: boolean
  writePermitted: boolean
  deletePermitted: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
  onScreenRecordings: (id: string) => void
  onResetToOverview: () => void
}

function SandboxDetailsTabsList({
  spendingTabAvailable,
  showOverview,
  leadingContent,
}: {
  spendingTabAvailable: boolean | undefined
  showOverview: boolean
  leadingContent?: ReactNode
}) {
  return (
    <FadeTabList leadingContent={leadingContent}>
      {showOverview && (
        <TabsTrigger value="overview" className="h-[41px] border-b py-0">
          Overview
        </TabsTrigger>
      )}
      <TabsTrigger value="logs" className="h-[41px] border-b py-0">
        Logs
      </TabsTrigger>
      <TabsTrigger value="traces" className="h-[41px] border-b py-0">
        Traces
      </TabsTrigger>
      <TabsTrigger value="metrics" className="h-[41px] border-b py-0">
        Metrics
      </TabsTrigger>
      {spendingTabAvailable && (
        <TabsTrigger value="spending" className="h-[41px] border-b py-0">
          Spending
        </TabsTrigger>
      )}
      <TabsTrigger value="terminal" className="h-[41px] border-b py-0">
        Terminal
      </TabsTrigger>
      <TabsTrigger value="filesystem" className="h-[41px] border-b py-0">
        Filesystem
      </TabsTrigger>
      <TabsTrigger value="vnc" className="h-[41px] border-b py-0">
        VNC
      </TabsTrigger>
    </FadeTabList>
  )
}

function SandboxDetailsTabContent({
  sandbox,
  spendingTabAvailable,
}: {
  sandbox: Sandbox
  spendingTabAvailable: boolean | undefined
}) {
  return (
    <>
      <TabsContent value="logs" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxLogsTab sandboxId={sandbox.id} />
      </TabsContent>
      <TabsContent value="traces" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxTracesTab sandboxId={sandbox.id} />
      </TabsContent>
      <TabsContent value="metrics" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxMetricsTab sandboxId={sandbox.id} />
      </TabsContent>
      {spendingTabAvailable && (
        <TabsContent value="spending" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
          <SandboxSpendingTab sandboxId={sandbox.id} />
        </TabsContent>
      )}
      <TabsContent value="terminal" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxTerminalTab sandbox={sandbox} />
      </TabsContent>
      <TabsContent value="filesystem" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxFileSystemTab sandbox={sandbox} />
      </TabsContent>
      <TabsContent value="vnc" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxVncTab sandbox={sandbox} />
      </TabsContent>
    </>
  )
}

function SandboxOverviewTabTrigger() {
  return (
    <div className="h-[42px] shrink-0 border-b border-border">
      <div className="inline-flex h-[41px] items-center px-4 py-0 text-sm font-medium text-foreground">Overview</div>
    </div>
  )
}

export function SandboxDetailsSheetContent({
  sandbox,
  activeTab,
  onTabChange,
  isDesktop,
  spendingTabAvailable,
  actionDisabled,
  writePermitted,
  deletePermitted,
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleRecover,
  getRegionName,
  onCreateSshAccess,
  onRevokeSshAccess,
  onScreenRecordings,
  onResetToOverview,
}: SandboxDetailsSheetContentProps) {
  const sandboxHeaderActions =
    writePermitted || deletePermitted ? (
      <SandboxActionsSegmented
        sandbox={sandbox}
        writePermitted={writePermitted}
        deletePermitted={deletePermitted}
        actionsDisabled={actionDisabled}
        onStart={() => handleStart(sandbox.id)}
        onStop={() => handleStop(sandbox.id)}
        onArchive={() => handleArchive(sandbox.id)}
        onRecover={() => handleRecover(sandbox.id)}
        onDelete={() => handleDelete(sandbox.id)}
        onCreateSshAccess={() => onCreateSshAccess(sandbox.id)}
        onRevokeSshAccess={() => onRevokeSshAccess(sandbox.id)}
        onScreenRecordings={() => onScreenRecordings(sandbox.id)}
      />
    ) : null

  const sandboxInfoPanel = (
    <ScrollArea fade="mask" className="flex-1 min-h-0">
      <SandboxInfoPanel
        sandbox={sandbox}
        getRegionName={getRegionName}
        actionsDisabled={actionDisabled}
        writePermitted={writePermitted}
        onCreateSshAccess={() => onCreateSshAccess(sandbox.id)}
        onRevokeSshAccess={() => onRevokeSshAccess(sandbox.id)}
        onScreenRecordings={() => onScreenRecordings(sandbox.id)}
      />
    </ScrollArea>
  )

  return (
    <Tabs value={activeTab} onValueChange={onTabChange} className="flex flex-1 min-h-0 flex-col gap-0">
      {isDesktop ? (
        activeTab === 'overview' ? (
          <div className="flex flex-1 min-h-0 flex-col overflow-hidden">
            <SandboxDetailsTabsList spendingTabAvailable={spendingTabAvailable} showOverview />
            <div className="flex min-h-0 flex-1 max-w-[450px] flex-col">
              <SandboxDetailsHeader sandbox={sandbox} actions={sandboxHeaderActions} />
              {sandboxInfoPanel}
            </div>
          </div>
        ) : (
          <div className="flex flex-1 min-h-0 overflow-hidden">
            <div className="flex w-[450px] shrink-0 min-h-0 flex-col border-r border-border">
              <SandboxOverviewTabTrigger />
              <SandboxDetailsHeader sandbox={sandbox} actions={sandboxHeaderActions} />
              {sandboxInfoPanel}
            </div>
            <div className="flex min-w-0 min-h-0 flex-1 flex-col overflow-hidden">
              <div className="animate-in fade-in slide-in-from-left-3 duration-500 ease-out">
                <SandboxDetailsTabsList
                  spendingTabAvailable={spendingTabAvailable}
                  showOverview={false}
                  leadingContent={
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          className="mx-1 size-8 text-muted-foreground hover:text-foreground"
                          onClick={() => onResetToOverview()}
                        >
                          <ChevronsRight className="size-4" />
                          <span className="sr-only">Collapse</span>
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>Collapse</TooltipContent>
                    </Tooltip>
                  }
                />
              </div>
              <div className="flex min-w-0 min-h-0 flex-1 flex-col overflow-hidden">
                <div className="flex min-h-0 flex-1 flex-col animate-in fade-in slide-in-from-right-20 duration-300 ease-out">
                  <SandboxDetailsTabContent sandbox={sandbox} spendingTabAvailable={spendingTabAvailable} />
                </div>
              </div>
            </div>
          </div>
        )
      ) : (
        <>
          <div className="flex min-h-0 shrink-0 flex-col">
            <SandboxDetailsHeader sandbox={sandbox} actions={sandboxHeaderActions} />
            <SandboxDetailsTabsList spendingTabAvailable={spendingTabAvailable} showOverview />
          </div>

          <TabsContent value="overview" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col">
            {sandboxInfoPanel}
          </TabsContent>

          <SandboxDetailsTabContent sandbox={sandbox} spendingTabAvailable={spendingTabAvailable} />
        </>
      )}
    </Tabs>
  )
}
