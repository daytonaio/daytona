/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon, DownloadIcon, EllipsisIcon } from 'lucide-react'
import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

import type { SandboxFileSystemNode } from './types'

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)
const copyIconMotionProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

export type FileNodeActionsProps = {
  canDelete?: boolean
  className?: string
  isDropdownOpen?: boolean
  isRefreshing?: boolean
  node: SandboxFileSystemNode
  onCopy?: () => void | Promise<void>
  onDelete?: () => void
  onDownload?: () => void | Promise<void>
  onDropdownOpenChange: (open: boolean) => void
  onRefresh: () => void | Promise<void>
  onStartCreateFolder?: () => void
  primaryAction?:
    | {
        copied?: boolean
        kind: 'copy'
        onClick: () => void | Promise<void>
      }
    | {
        kind: 'download'
        onClick: () => void | Promise<void>
      }
  triggerTabIndex?: number
  variant: 'compact' | 'compound'
}

export function FileNodeActions({
  canDelete = true,
  className,
  isDropdownOpen = false,
  isRefreshing = false,
  node,
  onCopy,
  onDelete,
  onDownload,
  onDropdownOpenChange,
  onRefresh,
  onStartCreateFolder,
  primaryAction,
  triggerTabIndex = 0,
  variant,
}: FileNodeActionsProps) {
  const handleRefresh = async () => {
    await onRefresh()
  }

  const handleDownload = async () => {
    if (!onDownload) {
      return
    }

    await onDownload()
  }

  const handleCopy = async () => {
    if (!onCopy) {
      return
    }

    await onCopy()
  }

  const handlePrimaryActionClick = async () => {
    if (!primaryAction) {
      return
    }

    await primaryAction.onClick()
  }

  const dropdownContent = (
    <DropdownMenuContent
      align="end"
      side={variant === 'compact' ? 'right' : 'bottom'}
      className={cn({
        'w-44': node.isDir,
        'w-48': !node.isDir,
      })}
      onCloseAutoFocus={(event) => event.preventDefault()}
    >
      <DropdownMenuItem onClick={handleRefresh} disabled={isRefreshing}>
        Refresh
      </DropdownMenuItem>

      {!node.isDir && onDownload ? <DropdownMenuItem onClick={handleDownload}>Download</DropdownMenuItem> : null}

      {!node.isDir && onCopy ? <DropdownMenuItem onClick={handleCopy}>Copy contents</DropdownMenuItem> : null}

      {node.isDir && onStartCreateFolder ? (
        <DropdownMenuItem onSelect={() => onStartCreateFolder()}>Create folder</DropdownMenuItem>
      ) : null}

      {canDelete ? <DropdownMenuSeparator /> : null}

      {canDelete && onDelete ? (
        <DropdownMenuItem variant="destructive" onClick={onDelete}>
          Delete
        </DropdownMenuItem>
      ) : null}
    </DropdownMenuContent>
  )

  if (variant === 'compound') {
    const showDownloadButton = !node.isDir && Boolean(onDownload) && primaryAction?.kind !== 'download'

    return (
      <ButtonGroup className={className}>
        {primaryAction ? (
          <TooltipButton
            tooltipText={primaryAction.kind === 'copy' ? 'Copy contents' : 'Download'}
            variant="outline"
            size="icon-sm"
            onClick={handlePrimaryActionClick}
          >
            {primaryAction.kind === 'copy' ? (
              <AnimatePresence initial={false} mode="wait">
                {primaryAction.copied ? (
                  <MotionCheckIcon key="copied" className="size-4 text-success" {...copyIconMotionProps} />
                ) : (
                  <MotionCopyIcon key="copy" className="size-4" {...copyIconMotionProps} />
                )}
              </AnimatePresence>
            ) : (
              <DownloadIcon className="size-4" />
            )}
          </TooltipButton>
        ) : null}
        {showDownloadButton || (!primaryAction && !node.isDir && onDownload) ? (
          <TooltipButton tooltipText="Download" variant="outline" size="icon-sm" onClick={handleDownload}>
            <DownloadIcon className="size-4" />
          </TooltipButton>
        ) : null}
        <DropdownMenu onOpenChange={onDropdownOpenChange}>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="icon-sm" className="text-muted-foreground" aria-label="More actions">
              <EllipsisIcon className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          {dropdownContent}
        </DropdownMenu>
      </ButtonGroup>
    )
  }

  return (
    <div
      className={cn(
        'ml-1 flex h-8 shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100',
        {
          'opacity-100': isDropdownOpen,
        },
        className,
      )}
    >
      <DropdownMenu onOpenChange={onDropdownOpenChange}>
        <DropdownMenuTrigger
          asChild
          onClick={(event) => event.stopPropagation()}
          onPointerDown={(event) => event.stopPropagation()}
        >
          <Button
            variant="ghost"
            size="icon-sm"
            tabIndex={triggerTabIndex}
            className="inline-flex h-8 w-8 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <EllipsisIcon className="size-4" />
          </Button>
        </DropdownMenuTrigger>
        {dropdownContent}
      </DropdownMenu>
    </div>
  )
}
