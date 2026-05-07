/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TabsContent } from '@/components/ui/tabs'
import { Sandbox } from '@daytona/api-client'
import { SandboxFileSystemTab } from '../SandboxFileSystemTab'
import { SandboxLogsTab } from '../SandboxLogsTab'
import { SandboxMetricsTab } from '../SandboxMetricsTab'
import { SandboxSpendingTab } from '../SandboxSpendingTab'
import { SandboxTerminalTab } from '../SandboxTerminalTab'
import { SandboxTracesTab } from '../SandboxTracesTab'
import { SandboxVncTab } from '../SandboxVncTab'

export function SandboxDetailsTabContent({
  sandbox,
  filesystemEnabled,
  spendingTabAvailable,
}: {
  sandbox: Sandbox
  filesystemEnabled: boolean | undefined
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
      {filesystemEnabled && (
        <TabsContent
          value="filesystem"
          className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden"
        >
          <SandboxFileSystemTab sandbox={sandbox} />
        </TabsContent>
      )}
      <TabsContent value="vnc" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden">
        <SandboxVncTab sandbox={sandbox} />
      </TabsContent>
    </>
  )
}
