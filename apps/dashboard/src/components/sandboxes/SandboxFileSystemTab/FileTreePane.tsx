/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Buffer } from 'buffer'
import {
  memo,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
  type FocusEvent,
  type KeyboardEvent,
} from 'react'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import TooltipButton from '@/components/TooltipButton'
import { cn } from '@/lib/utils'
import {
  asyncDataLoaderFeature,
  dragAndDropFeature,
  expandAllFeature,
  hotkeysCoreFeature,
  keyboardDragAndDropFeature,
} from '@headless-tree/core'
import { useTree } from '@headless-tree/react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ChevronsUpDownIcon, EllipsisIcon, RefreshCwIcon } from 'lucide-react'

import { toast } from 'sonner'
import { FILE_SEARCH_MIN_CHARS, ROOT_NODE, ROOT_PATH } from './constants'
import { FileSearchHeader, type FileSearchHeaderHandle } from './FileSearchHeader'
import { FileTreeRow } from './FileTreeRow'
import { useFileSystemStore } from './fileSystemStore'
import { useMoveNodeMutation } from './mutations'
import {
  useFileDetailsCache,
  useFileDetailsQueries,
  useFetchDirectoryChildrenQuery,
  useFetchFileDetailsQuery,
  useFileSearchQuery,
  useIsDirectoryRefreshing,
  useInvalidateDirectoryQuery,
} from './queries'
import type { SandboxFileSystemNode } from './types'
import { useSandboxInstance } from './useSandboxInstance'
import {
  createFallbackNode,
  downloadSandboxFile,
  getAncestorPaths,
  getCanvasFont,
  getImageMimeType,
  isSameOrDescendantPath,
  joinSandboxPath,
  getParentPath,
  handleFileSystemApiError,
  isProbablyBinary,
} from './utils'

export type FileTreePaneRef = {
  expandPathAncestors: (path: string) => Promise<void>
  getNode: (path: string) => SandboxFileSystemNode
  refreshPath: (path: string) => Promise<void>
  revealPath: (path: string) => void
  restoreFocus: (path: string) => void
}

const MemoizedFileSearchHeader = memo(FileSearchHeader)
const FILE_TREE_EDGE_PADDING = 8
const FILE_SEARCH_RESULT_LABEL_RESERVED_WIDTH = 40

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
  onRequestCreateFolder,
  onRequestDelete,
  ref,
  previewLoadingPath,
  sandboxId,
}: {
  onRequestCreateFolder: (parentPath: string) => void
  onRequestDelete: (node: SandboxFileSystemNode) => void
  ref?: React.Ref<FileTreePaneRef>
  previewLoadingPath?: string
  sandboxId: string
}) {
  const [rootLoadFailed, setRootLoadFailed] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [isTreeFocusWithin, setIsTreeFocusWithin] = useState(false)
  const didSetInitialFocusRef = useRef(false)
  const fileTreeScrollAreaRef = useRef<HTMLDivElement>(null)
  const fileTreeViewportRef = useRef<HTMLDivElement>(null)
  const searchHeaderRef = useRef<FileSearchHeaderHandle>(null)
  const sandboxInstanceQuery = useSandboxInstance(sandboxId)
  const sandboxInstance = sandboxInstanceQuery.data
  const moveNodeMutation = useMoveNodeMutation({ sandboxInstance })
  const fetchDirectoryChildren = useFetchDirectoryChildrenQuery({ sandboxInstance })
  const fetchFileDetails = useFetchFileDetailsQuery({ sandboxInstance })
  const { getCachedNode } = useFileDetailsCache({ sandboxInstance })
  const invalidateDirectory = useInvalidateDirectoryQuery({ sandboxInstance })

  const selectedNodePath = useFileSystemStore((state) => state.selectedNodePath)
  const [searchLabelAvailableWidth, setSearchLabelAvailableWidth] = useState(0)
  const [searchLabelFont, setSearchLabelFont] = useState('')
  const { openNode, setAdjacentFilePaths } = useFileSystemStore((state) => state.actions)

  const resetSearch = useCallback(() => {
    searchHeaderRef.current?.clear()
    setSearchQuery('')
  }, [])

  async function loadDirectory(path: string) {
    if (!sandboxInstance) {
      return []
    }

    try {
      const children = await fetchDirectoryChildren(path)

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
      const existingNode = getCachedNode(path)
      if (existingNode && !isFallbackNode(existingNode)) {
        return existingNode
      }

      if (!sandboxInstance) {
        return existingNode ?? createFallbackNode(path)
      }

      try {
        return await fetchFileDetails(path)
      } catch {
        return existingNode ?? createFallbackNode(path)
      }
    },
    [fetchFileDetails, getCachedNode, sandboxInstance],
  )

  const searchEnabled = searchQuery.trim().length >= FILE_SEARCH_MIN_CHARS
  const searchQueryResult = useFileSearchQuery({
    enabled: Boolean(sandboxInstance) && searchEnabled,
    query: searchQuery,
    sandboxInstance,
  })
  const searchResultPaths = searchQueryResult.data ?? defaultSearchResults
  const searchResultNodesByPath = useFileDetailsQueries({
    paths: searchEnabled ? searchResultPaths : defaultSearchResults,
    sandboxInstance,
  })
  const searchResults = useMemo(() => {
    return searchResultPaths.map((path) => searchResultNodesByPath.get(path) ?? createFallbackNode(path))
  }, [searchResultNodesByPath, searchResultPaths])
  const isSearchLoading = searchQueryResult.isFetching
  const searchFailed = searchQueryResult.isError

  const tree = useTree<SandboxFileSystemNode | null>({
    rootItemId: ROOT_PATH,
    initialState: { expandedItems: [ROOT_PATH] },
    canDrop: (items, target) => {
      const targetPath = target.item.getId()

      return (
        target.item.isFolder() &&
        items.every((item) => {
          const sourcePath = item.getId()
          return sourcePath !== targetPath && !isSameOrDescendantPath(targetPath, sourcePath)
        })
      )
    },
    canReorder: false,
    createLoadingItemData: () => null,
    getItemName: (item) => {
      const data = item.getItemData()
      return getCachedNode(item.getId())?.name ?? data?.name ?? item.getId()
    },
    isItemFolder: (item) => Boolean(getCachedNode(item.getId())?.isDir ?? item.getItemData()?.isDir),
    onDrop: async (items, target) => {
      const destinationDirectoryPath = target.item.getId()

      const moveResults = await Promise.allSettled(
        items.map(async (item) => {
          const node = getCachedNode(item.getId()) ?? item.getItemData() ?? createFallbackNode(item.getId())
          const destinationPath = joinSandboxPath(
            destinationDirectoryPath,
            node.name || item.getId().split('/').at(-1) || item.getId(),
          )

          if (destinationPath === node.path) {
            return null
          }

          await moveNodeMutation.mutateAsync({
            destinationPath,
            node,
          })

          return {
            destinationPath,
            node,
          }
        }),
      )

      const completedMoves = moveResults.flatMap((result) =>
        result.status === 'fulfilled' && result.value ? [result.value] : [],
      )
      const failedMoves = moveResults.flatMap((result) => (result.status === 'rejected' ? [result.reason] : []))

      if (completedMoves.length > 0) {
        if (target.item.isFolder() && !target.item.isExpanded()) {
          target.item.expand()
        }

        const sourceParentPaths = new Set(completedMoves.map(({ node }) => getParentPath(node.path)))
        const pathsToRefresh = new Set([destinationDirectoryPath, ...sourceParentPaths])

        await Promise.all(
          Array.from(pathsToRefresh).map(async (path) => {
            await refreshPath(path)
          }),
        )

        for (const { destinationPath } of completedMoves) {
          await resolveNode(destinationPath)
        }

        const remappedSelection = completedMoves.find(
          ({ node }) => selectedPath && isSameOrDescendantPath(selectedPath, node.path),
        )

        if (remappedSelection && selectedPath) {
          const remappedSelectedPath =
            selectedPath === remappedSelection.node.path
              ? remappedSelection.destinationPath
              : `${remappedSelection.destinationPath}${selectedPath.slice(remappedSelection.node.path.length)}`
          openNode(remappedSelectedPath)
          restoreFocus(remappedSelectedPath)
        } else {
          const destinationItem = tree.getItemInstance(destinationDirectoryPath)
          destinationItem?.setFocused()
          tree.updateDomFocus()
        }
      }

      if (failedMoves.length > 0) {
        handleFileSystemApiError(failedMoves[0], 'Failed to move item')
      }
    },
    onPrimaryAction: (item) => {
      const node = getCachedNode(item.getId()) ?? item.getItemData() ?? createFallbackNode(item.getId())
      openNode(node.path)
    },
    seperateDragHandle: true,
    dataLoader: {
      getItem: async (itemId) => getCachedNode(itemId) ?? createFallbackNode(itemId),
      getChildrenWithData: async (itemId) => loadDirectory(itemId),
    },
    features: [
      asyncDataLoaderFeature,
      expandAllFeature,
      hotkeysCoreFeature,
      dragAndDropFeature,
      keyboardDragAndDropFeature,
    ],
  })

  const visibleItems = tree.getItems()
  const selectedPath = selectedNodePath ?? ''
  const visibleItemPaths = useMemo(() => visibleItems.map((item) => item.getId()), [visibleItems])
  const observedDetailPaths = useMemo(
    () => Array.from(new Set([...visibleItemPaths, ...(selectedPath ? [selectedPath] : [])])),
    [selectedPath, visibleItemPaths],
  )
  const observedNodesByPath = useFileDetailsQueries({
    paths: observedDetailPaths,
    sandboxInstance,
  })
  const selectedNode = selectedPath
    ? (observedNodesByPath.get(selectedPath) ??
      getCachedNode(selectedPath) ??
      (selectedPath === ROOT_PATH ? ROOT_NODE : createFallbackNode(selectedPath)))
    : null
  const rootItem = tree.getItemInstance(ROOT_PATH)
  const isRootDirectoryRefreshing = useIsDirectoryRefreshing({
    path: ROOT_PATH,
    sandboxInstance,
  })
  const isInitialTreeLoading = (!sandboxInstance || rootItem.isLoading()) && visibleItems.length <= 1 && !rootLoadFailed
  const isTreeRefreshing = !isInitialTreeLoading && (rootItem.isLoading() || isRootDirectoryRefreshing)
  const visibleFileItems = useMemo(() => {
    return visibleItems.filter((item) => {
      const node = observedNodesByPath.get(item.getId()) ?? item.getItemData() ?? createFallbackNode(item.getId())
      return !node.isDir
    })
  }, [observedNodesByPath, visibleItems])

  const activeFileNodes = useMemo(() => {
    if (searchEnabled) {
      return searchResults.filter((node) => !node.isDir)
    }

    return visibleFileItems.map(
      (item) => observedNodesByPath.get(item.getId()) ?? item.getItemData() ?? createFallbackNode(item.getId()),
    )
  }, [observedNodesByPath, searchEnabled, searchResults, visibleFileItems])

  const selectedFileIndex = useMemo(() => {
    if (!selectedNode || selectedNode.isDir) {
      return -1
    }

    return activeFileNodes.findIndex((node) => node.path === selectedPath)
  }, [activeFileNodes, selectedNode, selectedPath])

  useEffect(() => {
    setAdjacentFilePaths({
      previousFilePath: selectedFileIndex > 0 ? (activeFileNodes[selectedFileIndex - 1]?.path ?? null) : null,
      nextFilePath:
        selectedFileIndex >= 0 && selectedFileIndex < activeFileNodes.length - 1
          ? (activeFileNodes[selectedFileIndex + 1]?.path ?? null)
          : null,
    })
  }, [activeFileNodes, selectedFileIndex, setAdjacentFilePaths])

  const fileTreeVirtualizer = useVirtualizer({
    count: searchEnabled ? searchResults.length : visibleItems.length,
    getScrollElement: () => fileTreeViewportRef.current,
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
    const viewport = fileTreeViewportRef.current
    if (!viewport) {
      return
    }

    const updateMeasurements = () => {
      setSearchLabelAvailableWidth(
        Math.max(0, viewport.getBoundingClientRect().width - FILE_SEARCH_RESULT_LABEL_RESERVED_WIDTH),
      )
      setSearchLabelFont(getCanvasFont(viewport))
    }

    updateMeasurements()

    const resizeObserver = new ResizeObserver(() => {
      updateMeasurements()
    })

    resizeObserver.observe(viewport)
    return () => resizeObserver.disconnect()
  }, [])

  useEffect(() => {
    if (didSetInitialFocusRef.current) {
      return
    }

    didSetInitialFocusRef.current = true
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

      await invalidateDirectory(path)

      if (path === ROOT_PATH) {
        await rootItem.invalidateChildrenIds(true)
        return
      }

      await tree.getItemInstance(path).invalidateChildrenIds(true)
    },
    [invalidateDirectory, rootItem, sandboxInstance, tree],
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

  const revealPath = useCallback(
    (path: string) => {
      if (searchEnabled) {
        const itemIndex = activeFileNodes.findIndex((node) => node.path === path)
        if (itemIndex >= 0) {
          fileTreeVirtualizer.scrollToIndex(itemIndex, { align: 'auto' })
        }
        return
      }

      const item = visibleFileItems.find((visibleItem) => visibleItem.getId() === path)
      item?.setFocused()
      tree.updateDomFocus()
    },
    [activeFileNodes, fileTreeVirtualizer, searchEnabled, tree, visibleFileItems],
  )

  const getNode = useCallback((path: string) => getCachedNode(path) ?? createFallbackNode(path), [getCachedNode])

  const handleRefreshRoot = useCallback(async () => {
    await refreshPath(ROOT_PATH)
  }, [refreshPath])

  const handleCollapseAll = useCallback(() => {
    tree.collapseAll()
    rootItem.expand()
  }, [rootItem, tree])

  const handleTreeContainerFocus = useCallback(
    (event: FocusEvent<HTMLDivElement>) => {
      if (event.target !== event.currentTarget) {
        return
      }

      if (searchEnabled) {
        const firstSearchResultButton =
          fileTreeScrollAreaRef.current?.querySelector<HTMLButtonElement>('[data-file-tree-row-button]')
        firstSearchResultButton?.focus()
        return
      }

      const focusedItem = tree.getItems().find((item) => item.isFocused()) ?? tree.getItems()[0]
      focusedItem?.setFocused()
      tree.updateDomFocus()
    },
    [searchEnabled, tree],
  )

  const handleTreeFocusCapture = useCallback(() => {
    setIsTreeFocusWithin(true)
  }, [])

  const handleTreeBlurCapture = useCallback((event: FocusEvent<HTMLDivElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
      setIsTreeFocusWithin(false)
    }
  }, [])

  useImperativeHandle(
    ref,
    () => ({
      expandPathAncestors,
      getNode,
      refreshPath,
      revealPath,
      restoreFocus,
    }),
    [expandPathAncestors, getNode, refreshPath, restoreFocus, revealPath],
  )

  const handleItemKeyDown = useCallback(
    (item: ReturnType<typeof tree.getItemInstance>, event: KeyboardEvent<HTMLDivElement>) => {
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

  async function handleRetryRoot() {
    await refreshPath(ROOT_PATH)
  }

  async function handleDownloadNode(node: SandboxFileSystemNode) {
    if (!sandboxInstance || node.isDir) {
      return
    }

    await downloadSandboxFile({
      node,
      sandboxInstance,
    })
  }

  async function handleCopyNode(node: SandboxFileSystemNode) {
    if (!sandboxInstance || node.isDir) {
      return
    }

    try {
      const fileContents = Buffer.from(await sandboxInstance.fs.downloadFile(node.path))
      if (isProbablyBinary(fileContents)) {
        toast.error('Binary file contents cannot be copied as text')
        return
      }

      await navigator.clipboard.writeText(fileContents.toString('utf-8'))
      toast.success(`Copied contents of ${node.name || node.path}`)
    } catch (error) {
      handleFileSystemApiError(error, `Failed to copy ${node.path}`)
    }
  }

  if (rootLoadFailed) {
    return (
      <div className="flex flex-1 min-h-0">
        <Empty className="border-0">
          <EmptyHeader>
            <EmptyTitle>Failed to load filesystem</EmptyTitle>
            <EmptyDescription>Something went wrong while listing the sandbox root directory.</EmptyDescription>
          </EmptyHeader>
          <Button variant="outline" size="sm" onClick={handleRetryRoot}>
            <RefreshCwIcon className="size-4" />
            Retry
          </Button>
        </Empty>
      </div>
    )
  }

  return (
    <div ref={fileTreeScrollAreaRef} className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <MemoizedFileSearchHeader
        actions={
          <>
            <TooltipButton
              tooltipText="Refresh files"
              variant="ghost"
              size="icon-sm"
              onClick={handleRefreshRoot}
              disabled={isTreeRefreshing}
            >
              <RefreshCwIcon
                className={cn('size-4', {
                  'animate-spin': isTreeRefreshing,
                })}
              />
            </TooltipButton>
            <TooltipButton
              tooltipText="Collapse all folders"
              variant="ghost"
              size="icon-sm"
              onClick={handleCollapseAll}
            >
              <ChevronsUpDownIcon className="size-4" />
            </TooltipButton>
          </>
        }
        onSearchQueryChange={setSearchQuery}
        ref={searchHeaderRef}
      />
      <div role="status" aria-live="polite" className="sr-only">
        {searchEnabled && !isSearchLoading && !searchFailed
          ? searchResults.length === 0
            ? 'No files found'
            : `${searchResults.length} files found`
          : null}
      </div>

      <div
        className="min-h-0 flex-1 focus-visible:outline-none"
        tabIndex={isTreeFocusWithin ? -1 : 0}
        onBlurCapture={handleTreeBlurCapture}
        onFocus={handleTreeContainerFocus}
        onFocusCapture={handleTreeFocusCapture}
      >
        <ScrollArea fade="mask" className="min-h-0 h-full" viewportRef={fileTreeViewportRef}>
          {isInitialTreeLoading || (searchEnabled && isSearchLoading && searchResults.length === 0) ? (
            <FileTreeSkeleton />
          ) : searchEnabled && searchResults.length === 0 && !isSearchLoading && !searchFailed ? (
            <Empty className="border-0">
              <EmptyHeader>
                <EmptyTitle>0 results</EmptyTitle>
                <EmptyDescription>
                  No files matched "{searchQuery}". Try another search or clear search.
                </EmptyDescription>
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
                  : (observedNodesByPath.get(item?.getId() ?? '') ?? item?.getItemData() ?? ROOT_NODE)

                if (!node) {
                  return null
                }

                const isDirectory = searchEnabled ? node.isDir : (item?.isFolder() ?? node.isDir)
                const isSelected = selectedPath === node.path
                const isPreviewLoading = previewLoadingPath === node.path
                const isFocused = searchEnabled ? false : (item?.isFocused() ?? false)
                const isLoading = (item?.isLoading() ?? false) || isPreviewLoading
                const itemProps = searchEnabled ? { tabIndex: -1 } : (item?.getProps() ?? {})
                return (
                  <FileTreeRow
                    key={searchEnabled ? node.path : item?.getId()}
                    actions={
                      !searchEnabled ? (
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            asChild
                            onClick={(event) => event.stopPropagation()}
                            onPointerDown={(event) => event.stopPropagation()}
                          >
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              tabIndex={isFocused || isSelected ? 0 : -1}
                              className="inline-flex h-8 w-8 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted hover:text-foreground"
                            >
                              <EllipsisIcon className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent
                            align="end"
                            side="right"
                            className={cn({
                              'w-44': node.isDir,
                              'w-48': !node.isDir,
                            })}
                            onCloseAutoFocus={(event) => event.preventDefault()}
                          >
                            <DropdownMenuItem
                              onClick={() => refreshPath(node.isDir ? node.path : getParentPath(node.path))}
                              disabled={
                                item?.isLoading() || (!node.isDir && selectedPath === node.path && isPreviewLoading)
                              }
                            >
                              Refresh
                            </DropdownMenuItem>
                            {!node.isDir ? (
                              <DropdownMenuItem onClick={() => handleDownloadNode(node)}>Download</DropdownMenuItem>
                            ) : null}
                            {!node.isDir && !getImageMimeType(node.path) ? (
                              <DropdownMenuItem onClick={() => handleCopyNode(node)}>Copy contents</DropdownMenuItem>
                            ) : null}
                            {node.isDir ? (
                              <DropdownMenuItem onSelect={() => onRequestCreateFolder(node.path)}>
                                Create folder
                              </DropdownMenuItem>
                            ) : null}
                            {node.path !== ROOT_PATH ? <DropdownMenuSeparator /> : null}
                            {node.path !== ROOT_PATH ? (
                              <DropdownMenuItem variant="destructive" onClick={() => onRequestDelete(node)}>
                                Delete
                              </DropdownMenuItem>
                            ) : null}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      ) : undefined
                    }
                    itemProps={itemProps}
                    depth={item?.getItemMeta().level ?? 0}
                    dragHandleProps={!searchEnabled ? (item?.getDragHandleProps() ?? undefined) : undefined}
                    isExpanded={item?.isExpanded() ?? false}
                    isDragTarget={item?.isDragTarget() ?? false}
                    isDragTargetAbove={item?.isDragTargetAbove() ?? false}
                    isDragTargetBelow={item?.isDragTargetBelow() ?? false}
                    isDraggingOver={item?.isDraggingOver() ?? false}
                    isFocused={isFocused}
                    isLoading={isLoading}
                    isSearchResult={searchEnabled}
                    isSelected={isSelected}
                    node={node}
                    onActivate={async (event) => {
                      itemProps.onClick?.(event)

                      if (event.defaultPrevented) {
                        return
                      }

                      if (!searchEnabled && item && node.isDir && !item.isExpanded()) {
                        item.expand()
                      }

                      const resolvedNode = searchEnabled ? await resolveNode(node.path) : node
                      openNode(resolvedNode.path)
                    }}
                    onItemKeyDown={!searchEnabled && item ? (event) => handleItemKeyDown(item, event) : undefined}
                    onToggleExpand={
                      !searchEnabled && isDirectory && item
                        ? () => {
                            if (item.isExpanded()) {
                              item.collapse()
                            } else {
                              item.expand()
                            }
                          }
                        : undefined
                    }
                    searchLabel={
                      searchEnabled
                        ? {
                            availableWidth: searchLabelAvailableWidth,
                            font: searchLabelFont,
                            query: searchQuery,
                          }
                        : undefined
                    }
                    top={virtualItem.start + FILE_TREE_EDGE_PADDING}
                  />
                )
              })}
              {!searchEnabled ? (
                <div className="pointer-events-none border-t-2 border-primary" style={tree.getDragLineStyle()} />
              ) : null}
            </div>
          )}
        </ScrollArea>
      </div>
    </div>
  )
}
