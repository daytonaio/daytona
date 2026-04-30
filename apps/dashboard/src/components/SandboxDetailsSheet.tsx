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
import { useConfig } from '@/hooks/useConfig'
import { cn } from '@/lib/utils'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import { Sandbox } from '@daytona/api-client'
import { ChevronDown, ChevronLeft, ChevronRight, ChevronUp, ChevronsRight, X } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import React, { useCallback, useEffect, useState } from 'react'
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
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'
import { useSidebar } from './ui/sidebar'

export type SandboxDetailsSheetTabValue =
  | 'overview'
  | 'logs'
  | 'traces'
  | 'metrics'
  | 'spending'
  | 'terminal'
  | 'filesystem'
  | 'vnc'

const OVERVIEW_WIDTH = 450
const EXPANDED_WIDTH = 1600
const MOBILE_BREAKPOINT = 1024
const SIDE_BY_SIDE_MIN_WIDTH = 1000
const TAB_RESIZE_DURATION = 0.5
const TAB_CONTENT_RENDER_DELAY_MS = 0

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
  initialTab?: SandboxDetailsSheetTabValue
  activeTab?: SandboxDetailsSheetTabValue
  onTabChange?: (tab: SandboxDetailsSheetTabValue) => void
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

function SandboxOverviewPanel({
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
  tabs,
  showDetails = true,
  className,
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
  tabs?: React.ReactNode
  showDetails?: boolean
  className?: string
}) {
  const hasCustomName = !!sandbox.name && sandbox.name !== sandbox.id

  return (
    <div className={cn('flex min-h-0 flex-col', showDetails ? 'flex-1' : 'shrink-0', className)}>
      <div className="shrink-0">
        <InfoSection title={null} className="last:border-b">
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
      </div>

      {tabs}

      {showDetails && (
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
      )}
    </div>
  )
}

function SandboxDetailsTabsList({
  experimentsEnabled,
  filesystemEnabled,
  spendingTabAvailable,
  showOverview,
  leadingContent,
}: {
  experimentsEnabled: boolean | undefined
  filesystemEnabled: boolean | undefined
  spendingTabAvailable: boolean | undefined
  showOverview: boolean
  leadingContent?: React.ReactNode
}) {
  const triggerClassName = 'h-[41px] border-b py-0'
  const tabViewportRef = React.useRef<HTMLDivElement | null>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const [isTabStripHovered, setIsTabStripHovered] = useState(false)

  const updateCanScrollRight = useCallback(() => {
    const viewport = tabViewportRef.current
    if (!viewport) {
      setCanScrollLeft(false)
      setCanScrollRight(false)
      return
    }

    const remainingScroll = viewport.scrollWidth - viewport.clientWidth - viewport.scrollLeft
    setCanScrollLeft(viewport.scrollLeft > 1)
    setCanScrollRight(remainingScroll > 1)
  }, [])

  const scrollTabs = useCallback((direction: 'left' | 'right') => {
    const viewport = tabViewportRef.current
    if (!viewport) {
      return
    }

    viewport.scrollBy({
      left: (direction === 'left' ? -1 : 1) * Math.max(180, viewport.clientWidth * 0.7),
      behavior: 'smooth',
    })
  }, [])

  useEffect(() => {
    const viewport = tabViewportRef.current
    if (!viewport) {
      return
    }

    updateCanScrollRight()
    viewport.addEventListener('scroll', updateCanScrollRight, { passive: true })
    window.addEventListener('resize', updateCanScrollRight)

    return () => {
      viewport.removeEventListener('scroll', updateCanScrollRight)
      window.removeEventListener('resize', updateCanScrollRight)
    }
  }, [updateCanScrollRight])

  useEffect(() => {
    updateCanScrollRight()
  }, [experimentsEnabled, filesystemEnabled, spendingTabAvailable, showOverview, leadingContent, updateCanScrollRight])

  return (
    <div
      className="relative flex h-[42px] shrink-0 border-b border-border"
      onMouseEnter={() => setIsTabStripHovered(true)}
      onMouseLeave={() => setIsTabStripHovered(false)}
      onFocusCapture={() => setIsTabStripHovered(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
          setIsTabStripHovered(false)
        }
      }}
    >
      {leadingContent ? <div className="flex h-full shrink-0 items-center">{leadingContent}</div> : null}
      <div className="relative h-full min-w-0 flex-1">
        <ScrollArea
          fade="mask"
          horizontal
          vertical={false}
          fadeOffset={36}
          viewportRef={tabViewportRef}
          className="h-full [&_[data-slot=scroll-area-scrollbar]]:hidden [&_[data-slot=scroll-area-viewport]]:pb-px"
        >
          <TabsList variant="underline" className="h-[41px] w-max min-w-full border-b-0">
            {showOverview && (
              <TabsTrigger value="overview" className={cn(triggerClassName, 'ml-2')}>
                Overview
              </TabsTrigger>
            )}
            {experimentsEnabled && (
              <>
                <TabsTrigger value="logs" className={triggerClassName}>
                  Logs
                </TabsTrigger>
                <TabsTrigger value="traces" className={triggerClassName}>
                  Traces
                </TabsTrigger>
                <TabsTrigger value="metrics" className={triggerClassName}>
                  Metrics
                </TabsTrigger>
                {spendingTabAvailable && (
                  <TabsTrigger value="spending" className={triggerClassName}>
                    Spending
                  </TabsTrigger>
                )}
              </>
            )}
            <TabsTrigger value="terminal" className={triggerClassName}>
              Terminal
            </TabsTrigger>
            {filesystemEnabled && (
              <TabsTrigger value="filesystem" className={triggerClassName}>
                Filesystem
              </TabsTrigger>
            )}
            <TabsTrigger value="vnc" className={triggerClassName}>
              VNC
            </TabsTrigger>
          </TabsList>
        </ScrollArea>
        <AnimatePresence initial={false}>
          {canScrollLeft && isTabStripHovered && (
            <motion.div
              key="scroll-tabs-left"
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -10 }}
              transition={{ duration: 0.16, ease: 'easeOut' }}
              className="pointer-events-none absolute inset-y-0 left-0 z-20 flex w-20 items-center justify-start pl-1 after:absolute after:inset-0 after:z-0 after:bg-background/90 after:[mask-image:linear-gradient(to_right,black_50%,transparent_100%)] after:content-['']"
            >
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="pointer-events-auto relative z-10 size-8 text-muted-foreground hover:text-foreground"
                onClick={() => scrollTabs('left')}
                aria-label="Scroll tabs left"
              >
                <ChevronLeft className="size-4" />
              </Button>
            </motion.div>
          )}
          {canScrollRight && isTabStripHovered && (
            <motion.div
              key="scroll-tabs-right"
              initial={{ opacity: 0, x: 10 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 10 }}
              transition={{ duration: 0.16, ease: 'easeOut' }}
              className="pointer-events-none absolute inset-y-0 right-0 z-20 flex w-20 items-center justify-end pr-1 after:absolute after:inset-0 after:z-0 after:bg-background/90 after:[mask-image:linear-gradient(to_left,black_50%,transparent_100%)] after:content-['']"
            >
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="pointer-events-auto relative z-10 size-8 text-muted-foreground hover:text-foreground"
                onClick={() => scrollTabs('right')}
                aria-label="Scroll tabs right"
              >
                <ChevronRight className="size-4" />
              </Button>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}

function SandboxOverviewTabTrigger() {
  return (
    <div className="h-[42px] shrink-0 border-b border-border">
      <div className="ml-2 inline-flex h-[41px] items-center px-4 py-0 text-sm font-medium text-foreground">
        Overview
      </div>
    </div>
  )
}

function SandboxDetailsTabContent({
  sandbox,
  experimentsEnabled,
  filesystemEnabled,
  spendingTabAvailable,
}: {
  sandbox: Sandbox
  experimentsEnabled: boolean | undefined
  filesystemEnabled: boolean | undefined
  spendingTabAvailable: boolean | undefined
}) {
  return (
    <>
      {experimentsEnabled && (
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
            <TabsContent
              value="spending"
              className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col overflow-hidden"
            >
              <SandboxSpendingTab sandboxId={sandbox.id} />
            </TabsContent>
          )}
        </>
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
  const hasManualExpandedWidthRef = React.useRef(false)
  const initializedOpenSandboxIdRef = React.useRef<string | null>(null)
  const [internalActiveTab, setInternalActiveTab] = useState<SandboxDetailsSheetTabValue>('overview')
  const activeTab = activeTabProp ?? internalActiveTab
  const [renderExpandedContent, setRenderExpandedContent] = useState(false)
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
      hasManualExpandedWidthRef.current = false
      setActiveTabValue('overview')
      requestAnimationFrame(() => resizeSheetToTab('overview', immediate))
    },
    [resizeSheetToTab, setActiveTabValue],
  )

  const resetToTab = useCallback(
    (tab: SandboxDetailsSheetTabValue, immediate = false) => {
      hasManualExpandedWidthRef.current = false
      setActiveTabValue(tab)
      setRenderExpandedContent(false)
      requestAnimationFrame(() => resizeSheetToTab(tab, immediate))
    },
    [resizeSheetToTab, setActiveTabValue],
  )

  const handleTabChange = useCallback(
    (value: string) => {
      const nextTab = value as SandboxDetailsSheetTabValue
      const expandingFromOverview = activeTab === 'overview' && nextTab !== 'overview'
      setActiveTabValue(nextTab)

      if (nextTab === 'overview') {
        setRenderExpandedContent(false)
        hasManualExpandedWidthRef.current = false
        resizeSheetToTab(nextTab)
        return
      }

      if (expandingFromOverview) {
        setRenderExpandedContent(false)
      }

      if (!hasManualExpandedWidthRef.current) {
        resizeSheetToTab(nextTab)
      }
    },
    [activeTab, resizeSheetToTab, setActiveTabValue],
  )

  useEffect(() => {
    if (!isDesktop || activeTab === 'overview') {
      setRenderExpandedContent(false)
      return
    }

    if (renderExpandedContent) {
      return
    }

    const timeout = window.setTimeout(() => {
      setRenderExpandedContent(true)
    }, TAB_CONTENT_RENDER_DELAY_MS)

    return () => window.clearTimeout(timeout)
  }, [activeTab, isDesktop, renderExpandedContent])

  const handleSheetWidthChange = useCallback(
    (width: number) => {
      hasManualExpandedWidthRef.current =
        isDesktop && activeTab !== 'overview' && width < getExpandedWidth(viewportWidth, sidebarWidth)
    },
    [activeTab, isDesktop, sidebarWidth, viewportWidth],
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

    hasManualExpandedWidthRef.current = false
    resetToTab(initialTab, true)
  }, [initialTab, open, resetToTab, sandbox?.id])

  useEffect(() => {
    if (!open || activeTabProp === undefined) {
      return
    }

    if (activeTabProp === 'overview') {
      setRenderExpandedContent(false)
      hasManualExpandedWidthRef.current = false
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
        onWidthChange={handleSheetWidthChange}
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
                    experimentsEnabled={experimentsEnabled}
                    filesystemEnabled={filesystemEnabled}
                    spendingTabAvailable={spendingTabAvailable}
                    showOverview
                  />
                  <SandboxOverviewPanel
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
                    className="max-w-[450px]"
                  />
                </div>
              ) : (
                <div className="flex flex-1 min-h-0 overflow-hidden">
                  <div className="flex w-[450px] shrink-0 min-h-0 flex-col border-r border-border">
                    <SandboxOverviewTabTrigger />
                    <SandboxOverviewPanel
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
                      showDetails={false}
                    />
                    <ScrollArea fade="mask" className="min-h-0 flex-1">
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
                  </div>
                  <div className="flex min-w-0 min-h-0 flex-1 flex-col overflow-hidden">
                    <div className="animate-in fade-in slide-in-from-left-3 duration-500 ease-out">
                      <SandboxDetailsTabsList
                        experimentsEnabled={experimentsEnabled}
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
                      {renderExpandedContent && (
                        <div className="flex min-h-0 flex-1 flex-col animate-in fade-in slide-in-from-right-20 duration-300 ease-out">
                          <SandboxDetailsTabContent
                            sandbox={sandbox}
                            experimentsEnabled={experimentsEnabled}
                            filesystemEnabled={filesystemEnabled}
                            spendingTabAvailable={spendingTabAvailable}
                          />
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )
            ) : (
              <>
                <SandboxOverviewPanel
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
                  showDetails={false}
                  tabs={
                    <SandboxDetailsTabsList
                      experimentsEnabled={experimentsEnabled}
                      filesystemEnabled={filesystemEnabled}
                      spendingTabAvailable={spendingTabAvailable}
                      showOverview
                    />
                  }
                />

                <TabsContent value="overview" className="m-0 min-h-0 flex-1 data-[state=active]:flex flex-col">
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
                </TabsContent>

                <SandboxDetailsTabContent
                  sandbox={sandbox}
                  experimentsEnabled={experimentsEnabled}
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
