/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  memo,
  type ComponentProps,
  type HTMLAttributes,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type ReactNode,
} from 'react'

import { cn } from '@/lib/utils'
import { ChevronDownIcon, ChevronRightIcon, FileTextIcon, FolderIcon, FolderOpenIcon } from 'lucide-react'

import { SearchResultLabel } from './searchLabels'
import type { SandboxFileSystemNode } from './types'
import { formatBytes } from './utils'

const MemoizedSearchResultLabel = memo(SearchResultLabel)
const FILE_TREE_INDENT = 16
const FILE_TREE_BASE_PADDING = 4
const FILE_TREE_TOGGLE_SIZE = 32
const FILE_TREE_TOGGLE_CENTER = FILE_TREE_TOGGLE_SIZE / 2
const FILE_TREE_ROW_PADDING_X = 8

type SearchLabelProps = {
  availableWidth: number
  font: string
  query: string
}

export type FileTreeRowProps = Omit<ComponentProps<'div'>, 'children'> & {
  actions?: ReactNode
  itemProps?: HTMLAttributes<HTMLDivElement>
  depth?: number
  dragHandleProps?: HTMLAttributes<HTMLDivElement>
  isDragTarget?: boolean
  isDragTargetAbove?: boolean
  isDragTargetBelow?: boolean
  isDraggingOver?: boolean
  isExpanded?: boolean
  isFocused?: boolean
  isLoading?: boolean
  isSearchResult?: boolean
  isSelected?: boolean
  node: Pick<SandboxFileSystemNode, 'isDir' | 'name' | 'path' | 'size'>
  onActivate: (event: ReactMouseEvent<HTMLDivElement>) => Promise<void> | void
  onItemKeyDown?: (event: ReactKeyboardEvent<HTMLDivElement>) => void
  onToggleExpand?: () => void
  searchLabel?: SearchLabelProps
  top: number
}

export function FileTreeRow({
  actions,
  itemProps,
  className,
  depth = 0,
  dragHandleProps,
  isDragTarget = false,
  isDragTargetAbove = false,
  isDragTargetBelow = false,
  isDraggingOver = false,
  isExpanded = false,
  isFocused = false,
  isLoading = false,
  isSearchResult = false,
  isSelected = false,
  node,
  onActivate,
  onItemKeyDown,
  onToggleExpand,
  searchLabel,
  style,
  top,
  ...props
}: FileTreeRowProps) {
  const itemLabel = isSearchResult ? node.path : node.name || node.path
  const { className: dragHandleClassName, style: dragHandleStyle, ...resolvedDragHandleProps } = dragHandleProps ?? {}
  const { className: itemClassName, style: itemStyle, ...resolvedItemProps } = itemProps ?? {}

  return (
    <div
      {...props}
      {...resolvedDragHandleProps}
      className={cn('group absolute left-0 top-0 flex h-8 w-full items-center px-2', dragHandleClassName, className)}
      style={{ ...dragHandleStyle, ...style, transform: `translateY(${top}px)` }}
    >
      {!isSearchResult && depth > 0 ? (
        <div className="pointer-events-none absolute inset-y-0 left-0">
          {Array.from({ length: depth }).map((_, levelIndex) => (
            <span
              key={levelIndex}
              className="absolute inset-y-0 w-px bg-border/45"
              style={{
                left: `${FILE_TREE_ROW_PADDING_X + FILE_TREE_BASE_PADDING + levelIndex * FILE_TREE_INDENT + FILE_TREE_TOGGLE_CENTER}px`,
              }}
            />
          ))}
        </div>
      ) : null}

      <div
        {...resolvedItemProps}
        data-file-tree-row-button
        onClick={onActivate}
        onKeyDown={onItemKeyDown}
        className={cn(
          'flex h-8 w-full items-center gap-1 rounded-md pr-1 text-sm hover:bg-muted focus-visible:outline-none',
          itemClassName,
          {
            'ring-1 ring-primary/20': isDragTarget,
            'bg-accent/60': isDraggingOver,
            'bg-muted': isSelected,
            'bg-muted/60': isFocused,
            'opacity-60': isSearchResult && isLoading,
          },
        )}
        style={itemStyle}
      >
        {isDragTargetAbove ? <div className="pointer-events-none absolute inset-x-2 top-0 h-px bg-primary" /> : null}
        {isDragTargetBelow ? <div className="pointer-events-none absolute inset-x-2 bottom-0 h-px bg-primary" /> : null}
        <div
          className="flex h-8 min-w-0 flex-1 items-center gap-1"
          style={{
            paddingLeft: `${isSearchResult ? FILE_TREE_BASE_PADDING : depth * FILE_TREE_INDENT + FILE_TREE_BASE_PADDING}px`,
          }}
        >
          {!isSearchResult && node.isDir && onToggleExpand ? (
            <button
              type="button"
              aria-label={isExpanded ? `Collapse ${node.name || node.path}` : `Expand ${node.name || node.path}`}
              tabIndex={-1}
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted/80 hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-border"
              onClick={(event) => {
                event.preventDefault()
                event.stopPropagation()
                onToggleExpand()
              }}
            >
              {isExpanded ? (
                <ChevronDownIcon className="size-4 text-muted-foreground" />
              ) : (
                <ChevronRightIcon className="size-4 text-muted-foreground" />
              )}
            </button>
          ) : null}

          <div className="flex h-8 min-w-0 flex-1 items-center gap-2 rounded-md px-1 text-left">
            {node.isDir ? (
              !isSearchResult && isExpanded ? (
                <FolderOpenIcon className="size-4 shrink-0 text-muted-foreground" />
              ) : (
                <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
              )
            ) : (
              <FileTextIcon className="size-4 shrink-0 text-muted-foreground" />
            )}

            {isSearchResult && searchLabel ? (
              <MemoizedSearchResultLabel
                availableWidth={searchLabel.availableWidth}
                font={searchLabel.font}
                text={itemLabel}
                query={searchLabel.query}
                className={cn('flex-1', {
                  'animate-shimmer-text': isLoading,
                })}
              />
            ) : (
              <span
                className={cn('truncate', {
                  'animate-shimmer-text': isLoading,
                })}
              >
                {itemLabel}
              </span>
            )}

            {!isSearchResult && !node.isDir ? (
              <span
                className={cn('ml-auto shrink-0 text-xs text-muted-foreground', {
                  'animate-shimmer-text': isLoading,
                })}
              >
                {formatBytes(node.size)}
              </span>
            ) : null}
          </div>
        </div>

        {!isSearchResult && actions ? (
          <div className="pointer-events-none ml-1 flex h-8 shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:pointer-events-auto group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:opacity-100 has-[[data-state=open]]:pointer-events-auto has-[[data-state=open]]:opacity-100">
            {actions}
          </div>
        ) : null}
      </div>
    </div>
  )
}
