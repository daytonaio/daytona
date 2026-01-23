/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Sandbox, SandboxState as SandboxStateEnum } from '@daytonaio/api-client'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { ChevronDown, ChevronRight, GitFork, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SandboxState } from './SandboxTable/SandboxState'
import { getRelativeTimeString } from '@/lib/utils'

// Request queue to limit concurrent API requests
class RequestQueue {
  private queue: Array<() => Promise<void>> = []
  private running = 0
  private maxConcurrent: number

  constructor(maxConcurrent = 3) {
    this.maxConcurrent = maxConcurrent
  }

  enqueue(request: () => Promise<void>): void {
    this.queue.push(request)
    this.processQueue()
  }

  private processQueue(): void {
    while (this.running < this.maxConcurrent && this.queue.length > 0) {
      this.running++
      const request = this.queue.shift()!
      request().finally(() => {
        this.running--
        this.processQueue()
      })
    }
  }

  clear(): void {
    this.queue = []
  }
}

// Tree node structure
interface TreeNode {
  sandbox: Sandbox
  children: TreeNode[]
  isExpanded: boolean
  isLoading: boolean
  childrenLoaded: boolean
}

interface ForkTreeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sandboxId: string | null
  onSelectSandbox?: (sandbox: Sandbox) => void
}

const MAX_DEPTH = 50
const MAX_CHILDREN_PER_LEVEL = 100

export function ForkTreeDialog({ open, onOpenChange, sandboxId, onSelectSandbox }: ForkTreeDialogProps) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const [rootNode, setRootNode] = useState<TreeNode | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loadingNodes, setLoadingNodes] = useState<Set<string>>(new Set())
  const queueRef = useRef(new RequestQueue(3))

  // Find a node in the tree by sandbox ID
  const findNode = useCallback((node: TreeNode | null, id: string): TreeNode | null => {
    if (!node) return null
    if (node.sandbox.id === id) return node
    for (const child of node.children) {
      const found = findNode(child, id)
      if (found) return found
    }
    return null
  }, [])

  // Update a node in the tree
  const updateNode = useCallback((nodeId: string, updater: (node: TreeNode) => TreeNode) => {
    setRootNode((prevRoot) => {
      if (!prevRoot) return null

      const updateRecursive = (node: TreeNode): TreeNode => {
        if (node.sandbox.id === nodeId) {
          return updater(node)
        }
        return {
          ...node,
          children: node.children.map(updateRecursive),
        }
      }

      return updateRecursive(prevRoot)
    })
  }, [])

  // Insert children into a node
  const insertChildren = useCallback(
    (parentId: string, children: Sandbox[]) => {
      updateNode(parentId, (node) => ({
        ...node,
        children: children.slice(0, MAX_CHILDREN_PER_LEVEL).map((sandbox) => ({
          sandbox,
          children: [],
          isExpanded: false,
          isLoading: false,
          childrenLoaded: false,
        })),
        childrenLoaded: true,
        isLoading: false,
      }))
    },
    [updateNode],
  )

  // Fetch children for a node
  const fetchChildren = useCallback(
    async (nodeId: string, depth: number) => {
      if (!selectedOrganization || depth >= MAX_DEPTH) return

      setLoadingNodes((prev) => new Set(prev).add(nodeId))
      updateNode(nodeId, (node) => ({ ...node, isLoading: true }))

      try {
        const response = await sandboxApi.getSandboxForks(nodeId, selectedOrganization.id)
        const children = response.data

        insertChildren(nodeId, children)

        // Queue fetching children for each child (progressive loading)
        children.slice(0, MAX_CHILDREN_PER_LEVEL).forEach((child: Sandbox) => {
          queueRef.current.enqueue(async () => {
            await fetchChildren(child.id, depth + 1)
          })
        })
      } catch (err) {
        console.error(`Failed to fetch children for ${nodeId}:`, err)
        updateNode(nodeId, (node) => ({ ...node, isLoading: false, childrenLoaded: true }))
      } finally {
        setLoadingNodes((prev) => {
          const next = new Set(prev)
          next.delete(nodeId)
          return next
        })
      }
    },
    [sandboxApi, selectedOrganization, insertChildren, updateNode],
  )

  // Build the initial tree from ancestors
  const buildTreeFromAncestors = useCallback((ancestors: Sandbox[], currentSandbox: Sandbox): TreeNode => {
    // Ancestors are ordered from direct parent to root, so we need to reverse
    const ancestorsReversed = [...ancestors].reverse()

    // Build tree from root down to current sandbox
    let currentNode: TreeNode | null = null

    for (const ancestor of ancestorsReversed) {
      const node: TreeNode = {
        sandbox: ancestor,
        children: currentNode ? [currentNode] : [],
        isExpanded: true, // Expand the path to the selected sandbox
        isLoading: false,
        childrenLoaded: false,
      }
      currentNode = node
    }

    // Add the current sandbox as a leaf (or the root if no ancestors)
    const currentSandboxNode: TreeNode = {
      sandbox: currentSandbox,
      children: [],
      isExpanded: true,
      isLoading: false,
      childrenLoaded: false,
    }

    if (currentNode) {
      // Find the deepest node (direct parent) and add current sandbox as child
      let deepest = currentNode
      while (deepest.children.length > 0) {
        deepest = deepest.children[0]
      }
      deepest.children = [currentSandboxNode]
      return currentNode
    }

    return currentSandboxNode
  }, [])

  // Load the initial tree
  useEffect(() => {
    if (!open || !sandboxId || !selectedOrganization) {
      return
    }

    const loadTree = async () => {
      setLoading(true)
      setError(null)
      setRootNode(null)
      queueRef.current.clear()

      try {
        // Get the current sandbox
        const sandboxResponse = await sandboxApi.getSandbox(sandboxId, selectedOrganization.id)
        const currentSandbox = sandboxResponse.data

        // Get ancestors (parent chain)
        // Note: When ancestors=true, the API returns Sandbox[], but the OpenAPI spec types it as Sandbox
        const ancestorsResponse = await sandboxApi.getSandboxParent(sandboxId, selectedOrganization.id, true)
        const ancestors = (ancestorsResponse.data as unknown as Sandbox[]) || []

        // Build initial tree
        const tree = buildTreeFromAncestors(ancestors, currentSandbox)
        setRootNode(tree)

        // Queue fetching children for the root and all ancestors
        const allNodeIds = [tree.sandbox.id, ...ancestors.map((a) => a.id)]
        allNodeIds.forEach((nodeId) => {
          queueRef.current.enqueue(async () => {
            await fetchChildren(nodeId, ancestors.length)
          })
        })
      } catch (err) {
        console.error('Failed to load fork tree:', err)
        setError('Failed to load fork tree')
      } finally {
        setLoading(false)
      }
    }

    loadTree()

    return () => {
      queueRef.current.clear()
    }
  }, [open, sandboxId, selectedOrganization, sandboxApi, buildTreeFromAncestors, fetchChildren])

  // Toggle node expansion
  const toggleExpand = useCallback(
    (nodeId: string) => {
      updateNode(nodeId, (node) => {
        const newExpanded = !node.isExpanded

        // If expanding and children not loaded, fetch them
        if (newExpanded && !node.childrenLoaded && !node.isLoading) {
          queueRef.current.enqueue(async () => {
            await fetchChildren(nodeId, 0)
          })
        }

        return { ...node, isExpanded: newExpanded }
      })
    },
    [updateNode, fetchChildren],
  )

  // Render a tree node
  const renderNode = useCallback(
    (node: TreeNode, depth: number, isLast: boolean, prefix: string): React.ReactNode => {
      const isSelected = node.sandbox.id === sandboxId
      const hasChildren = node.children.length > 0 || (!node.childrenLoaded && !node.isLoading)
      const showExpandIcon = hasChildren || node.isLoading

      // Build the tree line prefix
      const linePrefix = depth === 0 ? '' : prefix + (isLast ? '└── ' : '├── ')
      const childPrefix = depth === 0 ? '' : prefix + (isLast ? '    ' : '│   ')

      return (
        <div key={node.sandbox.id}>
          <div
            className={cn(
              'flex items-center gap-2 py-1.5 px-2 rounded-md cursor-pointer hover:bg-muted/50 transition-colors',
              isSelected && 'bg-primary/10 border border-primary/30',
            )}
            onClick={() => onSelectSandbox?.(node.sandbox)}
          >
            {/* Tree structure prefix */}
            <span className="text-muted-foreground font-mono text-xs whitespace-pre select-none">{linePrefix}</span>

            {/* Expand/collapse button */}
            {showExpandIcon && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  toggleExpand(node.sandbox.id)
                }}
                className="p-0.5 hover:bg-muted rounded"
              >
                {node.isLoading ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin text-muted-foreground" />
                ) : node.isExpanded ? (
                  <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
                ) : (
                  <ChevronRight className="w-3.5 h-3.5 text-muted-foreground" />
                )}
              </button>
            )}

            {!showExpandIcon && <div className="w-4" />}

            {/* Sandbox info */}
            <GitFork className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            <span className={cn('font-medium truncate', isSelected && 'text-primary')}>{node.sandbox.name}</span>

            {/* State badge */}
            <div className="flex-shrink-0">
              <SandboxState state={node.sandbox.state} />
            </div>

            {/* Created time */}
            {node.sandbox.createdAt && (
              <span className="text-xs text-muted-foreground ml-auto flex-shrink-0">
                {getRelativeTimeString(node.sandbox.createdAt).relativeTimeString}
              </span>
            )}

            {/* Selected indicator */}
            {isSelected && (
              <span className="text-xs bg-primary text-primary-foreground px-1.5 py-0.5 rounded flex-shrink-0">
                current
              </span>
            )}
          </div>

          {/* Render children */}
          {node.isExpanded && node.children.length > 0 && (
            <div className="ml-0">
              {node.children.map((child, index) =>
                renderNode(child, depth + 1, index === node.children.length - 1, childPrefix),
              )}
            </div>
          )}

          {/* Loading indicator for children */}
          {node.isExpanded && node.isLoading && node.children.length === 0 && (
            <div className="flex items-center gap-2 py-1.5 px-2 ml-6">
              <span className="text-muted-foreground font-mono text-xs whitespace-pre select-none">
                {childPrefix}└──
              </span>
              <Loader2 className="w-3.5 h-3.5 animate-spin text-muted-foreground" />
              <span className="text-muted-foreground text-sm">Loading...</span>
            </div>
          )}
        </div>
      )
    },
    [sandboxId, toggleExpand, onSelectSandbox],
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <GitFork className="w-5 h-5" />
            Fork Tree
          </DialogTitle>
          <DialogDescription>View the fork hierarchy for this sandbox and its related sandboxes.</DialogDescription>
        </DialogHeader>

        <div className="overflow-auto max-h-[60vh] border rounded-md p-2 bg-muted/20">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
              <span className="ml-2 text-muted-foreground">Loading fork tree...</span>
            </div>
          )}

          {error && (
            <div className="flex items-center justify-center py-8 text-destructive">
              <span>{error}</span>
            </div>
          )}

          {!loading && !error && rootNode && renderNode(rootNode, 0, true, '')}

          {!loading && !error && !rootNode && (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <span>No fork tree data available</span>
            </div>
          )}
        </div>

        {loadingNodes.size > 0 && (
          <div className="text-xs text-muted-foreground flex items-center gap-1">
            <Loader2 className="w-3 h-3 animate-spin" />
            Loading {loadingNodes.size} node{loadingNodes.size > 1 ? 's' : ''}...
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
