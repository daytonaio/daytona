/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  ResizableSheetContent,
  Sheet,
  SheetHeader,
  SheetTitle,
  type ResizableSheetContentRef,
} from '@/components/ui/sheet'
import { useSidebar } from '@/components/ui/sidebar'
import { Skeleton } from '@/components/ui/skeleton'
import { getSandboxQueryErrorStatus, useSandboxQuery } from '@/hooks/queries/useSandboxQuery'
import { useConfig } from '@/hooks/useConfig'
import { useSandboxDetailsWsSync } from '@/hooks/useSandboxWsSync'
import { lazyWithPreload } from '@/lib/lazy'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import { ChevronDown, ChevronUp, Container, X } from 'lucide-react'
import React, { Ref, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { InfoPanelSkeleton } from '../SandboxInfoPanel'
import type { TabValue } from '../SearchParams'

export type SandboxDetailsSheetTabValue = TabValue

export interface SandboxSheetRef {
  open: () => void
  close: () => void
}

export interface SandboxDetailsSheetProps {
  sandboxId: string | null
  ref?: Ref<SandboxSheetRef>
  onOpenChange: (open: boolean) => void
  sandboxIsLoading: Record<string, boolean>
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handlePause: (id: string) => void
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
  defaultTab?: SandboxDetailsSheetTabValue
}

const OVERVIEW_WIDTH = 450
const EXPANDED_WIDTH = 1600
const MOBILE_BREAKPOINT = 1024
const SIDE_BY_SIDE_MIN_WIDTH = 1000
const TAB_RESIZE_DURATION = 0.5
const SandboxDetailsSheetContent = lazyWithPreload(() => import('./SandboxDetailsSheetContent'), { preload: true })

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

function SandboxDetailsSheetSkeleton() {
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden max-w-[450px]">
      <div className="flex h-[41px] shrink-0 items-center gap-4 border-b border-border px-4 border-r">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-4 w-12" />
        <Skeleton className="h-4 w-14" />
        <Skeleton className="h-4 w-16" />
      </div>
      <div className="border-b border-border p-5">
        <Skeleton className="mb-2 h-5 w-48" />
        <Skeleton className="h-4 w-64 max-w-full" />
      </div>
      <ScrollArea fade="mask" className="min-h-0 flex-1">
        <InfoPanelSkeleton />
      </ScrollArea>
    </div>
  )
}

function SandboxDetailsSheetEmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex min-h-0 flex-1 items-center justify-center p-6">
      <Empty className="bg-transparent">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Container className="size-4" />
          </EmptyMedia>
          <EmptyTitle>{title}</EmptyTitle>
          <EmptyDescription>{description}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  )
}

const SandboxDetailsSheet: React.FC<SandboxDetailsSheetProps> = ({
  sandboxId,
  onOpenChange,
  sandboxIsLoading,
  handleStart,
  handleStop,
  handlePause,
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
  defaultTab = 'overview',
  ref,
}) => {
  const [open, setOpen] = useState(false)
  const { currentWidth: sidebarWidth } = useSidebar()
  const sheetContentRef = useRef<ResizableSheetContentRef>(null)
  const isHeaderNavigationRef = useRef(false)
  const initializedOpenSandboxIdRef = useRef<string | null>(null)
  const [internalActiveTab, setInternalActiveTab] = useState<SandboxDetailsSheetTabValue>('overview')
  const activeTab = internalActiveTab
  const [viewportWidth, setViewportWidth] = useState(() => getViewportWidth())
  const config = useConfig()
  const spendingTabAvailable = !!config.analyticsApiUrl
  const isDesktop = viewportWidth >= MOBILE_BREAKPOINT

  const { data: sandbox, error, isError, isLoading, isPending } = useSandboxQuery(sandboxId || '')
  useSandboxDetailsWsSync(sandboxId || '')

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    onOpenChange(isOpen)
  }

  useImperativeHandle(ref, () => ({
    open: () => handleOpenChange(true),
    close: () => handleOpenChange(false),
  }))

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

  const resetToOverview = useCallback(
    (immediate = false) => {
      setInternalActiveTab('overview')
      requestAnimationFrame(() => resizeSheetToTab('overview', immediate))
    },
    [resizeSheetToTab],
  )

  const resetToTab = useCallback(
    (tab: SandboxDetailsSheetTabValue, immediate = false) => {
      setInternalActiveTab(tab)
      requestAnimationFrame(() => resizeSheetToTab(tab, immediate))
    },
    [resizeSheetToTab],
  )

  const handleTabChange = useCallback(
    (value: string) => {
      const nextTab = value as SandboxDetailsSheetTabValue
      setInternalActiveTab(nextTab)
      resizeSheetToTab(nextTab)
    },
    [resizeSheetToTab],
  )

  useEffect(() => {
    if (!open || !sandboxId) {
      initializedOpenSandboxIdRef.current = null
      return
    }

    if (initializedOpenSandboxIdRef.current === sandboxId) {
      return
    }
    initializedOpenSandboxIdRef.current = sandboxId

    if (isHeaderNavigationRef.current) {
      isHeaderNavigationRef.current = false
      return
    }

    resetToTab(defaultTab, true)
  }, [defaultTab, open, resetToTab, sandboxId])

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
  }, [activeTab, handleTabChange, spendingTabAvailable])

  const actionDisabled = sandbox ? !!sandboxIsLoading[sandbox.id] : false

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

  const isNotFound = isError && getSandboxQueryErrorStatus(error) === 404

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
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
        className="p-0 flex flex-col gap-0 data-[state=closed]:slide-out-to-right-[400px] data-[state=closed]:duration-250 data-[state=open]:slide-in-from-right-[400px] ease-[cubic-bezier(0.22,1,0.36,1)] [&>button]:hidden"
      >
        <SheetHeader className="flex flex-row items-start justify-between p-4 px-5 space-y-0 border-b border-border">
          <div className="min-w-0">
            <SheetTitle>Sandbox Details</SheetTitle>
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
            <Button variant="ghost" size="icon-sm" onClick={() => handleOpenChange(false)} disabled={actionDisabled}>
              <X className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>
        </SheetHeader>

        <SandboxSessionProvider>
          {isNotFound ? (
            <SandboxDetailsSheetEmptyState
              title="Sandbox not found"
              description="This sandbox may not exist, or you may not have access to it in this organization."
            />
          ) : sandbox ? (
            <React.Suspense fallback={<SandboxDetailsSheetSkeleton />}>
              <SandboxDetailsSheetContent
                sandbox={sandbox}
                activeTab={activeTab}
                onTabChange={handleTabChange}
                isDesktop={isDesktop}
                spendingTabAvailable={spendingTabAvailable}
                actionDisabled={actionDisabled}
                writePermitted={writePermitted}
                deletePermitted={deletePermitted}
                handleStart={handleStart}
                handleStop={handleStop}
                handlePause={handlePause}
                handleDelete={handleDelete}
                handleArchive={handleArchive}
                handleRecover={handleRecover}
                getRegionName={getRegionName}
                onCreateSshAccess={onCreateSshAccess}
                onRevokeSshAccess={onRevokeSshAccess}
                onScreenRecordings={onScreenRecordings}
                onResetToOverview={resetToOverview}
              />
            </React.Suspense>
          ) : isLoading || isPending ? (
            <SandboxDetailsSheetSkeleton />
          ) : (
            <SandboxDetailsSheetEmptyState
              title="Failed to load sandbox"
              description="Something went wrong while loading this sandbox. Please try again."
            />
          )}
        </SandboxSessionProvider>
      </ResizableSheetContent>
    </Sheet>
  )
}

export default SandboxDetailsSheet
