/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  memo,
  type ButtonHTMLAttributes,
  type ComponentProps,
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
  buttonProps?: ButtonHTMLAttributes<HTMLButtonElement>
  depth?: number
  isExpanded?: boolean
  isFocused?: boolean
  isLoading?: boolean
  isSearchResult?: boolean
  isSelected?: boolean
  node: Pick<SandboxFileSystemNode, 'isDir' | 'name' | 'path' | 'size'>
  onActivate: (event: ReactMouseEvent<HTMLButtonElement>) => Promise<void> | void
  onItemKeyDown?: (event: ReactKeyboardEvent<HTMLButtonElement>) => void
  onToggleExpand?: () => void
  searchLabel?: SearchLabelProps
  top: number
}

export function FileTreeRow({
  actions,
  buttonProps,
  className,
  depth = 0,
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

  return (
    <div
      {...props}
      className={cn('group absolute left-0 top-0 flex h-8 w-full items-center px-2', className)}
      style={{ ...style, transform: `translateY(${top}px)` }}
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
        className={cn('flex h-8 w-full items-center gap-1 rounded-md pr-1 text-sm hover:bg-muted', {
          'bg-muted': isSelected,
          'bg-muted/60': isFocused,
          'opacity-60': isSearchResult && isLoading,
        })}
      >
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

          <button
            type="button"
            {...buttonProps}
            onClick={onActivate}
            onKeyDown={onItemKeyDown}
            className="flex h-8 min-w-0 flex-1 items-center gap-2 rounded-md px-1 text-left focus-visible:outline-none"
          >
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
          </button>
        </div>

        {!isSearchResult && actions ? (
          <div className="ml-1 flex h-8 shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 has-[[data-state=open]]:opacity-100">
            {actions}
          </div>
        ) : null}
      </div>
    </div>
  )
}
