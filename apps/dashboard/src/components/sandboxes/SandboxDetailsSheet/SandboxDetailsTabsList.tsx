/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TabsTrigger } from '@/components/ui/tabs'
import React from 'react'
import { FadeTabList } from './FadeTabList'

export function SandboxDetailsTabsList({
  filesystemEnabled,
  spendingTabAvailable,
  showOverview,
  leadingContent,
}: {
  filesystemEnabled: boolean | undefined
  spendingTabAvailable: boolean | undefined
  showOverview: boolean
  leadingContent?: React.ReactNode
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
      {filesystemEnabled && (
        <TabsTrigger value="filesystem" className="h-[41px] border-b py-0">
          Filesystem
        </TabsTrigger>
      )}
      <TabsTrigger value="vnc" className="h-[41px] border-b py-0">
        VNC
      </TabsTrigger>
    </FadeTabList>
  )
}

export function SandboxOverviewTabTrigger() {
  return (
    <div className="h-[42px] shrink-0 border-b border-border">
      <div className="inline-flex h-[41px] items-center px-4 py-0 text-sm font-medium text-foreground">Overview</div>
    </div>
  )
}
