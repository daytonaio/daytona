/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Sandbox, SandboxState as SandboxStateEnum } from '@daytonaio/api-client'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { AlertTriangle, ChevronDown, ChevronRight, GitFork, Loader2, Trash2, CheckCircle2, XCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SandboxState } from './SandboxTable/SandboxState'
import { getRelativeTimeString } from '@/lib/utils'
import { buttonVariants } from '@/components/ui/button'
import { handleApiError } from '@/lib/error-handling'

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
  deleteStatus: 'pending' | 'deleting' | 'deleted' | 'error'
}

interface RecursiveDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sandboxId: string | null
  onComplete: () => void
}

const MAX_DEPTH = 50
const MAX_CHILDREN_PER_LEVEL = 100

export function RecursiveDeleteDialog({ open, onOpenChange, sandboxId, onComplete }: RecursiveDeleteDialogProps) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const [rootNode, setRootNode] = useState<TreeNode | null>(null)
  const [loading, setLoading] = useState(false)
  const [loadingComplete, setLoadingComplete] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteProgress, setDeleteProgress] = useState({ current: 0, total: 0, currentName: '', status: '' })
  const [deleteComplete, setDeleteComplete] = useState(false)
  const [deleteFailed, setDeleteFailed] = useState(false)
  const abortRef = useRef(false)

  // Update a node in the tree (only used during deletion for status updates)
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

  // Recursively fetch and build tree node with all descendants
  const buildTreeNode = useCallback(
    async (sandbox: Sandbox, depth: number): Promise<TreeNode> => {
      const node: TreeNode = {
        sandbox,
        children: [],
        isExpanded: true,
        isLoading: false,
        childrenLoaded: true,
        deleteStatus: 'pending',
      }

      if (depth >= MAX_DEPTH) {
        return node
      }

      try {
        const response = await sandboxApi.getSandboxForks(sandbox.id, selectedOrganization!.id)
        const children = response.data.slice(0, MAX_CHILDREN_PER_LEVEL)

        // Recursively build child nodes
        const childNodes = await Promise.all(children.map((child: Sandbox) => buildTreeNode(child, depth + 1)))
        node.children = childNodes
      } catch (err) {
        console.error(`Failed to fetch children for ${sandbox.id}:`, err)
      }

      return node
    },
    [sandboxApi, selectedOrganization],
  )

  // Load the tree
  useEffect(() => {
    if (!open || !sandboxId || !selectedOrganization) {
      return
    }

    const loadTree = async () => {
      setLoading(true)
      setLoadingComplete(false)
      setError(null)
      setRootNode(null)
      setIsDeleting(false)
      setDeleteComplete(false)
      setDeleteFailed(false)
      setDeleteProgress({ current: 0, total: 0, currentName: '', status: '' })
      abortRef.current = false

      try {
        // Get the current sandbox
        const sandboxResponse = await sandboxApi.getSandbox(sandboxId, selectedOrganization.id)
        const currentSandbox = sandboxResponse.data

        // Build the entire tree recursively
        const tree = await buildTreeNode(currentSandbox, 0)

        setRootNode(tree)
        setLoadingComplete(true)
      } catch (err) {
        console.error('Failed to load sandbox tree:', err)
        setError('Failed to load sandbox tree')
      } finally {
        setLoading(false)
      }
    }

    loadTree()

    return () => {
      abortRef.current = true
    }
  }, [open, sandboxId, selectedOrganization, sandboxApi, buildTreeNode])

  // Count total sandboxes in tree
  const countSandboxes = useCallback((node: TreeNode | null): number => {
    if (!node) return 0
    return 1 + node.children.reduce((sum, child) => sum + countSandboxes(child), 0)
  }, [])

  // Flatten tree to list using post-order traversal (children before parents)
  const flattenTreePostOrder = useCallback((node: TreeNode | null): TreeNode[] => {
    if (!node) return []
    const result: TreeNode[] = []
    for (const child of node.children) {
      result.push(...flattenTreePostOrder(child))
    }
    result.push(node)
    return result
  }, [])

  // Wait for a sandbox to reach DESTROYED state
  const waitForDestroyed = useCallback(
    async (sandboxId: string, maxAttempts = 30, delayMs = 1000): Promise<boolean> => {
      for (let attempt = 0; attempt < maxAttempts; attempt++) {
        if (abortRef.current) return false

        try {
          const response = await sandboxApi.getSandbox(sandboxId, selectedOrganization!.id)
          const state = response.data.state

          if (state === SandboxStateEnum.DESTROYED) {
            return true
          }

          // If not destroyed yet, wait and retry
          await new Promise((resolve) => setTimeout(resolve, delayMs))
        } catch (err: any) {
          // 404 means the sandbox is deleted
          if (err?.message?.includes('not found') || err?.message?.includes('404')) {
            return true
          }
          // Other errors - continue waiting
          await new Promise((resolve) => setTimeout(resolve, delayMs))
        }
      }
      return false
    },
    [sandboxApi, selectedOrganization],
  )

  // Perform recursive deletion
  const performRecursiveDelete = useCallback(async () => {
    if (!rootNode || !selectedOrganization) return

    setIsDeleting(true)
    abortRef.current = false

    const nodesToDelete = flattenTreePostOrder(rootNode)
    const total = nodesToDelete.length

    setDeleteProgress({ current: 0, total, currentName: '', status: '' })

    for (let i = 0; i < nodesToDelete.length; i++) {
      if (abortRef.current) {
        setDeleteFailed(true)
        setIsDeleting(false)
        return
      }

      const node = nodesToDelete[i]
      const sandboxName = node.sandbox.name || node.sandbox.id

      setDeleteProgress({ current: i, total, currentName: sandboxName, status: 'Initiating deletion...' })
      updateNode(node.sandbox.id, (n) => ({ ...n, deleteStatus: 'deleting' }))

      try {
        // Initiate deletion
        await sandboxApi.deleteSandbox(node.sandbox.id, selectedOrganization.id)

        // Wait for the sandbox to be fully destroyed before proceeding
        // This is important because parent sandboxes cannot be deleted until
        // all their children are fully destroyed (not just DESTROYING)
        setDeleteProgress({ current: i, total, currentName: sandboxName, status: 'Waiting for destruction...' })
        const destroyed = await waitForDestroyed(node.sandbox.id)

        if (!destroyed && !abortRef.current) {
          console.warn(`Sandbox ${node.sandbox.id} did not reach DESTROYED state in time, proceeding anyway`)
        }

        updateNode(node.sandbox.id, (n) => ({ ...n, deleteStatus: 'deleted' }))
      } catch (err) {
        console.error(`Failed to delete sandbox ${node.sandbox.id}:`, err)
        updateNode(node.sandbox.id, (n) => ({ ...n, deleteStatus: 'error' }))
        handleApiError(err, `Failed to delete sandbox ${sandboxName}`)
        setDeleteFailed(true)
        setIsDeleting(false)
        return
      }
    }

    setDeleteProgress({ current: total, total, currentName: '', status: 'Complete' })
    setDeleteComplete(true)
    setIsDeleting(false)
  }, [rootNode, selectedOrganization, sandboxApi, flattenTreePostOrder, updateNode, waitForDestroyed])

  // Handle dialog close - prevent closing during deletion
  const handleClose = useCallback(
    (isOpen: boolean) => {
      // Prevent closing while deletion is in progress
      if (isDeleting && !isOpen) {
        return
      }

      if (!isOpen) {
        abortRef.current = true
        if (deleteComplete) {
          onComplete()
        }
        onOpenChange(false)
      }
    },
    [isDeleting, deleteComplete, onComplete, onOpenChange],
  )

  // Force close the dialog (for the Close button after completion)
  const forceClose = useCallback(() => {
    onComplete()
    onOpenChange(false)
  }, [onComplete, onOpenChange])

  // Toggle node expansion
  const toggleExpand = useCallback(
    (nodeId: string) => {
      updateNode(nodeId, (node) => ({ ...node, isExpanded: !node.isExpanded }))
    },
    [updateNode],
  )

  // Render a tree node
  const renderNode = useCallback(
    (node: TreeNode, depth: number, isLast: boolean, prefix: string): React.ReactNode => {
      const isRoot = node.sandbox.id === sandboxId
      const hasChildren = node.children.length > 0
      const showExpandIcon = hasChildren || node.isLoading

      // Build the tree line prefix
      const linePrefix = depth === 0 ? '' : prefix + (isLast ? '└── ' : '├── ')
      const childPrefix = depth === 0 ? '' : prefix + (isLast ? '    ' : '│   ')

      const getStatusIcon = () => {
        switch (node.deleteStatus) {
          case 'deleting':
            return <Loader2 className="w-4 h-4 animate-spin text-yellow-500" />
          case 'deleted':
            return <CheckCircle2 className="w-4 h-4 text-green-500" />
          case 'error':
            return <XCircle className="w-4 h-4 text-red-500" />
          default:
            return <Trash2 className="w-4 h-4 text-muted-foreground" />
        }
      }

      return (
        <div key={node.sandbox.id}>
          <div
            className={cn(
              'flex items-center gap-2 py-1.5 px-2 rounded-md transition-colors',
              isRoot && 'bg-destructive/10 border border-destructive/30',
              node.deleteStatus === 'deleting' && 'bg-yellow-500/10',
              node.deleteStatus === 'deleted' && 'bg-green-500/10 opacity-50',
              node.deleteStatus === 'error' && 'bg-red-500/10',
            )}
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
                disabled={isDeleting}
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

            {/* Status icon */}
            {getStatusIcon()}

            {/* Sandbox info */}
            <GitFork className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            <span className={cn('font-medium truncate', isRoot && 'text-destructive')}>{node.sandbox.name}</span>

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

            {/* Root indicator */}
            {isRoot && (
              <span className="text-xs bg-destructive text-destructive-foreground px-1.5 py-0.5 rounded flex-shrink-0">
                target
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
    [sandboxId, toggleExpand, isDeleting],
  )

  const totalSandboxes = countSandboxes(rootNode)
  const progressPercent = deleteProgress.total > 0 ? (deleteProgress.current / deleteProgress.total) * 100 : 0

  return (
    <AlertDialog open={open} onOpenChange={handleClose}>
      <AlertDialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="w-5 h-5" />
            Delete Sandbox and All Descendants
          </AlertDialogTitle>
          <AlertDialogDescription>
            {deleteComplete ? (
              <span className="text-green-600">All sandboxes have been deleted successfully.</span>
            ) : deleteFailed ? (
              <span className="text-red-600">
                Deletion failed. Some sandboxes may have been deleted. Please try again.
              </span>
            ) : isDeleting ? (
              <span>Deleting sandboxes... Please do not close this dialog.</span>
            ) : (
              <>
                This sandbox has forked children that depend on it. To delete this sandbox, all descendant sandboxes
                must be deleted first.
                <br />
                <br />
                {!loadingComplete ? (
                  <span className="text-muted-foreground">Loading fork tree... Please wait.</span>
                ) : (
                  <>
                    <strong className="text-destructive">
                      {totalSandboxes} sandbox{totalSandboxes !== 1 ? 'es' : ''} will be permanently deleted.
                    </strong>{' '}
                    This action cannot be undone.
                  </>
                )}
              </>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>

        {/* Progress bar during deletion */}
        {(isDeleting || deleteComplete) && (
          <div className="space-y-2">
            <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
              <div
                className={cn(
                  'h-full transition-all duration-300 ease-out',
                  deleteComplete ? 'bg-green-500' : 'bg-primary',
                )}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
            <div className="text-sm text-muted-foreground">
              {deleteComplete ? (
                <span className="text-green-600">
                  Successfully deleted {deleteProgress.total} sandbox{deleteProgress.total !== 1 ? 'es' : ''}
                </span>
              ) : (
                <>
                  Deleting {deleteProgress.current + 1} of {deleteProgress.total}
                  {deleteProgress.currentName && `: ${deleteProgress.currentName}`}
                  {deleteProgress.status && <span className="ml-2 text-xs">({deleteProgress.status})</span>}
                </>
              )}
            </div>
          </div>
        )}

        {/* Tree view */}
        <div className="overflow-auto flex-1 min-h-0 max-h-[40vh] border rounded-md p-2 bg-muted/20">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
              <span className="ml-2 text-muted-foreground">Loading sandbox tree...</span>
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
              <span>No sandbox data available</span>
            </div>
          )}
        </div>

        <AlertDialogFooter>
          {deleteComplete ? (
            <AlertDialogAction onClick={forceClose}>Close</AlertDialogAction>
          ) : (
            <>
              <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
              <AlertDialogAction
                className={buttonVariants({ variant: 'destructive' })}
                onClick={(e) => {
                  e.preventDefault() // Prevent dialog from closing
                  performRecursiveDelete()
                }}
                disabled={loading || !loadingComplete || isDeleting || !rootNode || deleteFailed}
              >
                {isDeleting ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Deleting...
                  </>
                ) : !loadingComplete ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Loading tree...
                  </>
                ) : (
                  <>
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete All ({totalSandboxes})
                  </>
                )}
              </AlertDialogAction>
            </>
          )}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
