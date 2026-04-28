/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import {
  ResizableSheetContent,
  Sheet,
  SheetHeader,
  SheetTitle,
  type ResizableSheetContentRef,
} from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { RoutePath } from '@/enums/RoutePath'
import { useConfig } from '@/hooks/useConfig'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import { Sandbox } from '@daytona/api-client'
import { ArrowRight, ChevronDown, ChevronUp, X } from 'lucide-react'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import React, { useCallback, useEffect, useState } from 'react'
import { Link, generatePath } from 'react-router-dom'
import { CopyButton } from './CopyButton'
import { SandboxState as SandboxStateComponent } from './SandboxTable/SandboxState'
import { SandboxActionsSegmented } from './sandboxes/SandboxActionsSegmented'
import { SandboxFileSystemTab } from './sandboxes/SandboxFileSystemTab'
import { InfoRow, InfoSection, SandboxInfoPanel } from './sandboxes/SandboxInfoPanel'
import { SandboxLogsTab } from './sandboxes/SandboxLogsTab'
import { SandboxMetricsTab } from './sandboxes/SandboxMetricsTab'
import { SandboxSpendingTab } from './sandboxes/SandboxSpendingTab'
import { SandboxTerminalTab } from './sandboxes/SandboxTerminalTab'
import { SandboxTracesTab } from './sandboxes/SandboxTracesTab'
import { SandboxVncTab } from './sandboxes/SandboxVncTab'
import { ScrollArea } from './ui/scroll-area'
import { useSidebar } from './ui/sidebar'

type SheetTabValue = 'overview' | 'logs' | 'traces' | 'metrics' | 'spending' | 'terminal' | 'filesystem' | 'vnc'

const OVERVIEW_WIDTH = 580
const EXPANDED_WIDTH = 1024
const MOBILE_BREAKPOINT = 640
const TAB_RESIZE_DURATION = 0.5

interface SandboxDetailsSheetProps {
  sandbox: Sandbox | null
  open: boolean
  onOpenChange: (open: boolean) => void
  sandboxIsLoading: Record<string, boolean>
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  writePermitted: boolean
  deletePermitted: boolean
  handleRecover: (id: string) => void
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
  onScreenRecordings: (id: string) => void
  onNavigate: (direction: 'prev' | 'next') => void
  hasPrev: boolean
  hasNext: boolean
}

function getViewportWidth() {
  return typeof window === 'undefined' ? EXPANDED_WIDTH : window.innerWidth
}

function getOverviewWidth(viewportWidth: number) {
  return Math.min(OVERVIEW_WIDTH, viewportWidth)
}

function getExpandedWidth(viewportWidth: number, sidebarWidth: number) {
  return Math.max(OVERVIEW_WIDTH, Math.min(EXPANDED_WIDTH, viewportWidth - sidebarWidth))
}

function getWidthForTab(tab: SheetTabValue, viewportWidth: number, sidebarWidth: number) {
  return tab === 'overview' ? getOverviewWidth(viewportWidth) : getExpandedWidth(viewportWidth, sidebarWidth)
}

function clampSheetWidth(width: number, viewportWidth: number, sidebarWidth: number) {
  return Math.min(Math.max(width, getOverviewWidth(viewportWidth)), getExpandedWidth(viewportWidth, sidebarWidth))
}

function SandboxOverviewTab({
  sandbox,
  getRegionName,
  actionDisabled,
  writePermitted,
  deletePermitted,
  handleStart,
  handleStop,
  handleRecover,
  handleArchive,
  handleDelete,
  onCreateSshAccess,
  onRevokeSshAccess,
  onScreenRecordings,
  detailsPath,
}: {
  sandbox: Sandbox
  getRegionName: (regionId: string) => string | undefined
  actionDisabled: boolean
  writePermitted: boolean
  deletePermitted: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleRecover: (id: string) => void
  handleArchive: (id: string) => void
  handleDelete: (id: string) => void
  onCreateSshAccess: (id: string) => void
  onRevokeSshAccess: (id: string) => void
  onScreenRecordings: (id: string) => void
  detailsPath: string
}) {
  const hasCustomName = !!sandbox.name && sandbox.name !== sandbox.id

  return (
    <ScrollArea fade="mask" className="flex-1 min-h-0">
      <div className="flex flex-col pt-2">
        <InfoSection
          title={
            <div className="flex items-center justify-between gap-3">
              <span>Overview</span>
              <Button variant="link" asChild className="h-auto px-0 text-sm tracking-normal normal-case py-0">
                <Link to={detailsPath}>
                  Details
                  <ArrowRight className="size-4" />
                </Link>
              </Button>
            </div>
          }
        >
          <InfoRow label={hasCustomName ? 'Name' : 'Name / UUID'} className="-mr-2">
            <div className="flex items-center gap-1 min-w-0">
              <span className="truncate">{hasCustomName ? sandbox.name : sandbox.id}</span>
              <CopyButton
                value={hasCustomName ? sandbox.name : sandbox.id}
                tooltipText={hasCustomName ? 'Copy name' : 'Copy name / UUID'}
                size="icon-xs"
              />
            </div>
          </InfoRow>
          {hasCustomName && (
            <InfoRow label="UUID" className="-mr-2">
              <div className="flex items-center gap-1 min-w-0">
                <span className="truncate">{sandbox.id}</span>
                <CopyButton value={sandbox.id} tooltipText="Copy UUID" size="icon-xs" />
              </div>
            </InfoRow>
          )}
          <div className="flex items-center justify-between gap-3 pt-3">
            <SandboxStateComponent
              state={sandbox.state}
              errorReason={sandbox.errorReason}
              recoverable={sandbox.recoverable}
            />
            {(writePermitted || deletePermitted) && (
              <div className="flex justify-end">
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
              </div>
            )}
          </div>
        </InfoSection>

        <SandboxInfoPanel sandbox={sandbox} getRegionName={getRegionName} />
      </div>
    </ScrollArea>
  )
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
  getRegionName,
  writePermitted,
  deletePermitted,
  handleRecover,
  onCreateSshAccess,
  onRevokeSshAccess,
  onScreenRecordings,
  onNavigate,
  hasPrev,
  hasNext,
}) => {
  const { currentWidth: sidebarWidth } = useSidebar()
  const sheetContentRef = React.useRef<ResizableSheetContentRef>(null)
  const [activeTab, setActiveTab] = useState<SheetTabValue>('overview')
  const [viewportWidth, setViewportWidth] = useState(() => getViewportWidth())
  const experimentsEnabled = useFeatureFlagEnabled(FeatureFlags.ORGANIZATION_EXPERIMENTS)
  const filesystemEnabled = useFeatureFlagEnabled(FeatureFlags.DASHBOARD_FILESYSTEM)
  const spendingEnabled = useFeatureFlagEnabled(FeatureFlags.SANDBOX_SPENDING)
  const config = useConfig()
  const spendingTabAvailable = experimentsEnabled && spendingEnabled && !!config.analyticsApiUrl
  const isDesktop = viewportWidth >= MOBILE_BREAKPOINT

  useEffect(() => {
    const handleResize = () => {
      setViewportWidth(getViewportWidth())
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  const resizeSheetToTab = useCallback(
    (tab: SheetTabValue, immediate = false) => {
      if (!isDesktop) {
        return
      }

      const nextWidth = clampSheetWidth(getWidthForTab(tab, viewportWidth, sidebarWidth), viewportWidth, sidebarWidth)
      sheetContentRef.current?.resize(nextWidth, {
        duration: TAB_RESIZE_DURATION,
        immediate,
      })
    },
    [isDesktop, sidebarWidth, viewportWidth],
  )

  const resetToOverview = useCallback(
    (immediate = false) => {
      setActiveTab('overview')
      resizeSheetToTab('overview', immediate)
    },
    [resizeSheetToTab],
  )

  const handleTabChange = useCallback(
    (value: string) => {
      const nextTab = value as SheetTabValue
      setActiveTab(nextTab)
      resizeSheetToTab(nextTab)
    },
    [resizeSheetToTab],
  )

  useEffect(() => {
    if (!open) {
      return
    }

    resetToOverview(true)
  }, [open, sandbox?.id, resetToOverview])

  useEffect(() => {
    if (!open || isDesktop) {
      return
    }

    sheetContentRef.current?.resize(viewportWidth, { immediate: true })
  }, [isDesktop, open, viewportWidth])

  useEffect(() => {
    if (!experimentsEnabled && ['logs', 'traces', 'metrics', 'spending'].includes(activeTab)) {
      resetToOverview()
      return
    }

    if (!spendingTabAvailable && activeTab === 'spending') {
      handleTabChange('logs')
    }

    if (filesystemEnabled === false && activeTab === 'filesystem') {
      handleTabChange('terminal')
    }
  }, [activeTab, experimentsEnabled, filesystemEnabled, handleTabChange, resetToOverview, spendingTabAvailable])

  if (!sandbox) return null

  const actionDisabled = sandboxIsLoading[sandbox.id]
  const detailsPath = generatePath(RoutePath.SANDBOX_DETAILS, { sandboxId: sandbox.id })
  const minWidth = isDesktop ? getOverviewWidth(viewportWidth) : viewportWidth
  const maxWidth = isDesktop ? getExpandedWidth(viewportWidth, sidebarWidth) : viewportWidth
  const targetWidth = isDesktop
    ? clampSheetWidth(getWidthForTab(activeTab, viewportWidth, sidebarWidth), viewportWidth, sidebarWidth)
    : viewportWidth

  const handleNavigate = (direction: 'prev' | 'next') => {
    resetToOverview(true)
    onNavigate(direction)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <ResizableSheetContent
        ref={sheetContentRef}
        side="right"
        showCloseButton={false}
        defaultWidth={isDesktop ? targetWidth : viewportWidth}
        minWidth={minWidth}
        maxWidth={maxWidth}
        resizable={isDesktop}
        className="p-0 flex flex-col gap-0 [&>button]:hidden"
      >
        <SheetHeader className="flex flex-row items-start justify-between p-4 px-5 space-y-0 border-b border-border">
          <div className="min-w-0">
            <SheetTitle className="text-2xl font-medium">Sandbox Details</SheetTitle>
          </div>
          <div className="flex flex-wrap items-center justify-end gap-2 shrink-0">
            <Button variant="ghost" size="icon-sm" disabled={!hasPrev} onClick={() => handleNavigate('prev')}>
              <ChevronUp className="size-4" />
              <span className="sr-only">Previous sandbox</span>
            </Button>
            <Button variant="ghost" size="icon-sm" disabled={!hasNext} onClick={() => handleNavigate('next')}>
              <ChevronDown className="size-4" />
              <span className="sr-only">Next sandbox</span>
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => onOpenChange(false)} disabled={actionDisabled}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <SandboxSessionProvider>
          <Tabs value={activeTab} onValueChange={handleTabChange} className="flex-1 min-h-0 gap-0">
            <ScrollArea
              fade="mask"
              horizontal
              vertical={false}
              fadeOffset={36}
              className="h-[41px] shrink-0 border-b border-border"
            >
              <TabsList variant="underline" className="h-[41px] w-max min-w-full border-b-0">
                <TabsTrigger value="overview">Overview</TabsTrigger>
                {experimentsEnabled && (
                  <>
                    <TabsTrigger value="logs">Logs</TabsTrigger>
                    <TabsTrigger value="traces">Traces</TabsTrigger>
                    <TabsTrigger value="metrics">Metrics</TabsTrigger>
                    {spendingTabAvailable && <TabsTrigger value="spending">Spending</TabsTrigger>}
                  </>
                )}
                <TabsTrigger value="terminal">Terminal</TabsTrigger>
                {filesystemEnabled && <TabsTrigger value="filesystem">Filesystem</TabsTrigger>}
                <TabsTrigger value="vnc">VNC</TabsTrigger>
              </TabsList>
            </ScrollArea>

            <TabsContent value="overview" className="m-0 min-h-0 data-[state=active]:flex flex-col">
              <SandboxOverviewTab
                sandbox={sandbox}
                getRegionName={getRegionName}
                actionDisabled={actionDisabled}
                writePermitted={writePermitted}
                deletePermitted={deletePermitted}
                handleStart={handleStart}
                handleStop={handleStop}
                handleRecover={handleRecover}
                handleArchive={handleArchive}
                handleDelete={handleDelete}
                onCreateSshAccess={onCreateSshAccess}
                onRevokeSshAccess={onRevokeSshAccess}
                onScreenRecordings={onScreenRecordings}
                detailsPath={detailsPath}
              />
            </TabsContent>

            {experimentsEnabled && (
              <>
                <TabsContent value="logs" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
                  <SandboxLogsTab sandboxId={sandbox.id} persistFilters={false} />
                </TabsContent>
                <TabsContent value="traces" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
                  <SandboxTracesTab sandboxId={sandbox.id} persistFilters={false} />
                </TabsContent>
                <TabsContent value="metrics" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
                  <SandboxMetricsTab sandboxId={sandbox.id} persistFilters={false} />
                </TabsContent>
                {spendingTabAvailable && (
                  <TabsContent
                    value="spending"
                    className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden"
                  >
                    <SandboxSpendingTab sandboxId={sandbox.id} persistFilters={false} />
                  </TabsContent>
                )}
              </>
            )}

            <TabsContent value="terminal" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
              <SandboxTerminalTab sandbox={sandbox} />
            </TabsContent>
            {filesystemEnabled && (
              <TabsContent value="filesystem" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
                <SandboxFileSystemTab sandbox={sandbox} />
              </TabsContent>
            )}
            <TabsContent value="vnc" className="m-0 min-h-0 data-[state=active]:flex flex-col overflow-hidden">
              <SandboxVncTab sandbox={sandbox} />
            </TabsContent>
          </Tabs>
        </SandboxSessionProvider>
      </ResizableSheetContent>
    </Sheet>
  )
}

export default SandboxDetailsSheet
