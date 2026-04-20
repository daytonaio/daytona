/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { memo, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState, type KeyboardEvent } from 'react'

import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { asyncDataLoaderFeature } from '@headless-tree/core'
import { useTree } from '@headless-tree/react'
import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { useVirtualizer } from '@tanstack/react-virtual'
import {
  ChevronDownIcon,
  ChevronRightIcon,
  FileTextIcon,
  FolderIcon,
  FolderOpenIcon,
  RefreshCwIcon,
} from 'lucide-react'

import {
  FILE_SEARCH_MIN_CHARS,
  FILE_SEARCH_RESULT_LABEL_RESERVED_WIDTH,
  FILE_TREE_BASE_PADDING,
  FILE_TREE_EDGE_PADDING,
  FILE_TREE_INDENT,
  FILE_TREE_ROW_PADDING_X,
  FILE_TREE_TOGGLE_CENTER,
  ROOT_NODE,
  ROOT_PATH,
} from './constants'
import { FileNodeActions } from './FileNodeActions'
import { FileSearchHeader } from './FileSearchHeader'
import { useFileSystemStore } from './fileSystemStore'
import {
  fileSystemQueryKeys,
  getDirectoryChildrenQueryOptions,
  invalidateDirectoryQuery,
  useFileSearchQuery,
} from './queries'
import { SearchResultLabel } from './searchLabels'
import type { PreviewState, SandboxFileSystemNode, SandboxInstance } from './types'
import {
  createFallbackNode,
  formatBytes,
  getAncestorPaths,
  getCanvasFont,
  getImageMimeType,
  getParentPath,
  toNode,
} from './utils'

export type FileTreePaneHandle = {
  expandPathAncestors: (path: string) => Promise<void>
  getAdjacentFileNode: (currentPath: string, direction: -1 | 1) => SandboxFileSystemNode | null
  getNode: (path: string) => SandboxFileSystemNode
  refreshPath: (path: string) => Promise<void>
  restoreFocus: (path: string) => void
}

const MemoizedSearchResultLabel = memo(SearchResultLabel)
const MemoizedFileSearchHeader = memo(FileSearchHeader)

function FileTreeSkeleton() {
  return (
    <div className="space-y-1 p-2">
      {Array.from({ length: 20 }).map((_, index) => (
        <div key={index} className="flex items-center gap-2 rounded-md px-2 py-1">
          <Skeleton className="size-4 shrink-0 rounded-sm" />
          <Skeleton
            className={cn('h-4 rounded-sm', {
              'w-28': index % 3 === 1,
              'w-36': index % 3 === 0,
              'w-44': index % 3 === 2,
            })}
          />
          <Skeleton className="ml-auto h-3 w-12 rounded-sm" />
        </div>
      ))}
    </div>
  )
}

const defaultSearchResults: string[] = []

function isFallbackNode(node: SandboxFileSystemNode) {
  return (
    node.path !== ROOT_PATH &&
    node.size === 0 &&
    !node.isDir &&
    !node.modTime &&
    !node.mode &&
    !node.owner &&
    !node.group &&
    !node.permissions
  )
}

export function FileTreePane({
  onCopyNodeContents,
  onDownloadNode,
  onSelectedDirectoryRefreshingChange,
  onVisibleFileNavigationChange,
  previewState,
  ref,
  sandboxInstance,
}: {
  onCopyNodeContents: (node: SandboxFileSystemNode) => void | Promise<void>
  onDownloadNode: (node: SandboxFileSystemNode) => void | Promise<void>
  onSelectedDirectoryRefreshingChange: (value: boolean) => void
  onVisibleFileNavigationChange: (value: { canNavigateNext: boolean; canNavigatePrevious: boolean }) => void
  previewState: PreviewState
  ref?: React.Ref<FileTreePaneHandle>
  sandboxInstance: SandboxInstance | undefined
}) {
  const queryClient = useQueryClient()
  const [rootLoadFailed, setRootLoadFailed] = useState(false)
  const itemDataRef = useRef<Record<string, SandboxFileSystemNode>>({ [ROOT_PATH]: ROOT_NODE })
  const fileTreeScrollAreaRef = useRef<HTMLDivElement>(null)

  const selectedNode = useFileSystemStore((state) => state.selectedNode)
  const isSearchOpen = useFileSystemStore((state) => state.isSearchOpen)
  const openDropdownPath = useFileSystemStore((state) => state.openDropdownPath)
  const searchLabelAvailableWidth = useFileSystemStore((state) => state.searchLabelAvailableWidth)
  const searchLabelFont = useFileSystemStore((state) => state.searchLabelFont)
  const searchQuery = useFileSystemStore((state) => state.searchQuery)
  const {
    openNode,
    resetSearch,
    setDeleteTarget,
    setFolderCreationParentPath,
    setNewFolderName,
    setOpenDropdownPath,
    setSearchLabelMeasurements,
  } = useFileSystemStore((state) => state.actions)

  async function loadDirectory(path: string) {
    if (!sandboxInstance) {
      return []
    }

    try {
      const children = await queryClient.fetchQuery(
        getDirectoryChildrenQueryOptions({
          path,
          sandboxInstance,
        }),
      )

      children.forEach((node) => {
        itemDataRef.current[node.id] = node
      })

      if (path === ROOT_PATH) {
        setRootLoadFailed(false)
      }

      return children.map((node) => ({
        id: node.id,
        data: node,
      }))
    } catch {
      if (path === ROOT_PATH) {
        setRootLoadFailed(true)
      }

      return []
    }
  }

  const resolveNode = useCallback(
    async (path: string) => {
      const existingNode = itemDataRef.current[path]
      if (existingNode && !isFallbackNode(existingNode)) {
        return existingNode
      }

      if (!sandboxInstance) {
        return existingNode ?? createFallbackNode(path)
      }

      try {
        const fileInfo = await sandboxInstance.fs.getFileDetails(path)
        const resolvedNode = {
          ...toNode(getParentPath(path), fileInfo),
          id: path,
          path,
        }
        itemDataRef.current[path] = resolvedNode
        return resolvedNode
      } catch {
        return existingNode ?? createFallbackNode(path)
      }
    },
    [sandboxInstance],
  )

  const searchEnabled = isSearchOpen && searchQuery.trim().length >= FILE_SEARCH_MIN_CHARS
  const searchQueryResult = useFileSearchQuery({
    enabled: Boolean(sandboxInstance) && searchEnabled,
    query: searchQuery,
    sandboxInstance,
  })
  const searchResultPaths = searchQueryResult.data ?? defaultSearchResults
  const searchResults = useMemo(() => {
    return searchResultPaths.map((path) => itemDataRef.current[path] ?? createFallbackNode(path))
  }, [searchResultPaths])
  const isSearchLoading = searchQueryResult.isFetching
  const searchFailed = searchQueryResult.isError

  const tree = useTree<SandboxFileSystemNode | null>({
    rootItemId: ROOT_PATH,
    initialState: { expandedItems: [ROOT_PATH] },
    createLoadingItemData: () => null,
    getItemName: (item) => {
      const data = item.getItemData()
      return data?.name ?? itemDataRef.current[item.getId()]?.name ?? item.getId()
    },
    isItemFolder: (item) => Boolean(item.getItemData()?.isDir ?? itemDataRef.current[item.getId()]?.isDir),
    onPrimaryAction: (item) => {
      const node = item.getItemData() ?? itemDataRef.current[item.getId()] ?? createFallbackNode(item.getId())
      openNode(node)
    },
    dataLoader: {
      getItem: async (itemId) => itemDataRef.current[itemId] ?? createFallbackNode(itemId),
      getChildrenWithData: async (itemId) => loadDirectory(itemId),
    },
    features: [asyncDataLoaderFeature],
  })

  const visibleItems = tree.getItems()
  const selectedPath = selectedNode?.path ?? ''
  const rootItem = tree.getItemInstance(ROOT_PATH)
  const rootDirectoryFetchCount = useIsFetching({
    queryKey: sandboxInstance ? fileSystemQueryKeys.directory(sandboxInstance.id, ROOT_PATH) : undefined,
  })
  const isInitialTreeLoading = (!sandboxInstance || rootItem.isLoading()) && visibleItems.length <= 1 && !rootLoadFailed
  const isTreeRefreshing = !isInitialTreeLoading && (rootItem.isLoading() || rootDirectoryFetchCount > 0)
  const selectedItem = selectedPath ? tree.getItemInstance(selectedPath) : null
  const isSelectedDirectoryRefreshing = Boolean(selectedNode?.isDir && selectedItem?.isLoading())
  const visibleFileItems = useMemo(() => {
    return visibleItems.filter((item) => {
      const node = item.getItemData() ?? itemDataRef.current[item.getId()] ?? createFallbackNode(item.getId())
      return !node.isDir
    })
  }, [visibleItems])

  const activeFileNodes = useMemo(() => {
    if (searchEnabled) {
      return searchResults.filter((node) => !node.isDir)
    }

    return visibleFileItems.map(
      (item) => item.getItemData() ?? itemDataRef.current[item.getId()] ?? createFallbackNode(item.getId()),
    )
  }, [searchEnabled, searchResults, visibleFileItems])

  const selectedFileIndex = useMemo(() => {
    if (!selectedNode || selectedNode.isDir) {
      return -1
    }

    return activeFileNodes.findIndex((node) => node.path === selectedPath)
  }, [activeFileNodes, selectedNode, selectedPath])

  useEffect(() => {
    onSelectedDirectoryRefreshingChange(isSelectedDirectoryRefreshing)
  }, [isSelectedDirectoryRefreshing, onSelectedDirectoryRefreshingChange])

  useEffect(() => {
    onVisibleFileNavigationChange({
      canNavigatePrevious: selectedFileIndex > 0,
      canNavigateNext: selectedFileIndex >= 0 && selectedFileIndex < activeFileNodes.length - 1,
    })
  }, [activeFileNodes.length, onVisibleFileNavigationChange, selectedFileIndex])

  const fileTreeVirtualizer = useVirtualizer({
    count: searchEnabled ? searchResults.length : visibleItems.length,
    getScrollElement: () =>
      fileTreeScrollAreaRef.current?.querySelector<HTMLDivElement>('[data-slot="scroll-area-viewport"]') ?? null,
    estimateSize: () => 32,
    overscan: 12,
  })

  const focusedItemIndex = visibleItems.findIndex((item) => item.isFocused())

  useEffect(() => {
    if (focusedItemIndex < 0 || searchEnabled) {
      return
    }

    fileTreeVirtualizer.scrollToIndex(focusedItemIndex, { align: 'auto' })
  }, [fileTreeVirtualizer, focusedItemIndex, searchEnabled])

  useEffect(() => {
    const viewport = fileTreeScrollAreaRef.current?.querySelector<HTMLDivElement>('[data-slot="scroll-area-viewport"]')
    if (!viewport) {
      return
    }

    const updateMeasurements = () => {
      setSearchLabelMeasurements({
        availableWidth: Math.max(0, viewport.getBoundingClientRect().width - FILE_SEARCH_RESULT_LABEL_RESERVED_WIDTH),
        font: getCanvasFont(viewport),
      })
    }

    updateMeasurements()

    const resizeObserver = new ResizeObserver(() => {
      updateMeasurements()
    })

    resizeObserver.observe(viewport)
    return () => resizeObserver.disconnect()
  }, [setSearchLabelMeasurements])

  useEffect(() => {
    const frameId = window.requestAnimationFrame(() => {
      tree.getItems()[0]?.setFocused()
      tree.updateDomFocus()
    })

    return () => window.cancelAnimationFrame(frameId)
  }, [tree])

  const refreshPath = useCallback(
    async (path: string) => {
      if (!sandboxInstance) {
        return
      }

      await invalidateDirectoryQuery({
        path,
        queryClient,
        sandboxInstance,
      })

      if (path === ROOT_PATH) {
        await rootItem.invalidateChildrenIds(true)
        return
      }

      await tree.getItemInstance(path).invalidateChildrenIds(true)
    },
    [queryClient, rootItem, sandboxInstance, tree],
  )

  const expandPathAncestors = useCallback(
    async (path: string) => {
      const ancestorPaths = getAncestorPaths(path)

      for (const ancestorPath of ancestorPaths) {
        const item = tree.getItemInstance(ancestorPath)
        item.expand()
        await item.invalidateChildrenIds()
      }
    },
    [tree],
  )

  const restoreFocus = useCallback(
    (path: string) => {
      const item = tree.getItemInstance(path)
      item?.setFocused()
      tree.updateDomFocus()
    },
    [tree],
  )

  const getAdjacentFileNode = useCallback(
    (currentPath: string, direction: -1 | 1) => {
      const currentIndex = activeFileNodes.findIndex((node) => node.path === currentPath)
      const nextNode = currentIndex >= 0 ? activeFileNodes[currentIndex + direction] : null

      if (!nextNode) {
        return null
      }

      if (searchEnabled) {
        fileTreeVirtualizer.scrollToIndex(currentIndex + direction, { align: 'auto' })
        return nextNode
      }

      const nextItem = visibleFileItems.find((item) => item.getId() === nextNode.path)
      nextItem?.setFocused()
      tree.updateDomFocus()

      return nextNode
    },
    [activeFileNodes, fileTreeVirtualizer, searchEnabled, tree, visibleFileItems],
  )

  const getNode = useCallback((path: string) => itemDataRef.current[path] ?? createFallbackNode(path), [])

  const handleRefreshRoot = useCallback(async () => {
    await refreshPath(ROOT_PATH)
  }, [refreshPath])

  useImperativeHandle(
    ref,
    () => ({
      expandPathAncestors,
      getAdjacentFileNode,
      getNode,
      refreshPath,
      restoreFocus,
    }),
    [expandPathAncestors, getAdjacentFileNode, getNode, refreshPath, restoreFocus],
  )

  const handleItemKeyDown = useCallback(
    (item: ReturnType<typeof tree.getItemInstance>, event: KeyboardEvent<HTMLButtonElement>) => {
      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault()
          tree.focusNextItem()
          tree.updateDomFocus()
          return
        case 'ArrowUp':
          event.preventDefault()
          tree.focusPreviousItem()
          tree.updateDomFocus()
          return
        case 'ArrowRight':
          event.preventDefault()
          if (item.isFolder() && !item.isExpanded()) {
            item.expand()
            return
          }

          tree.focusNextItem()
          tree.updateDomFocus()
          return
        case 'ArrowLeft':
          event.preventDefault()
          if (item.isFolder() && item.isExpanded()) {
            item.collapse()
            return
          }

          item.getParent()?.setFocused()
          tree.updateDomFocus()
          return
        case 'Home':
          event.preventDefault()
          tree.getItems()[0]?.setFocused()
          tree.updateDomFocus()
          return
        case 'End':
          event.preventDefault()
          tree.getItems()[tree.getItems().length - 1]?.setFocused()
          tree.updateDomFocus()
          return
        default:
          return
      }
    },
    [tree],
  )

  if (rootLoadFailed) {
    return (
      <div className="flex flex-1 min-h-0">
        <Empty className="border-0">
          <EmptyHeader>
            <EmptyTitle>Failed to load filesystem</EmptyTitle>
            <EmptyDescription>Something went wrong while listing the sandbox root directory.</EmptyDescription>
          </EmptyHeader>
          <Button variant="outline" size="sm" onClick={() => void refreshPath(ROOT_PATH)}>
            <RefreshCwIcon className="size-4" />
            Retry
          </Button>
        </Empty>
      </div>
    )
  }

  return (
    <div ref={fileTreeScrollAreaRef} className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <MemoizedFileSearchHeader isRefreshing={isTreeRefreshing} onRefresh={handleRefreshRoot} />

      <ScrollArea fade="mask" className="min-h-0 flex-1">
        {isInitialTreeLoading || (searchEnabled && isSearchLoading && searchResults.length === 0) ? (
          <FileTreeSkeleton />
        ) : searchEnabled && searchResults.length === 0 && !isSearchLoading && !searchFailed ? (
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyTitle>0 results</EmptyTitle>
              <EmptyDescription>No files matched "{searchQuery}". Try another search or clear search.</EmptyDescription>
            </EmptyHeader>
            <Button variant="outline" size="sm" onClick={resetSearch}>
              Clear search
            </Button>
          </Empty>
        ) : (
          <div
            {...(!searchEnabled ? tree.getContainerProps('Sandbox filesystem') : {})}
            className={cn('relative min-h-full px-2 transition-opacity', {
              'opacity-60': isTreeRefreshing,
            })}
            style={{ height: `${fileTreeVirtualizer.getTotalSize() + FILE_TREE_EDGE_PADDING * 2}px` }}
          >
            {fileTreeVirtualizer.getVirtualItems().map((virtualItem) => {
              const item = searchEnabled ? null : visibleItems[virtualItem.index]
              const node = searchEnabled
                ? searchResults[virtualItem.index]
                : (item?.getItemData() ?? itemDataRef.current[item?.getId() ?? ''] ?? ROOT_NODE)

              if (!node) {
                return null
              }

              const isDirectory = searchEnabled ? node.isDir : (item?.isFolder() ?? node.isDir)
              const isSelected = selectedPath === node.path
              const isPreviewLoading = previewState.status === 'loading' && previewState.path === node.path
              const isFocused = searchEnabled ? false : (item?.isFocused() ?? false)
              const isLoading = (item?.isLoading() ?? false) || isPreviewLoading
              const itemButtonProps = !searchEnabled ? (item?.getProps() ?? {}) : {}
              const itemLabel = searchEnabled ? node.path : node.name || node.path

              return (
                <div
                  key={searchEnabled ? node.path : item?.getId()}
                  className="group absolute left-0 top-0 flex h-8 w-full items-center px-2"
                  style={{ transform: `translateY(${virtualItem.start + FILE_TREE_EDGE_PADDING}px)` }}
                >
                  {!searchEnabled && item && item.getItemMeta().level > 0 ? (
                    <div className="pointer-events-none absolute inset-y-0 left-0">
                      {Array.from({ length: item.getItemMeta().level }).map((_, levelIndex) => (
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
                      'opacity-60': isSearchLoading && searchEnabled,
                    })}
                  >
                    <div
                      className="flex h-8 min-w-0 flex-1 items-center gap-1"
                      style={{
                        paddingLeft: `${searchEnabled ? FILE_TREE_BASE_PADDING : (item?.getItemMeta().level ?? 0) * FILE_TREE_INDENT + FILE_TREE_BASE_PADDING}px`,
                      }}
                    >
                      {!searchEnabled && isDirectory ? (
                        <button
                          type="button"
                          aria-label={
                            item?.isExpanded()
                              ? `Collapse ${node.name || node.path}`
                              : `Expand ${node.name || node.path}`
                          }
                          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted/80 hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-border"
                          onClick={(event) => {
                            event.preventDefault()
                            event.stopPropagation()

                            if (item?.isExpanded()) {
                              item.collapse()
                            } else {
                              item?.expand()
                            }
                          }}
                        >
                          {item?.isExpanded() ? (
                            <ChevronDownIcon className="size-4 text-muted-foreground" />
                          ) : (
                            <ChevronRightIcon className="size-4 text-muted-foreground" />
                          )}
                        </button>
                      ) : null}

                      <button
                        type="button"
                        {...itemButtonProps}
                        onClick={async (event) => {
                          itemButtonProps.onClick?.(event)

                          if (event.defaultPrevented) {
                            return
                          }

                          if (!searchEnabled && item && node.isDir && !item.isExpanded()) {
                            item.expand()
                          }

                          const resolvedNode = searchEnabled ? await resolveNode(node.path) : node
                          openNode(resolvedNode)
                        }}
                        onKeyDown={!searchEnabled && item ? (event) => handleItemKeyDown(item, event) : undefined}
                        className="flex h-8 min-w-0 flex-1 items-center gap-2 rounded-md px-1 text-left focus-visible:outline-none"
                      >
                        {isDirectory ? (
                          !searchEnabled && item?.isExpanded() ? (
                            <FolderOpenIcon className="size-4 shrink-0 text-muted-foreground" />
                          ) : (
                            <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
                          )
                        ) : (
                          <FileTextIcon className="size-4 shrink-0 text-muted-foreground" />
                        )}

                        {searchEnabled ? (
                          <MemoizedSearchResultLabel
                            availableWidth={searchLabelAvailableWidth}
                            font={searchLabelFont}
                            text={itemLabel}
                            query={searchQuery}
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

                        {!searchEnabled && !isDirectory ? (
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

                    {!searchEnabled ? (
                      <FileNodeActions
                        variant="compact"
                        node={node}
                        isDropdownOpen={openDropdownPath === node.path}
                        triggerTabIndex={isFocused || isSelected ? 0 : -1}
                        isRefreshing={
                          item?.isLoading() || (!node.isDir && selectedPath === node.path && isPreviewLoading)
                        }
                        onRefresh={() => void refreshPath(node.isDir ? node.path : getParentPath(node.path))}
                        onDownload={!node.isDir ? () => void onDownloadNode(node) : undefined}
                        onCopy={
                          !node.isDir && !getImageMimeType(node.path) ? () => void onCopyNodeContents(node) : undefined
                        }
                        onDelete={() => {
                          setOpenDropdownPath(null)
                          setDeleteTarget(node)
                        }}
                        onDropdownOpenChange={(open) => {
                          setOpenDropdownPath(open ? node.path : null)
                        }}
                        onStartCreateFolder={
                          node.isDir
                            ? () => {
                                setOpenDropdownPath(null)
                                setFolderCreationParentPath(node.path)
                                setNewFolderName('')
                              }
                            : undefined
                        }
                      />
                    ) : null}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
