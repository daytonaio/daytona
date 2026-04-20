/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ArrowDownIcon, ArrowUpIcon, TextWrapIcon, XIcon } from 'lucide-react'
import { useEffect, useState, type ReactNode } from 'react'

import TooltipButton from '@/components/TooltipButton'
import { Toggle } from '@/components/ui/toggle'

import { LARGE_TEXT_WRAP_THRESHOLD } from './constants'
import { useFileSystemStore } from './fileSystemStore'
import { PathHeaderLabel } from './searchLabels'
import { getNodeMetaLine } from './utils'

export function FileContentsPanel({
  actions,
  canNavigateNext,
  canNavigatePrevious,
  children,
  isContentsRefreshing,
  onClose,
  onNavigateNext,
  onNavigatePrevious,
  overlay = false,
  showWrapToggle = false,
}: {
  actions?: ReactNode
  canNavigateNext: boolean
  canNavigatePrevious: boolean
  children: (args: { isWrapEnabled: boolean }) => ReactNode
  isContentsRefreshing: boolean
  onClose: () => void
  onNavigateNext: () => void
  onNavigatePrevious: () => void
  overlay?: boolean
  showWrapToggle?: boolean
}) {
  const node = useFileSystemStore((state) => state.selectedNode)
  const headerText = node?.path ?? 'Contents'
  const [isWrapEnabled, setIsWrapEnabled] = useState(true)

  useEffect(() => {
    setIsWrapEnabled((node?.size ?? 0) <= LARGE_TEXT_WRAP_THRESHOLD)
  }, [node?.path, node?.size])

  return (
    <div
      className={
        overlay ? 'flex h-full w-full min-h-0 flex-col bg-background' : 'flex h-full min-h-0 flex-col bg-background'
      }
    >
      <div className="flex h-11 shrink-0 items-center gap-1 border-b border-border px-3">
        <PathHeaderLabel text={headerText} className="flex-1 text-sm font-medium" />
        {node ? (
          <>
            <TooltipButton
              tooltipText="Previous file"
              variant="ghost"
              size="icon-sm"
              onClick={onNavigatePrevious}
              disabled={!canNavigatePrevious}
            >
              <ArrowUpIcon className="size-4" />
            </TooltipButton>
            <TooltipButton
              tooltipText="Next file"
              variant="ghost"
              size="icon-sm"
              onClick={onNavigateNext}
              disabled={!canNavigateNext}
            >
              <ArrowDownIcon className="size-4" />
            </TooltipButton>
            <TooltipButton tooltipText="Close contents" variant="ghost" size="icon-sm" onClick={onClose}>
              <XIcon className="size-4" />
            </TooltipButton>
          </>
        ) : null}
      </div>
      <div className="flex shrink-0 items-center gap-3 px-3 py-3 text-xs text-muted-foreground">
        <span>{node ? getNodeMetaLine(node) : ''}</span>
        <div className="ml-auto flex items-center gap-1">
          {showWrapToggle ? (
            <Toggle
              size="sm"
              variant="outline"
              pressed={isWrapEnabled}
              onPressedChange={setIsWrapEnabled}
              aria-label="Toggle wrapped lines"
            >
              <TextWrapIcon className="size-4" />
            </Toggle>
          ) : null}
          {actions}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-hidden px-3 pb-4">
        <div
          key={node?.path ?? 'empty'}
          className={
            isContentsRefreshing ? 'h-full min-h-0 opacity-60 transition-opacity' : 'h-full min-h-0 transition-opacity'
          }
        >
          {children({ isWrapEnabled })}
        </div>
      </div>
    </div>
  )
}
