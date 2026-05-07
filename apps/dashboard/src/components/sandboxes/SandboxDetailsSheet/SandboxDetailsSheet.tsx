/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  ResizableSheetContent,
  Sheet,
  SheetHeader,
  SheetTitle,
  type ResizableSheetContentRef,
} from '@/components/ui/sheet'
import { useSidebar } from '@/components/ui/sidebar'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { useConfig } from '@/hooks/useConfig'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import type { Sandbox } from '@daytona/api-client'
import { ChevronDown, ChevronUp, ChevronsRight, X } from 'lucide-react'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import React, { useCallback, useEffect, useState } from 'react'
import { SandboxActionsSegmented } from '../SandboxActionsSegmented'
import { SandboxInfoPanel } from '../SandboxInfoPanel'
import { SandboxDetailsHeader } from './SandboxDetailsHeader'
import { SandboxDetailsTabContent } from './SandboxDetailsTabContent'
import { SandboxDetailsTabsList, SandboxOverviewTabTrigger } from './SandboxDetailsTabsList'

export type SandboxDetailsSheetTabValue =
  | 'overview'
  | 'logs'
  | 'traces'
  | 'metrics'
  | 'spending'
  | 'terminal'
  | 'filesystem'
  | 'vnc'

export interface SandboxDetailsSheetProps {
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
  initialTab?: SandboxDetailsSheetTabValue
  activeTab?: SandboxDetailsSheetTabValue
  onTabChange?: (tab: SandboxDetailsSheetTabValue) => void
}

const OVERVIEW_WIDTH = 450
const EXPANDED_WIDTH = 1600
const MOBILE_BREAKPOINT = 1024
const SIDE_BY_SIDE_MIN_WIDTH = 1000
const TAB_RESIZE_DURATION = 0.5

function getViewportWidth() {
  return typeof window === 'undefined' ? EXPANDED_WIDTH : window.innerWidth
}

function getOverviewWidth(viewportWidth: number) {
  return Math.min(OVERVIEW_WIDTH, viewportWidth)
}

function getExpandedWidth(viewportWidth: number, sidebarWidth: number) {
  return Math.min(viewportWidth, Math.max(OVERVIEW_WIDTH, Math.min(EXPANDED_WIDTH, viewportWidth - sidebarWidth)))
}

function getWidthForTab(tab: SandboxDetailsSheetTabValue, viewportWidth: number, sidebarWidth: number) {
  return tab === 'overview' ? getOverviewWidth(viewportWidth) : getExpandedWidth(viewportWidth, sidebarWidth)
}

function clampSheetWidth(width: number, viewportWidth: number, sidebarWidth: number) {
  return Math.min(Math.max(width, getOverviewWidth(viewportWidth)), getExpandedWidth(viewportWidth, sidebarWidth))
}

function isNestedAlertDialogEvent(event: Event) {
  return (
    event.target instanceof HTMLElement &&
    Boolean(event.target.closest('[data-slot="alert-dialog-content"], [data-slot="alert-dialog-overlay"]'))
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
  initialTab = 'overview',
  activeTab: activeTabProp,
  onTabChange,
}) => {
  const { currentWidth: sidebarWidth } = useSidebar()
  const sheetContentRef = React.useRef<ResizableSheetContentRef>(null)
  const isHeaderNavigationRef = React.useRef(false)
  const initializedOpenSandboxIdRef = React.useRef<string | null>(null)
  const [internalActiveTab, setInternalActiveTab] = useState<SandboxDetailsSheetTabValue>('overview')
  const activeTab = activeTabProp ?? internalActiveTab
  const [viewportWidth, setViewportWidth] = useState(() => getViewportWidth())
  const filesystemEnabled = useFeatureFlagEnabled(FeatureFlags.DASHBOARD_FILESYSTEM)
  const spendingEnabled = useFeatureFlagEnabled(FeatureFlags.SANDBOX_SPENDING)
  const config = useConfig()
  const spendingTabAvailable = spendingEnabled && !!config.analyticsApiUrl
  const isDesktop = viewportWidth >= MOBILE_BREAKPOINT

  useEffect(() => {
    const handleResize = () => {
      setViewportWidth(getViewportWidth())
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  const resizeSheetToTab = useCallback(
    (tab: SandboxDetailsSheetTabValue, immediate = false) => {
      if (!isDesktop) {
        return
      }

      const nextWidth = clampSheetWidth(getWidthForTab(tab, viewportWidth, sidebarWidth), viewportWidth, sidebarWidth)
      sheetContentRef.current?.resize(nextWidth, {
        duration: TAB_RESIZE_DURATION,
        immediate,
        notify: false,
      })
    },
    [isDesktop, sidebarWidth, viewportWidth],
  )

  const setActiveTabValue = useCallback(
    (tab: SandboxDetailsSheetTabValue) => {
      if (activeTabProp === undefined) {
        setInternalActiveTab(tab)
      }

      onTabChange?.(tab)
    },
    [activeTabProp, onTabChange],
  )

  const resetToOverview = useCallback(
    (immediate = false) => {
      setActiveTabValue('overview')
      requestAnimationFrame(() => resizeSheetToTab('overview', immediate))
    },
    [resizeSheetToTab, setActiveTabValue],
  )

  const resetToTab = useCallback(
    (tab: SandboxDetailsSheetTabValue, immediate = false) => {
      setActiveTabValue(tab)
      requestAnimationFrame(() => resizeSheetToTab(tab, immediate))
    },
    [resizeSheetToTab, setActiveTabValue],
  )

  const handleTabChange = useCallback(
    (value: string) => {
      const nextTab = value as SandboxDetailsSheetTabValue
      setActiveTabValue(nextTab)
      resizeSheetToTab(nextTab)
    },
    [resizeSheetToTab, setActiveTabValue],
  )

  useEffect(() => {
    if (!open) {
      initializedOpenSandboxIdRef.current = null
      return
    }

    if (initializedOpenSandboxIdRef.current === sandbox?.id) {
      return
    }
    initializedOpenSandboxIdRef.current = sandbox?.id ?? null

    if (isHeaderNavigationRef.current) {
      isHeaderNavigationRef.current = false
      return
    }

    resetToTab(initialTab, true)
  }, [initialTab, open, resetToTab, sandbox?.id])

  useEffect(() => {
    if (!open || activeTabProp === undefined) {
      return
    }

    requestAnimationFrame(() => resizeSheetToTab(activeTabProp))
  }, [activeTabProp, open, resizeSheetToTab])

  useEffect(() => {
    if (!open || isDesktop) {
      return
    }

    sheetContentRef.current?.resize(viewportWidth, { immediate: true })
  }, [isDesktop, open, viewportWidth])

  useEffect(() => {
    if (!spendingTabAvailable && activeTab === 'spending') {
      handleTabChange('logs')
    }

    if (filesystemEnabled === false && activeTab === 'filesystem') {
      handleTabChange('terminal')
    }
  }, [activeTab, filesystemEnabled, handleTabChange, resetToOverview, spendingTabAvailable])

  if (!sandbox) return null

  const actionDisabled = sandboxIsLoading[sandbox.id]
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

  const maxWidth = isDesktop ? getExpandedWidth(viewportWidth, sidebarWidth) : viewportWidth
  const minWidth = isDesktop
    ? activeTab === 'overview'
      ? getOverviewWidth(viewportWidth)
      : Math.min(SIDE_BY_SIDE_MIN_WIDTH, maxWidth)
    : viewportWidth
  const targetWidth = isDesktop
    ? clampSheetWidth(getWidthForTab(activeTab, viewportWidth, sidebarWidth), viewportWidth, sidebarWidth)
    : viewportWidth

  const handleNavigate = (direction: 'prev' | 'next') => {
    isHeaderNavigationRef.current = true
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
        resizable={false}
        onPointerDownOutside={(event) => {
          if (isNestedAlertDialogEvent(event)) {
            event.preventDefault()
          }
        }}
        onFocusOutside={(event) => {
          if (isNestedAlertDialogEvent(event)) {
            event.preventDefault()
          }
        }}
        className="p-0 flex flex-col gap-0 data-[state=closed]:slide-out-to-right-[400px] data-[state=closed]:duration-150 data-[state=open]:slide-in-from-right-[400px] ease-[cubic-bezier(0.22,1,0.36,1)] [&>button]:hidden"
      >
        <SheetHeader className="flex flex-row items-start justify-between p-4 px-5 space-y-0 border-b border-border">
          <div className="min-w-0">
            <SheetTitle className="text-2xl font-medium">Sandbox Details</SheetTitle>
          </div>
          <div className="flex flex-wrap items-center justify-end shrink-0">
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
          <Tabs value={activeTab} onValueChange={handleTabChange} className="flex flex-1 min-h-0 flex-col gap-0">
            {isDesktop ? (
              activeTab === 'overview' ? (
                <div className="flex flex-1 min-h-0 flex-col overflow-hidden">
                  <SandboxDetailsTabsList
                    filesystemEnabled={filesystemEnabled}
                    spendingTabAvailable={spendingTabAvailable}
                    showOverview
                  />
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
                        filesystemEnabled={filesystemEnabled}
                        spendingTabAvailable={spendingTabAvailable}
                        showOverview={false}
                        leadingContent={
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                className="mx-1 size-8 text-muted-foreground hover:text-foreground"
                                onClick={() => resetToOverview()}
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
                        <SandboxDetailsTabContent
                          sandbox={sandbox}
                          filesystemEnabled={filesystemEnabled}
                          spendingTabAvailable={spendingTabAvailable}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              )
            ) : (
              <>
                <div className="flex min-h-0 shrink-0 flex-col">
                  <SandboxDetailsHeader sandbox={sandbox} actions={sandboxHeaderActions} />
                  <SandboxDetailsTabsList
                    filesystemEnabled={filesystemEnabled}
                    spendingTabAvailable={spendingTabAvailable}
                    showOverview
                  />
                </div>

                <TabsContent value="overview" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col">
                  {sandboxInfoPanel}
                </TabsContent>

                <SandboxDetailsTabContent
                  sandbox={sandbox}
                  filesystemEnabled={filesystemEnabled}
                  spendingTabAvailable={spendingTabAvailable}
                />
              </>
            )}
          </Tabs>
        </SandboxSessionProvider>
      </ResizableSheetContent>
    </Sheet>
  )
}

export default SandboxDetailsSheet
