/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getRelativeTimeString } from '@/lib/utils'
import { Sandbox } from '@daytona/api-client'
import { ChevronDown, ChevronRight, GitFork } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { SandboxState } from './SandboxTable/SandboxState'

interface ForkTreeDialogProps {
  sandboxId: string
  open: boolean
  onClose: () => void
}

interface TreeNode {
  sandbox: Sandbox
  children: TreeNode[]
  expanded: boolean
  loading: boolean
}

const MAX_DEPTH = 3

function findNode(nodes: TreeNode[], id: string): TreeNode | null {
  for (const n of nodes) {
    if (n.sandbox.id === id) return n
    const found = findNode(n.children, id)
    if (found) return found
  }
  return null
}

function updateNodes(nodes: TreeNode[], targetId: string, updater: (node: TreeNode) => TreeNode): TreeNode[] {
  return nodes.map((n) => {
    if (n.sandbox.id === targetId) return updater(n)
    if (n.children.length > 0) return { ...n, children: updateNodes(n.children, targetId, updater) }
    return n
  })
}

function TreeNodeRow({ node, depth, onExpand }: { node: TreeNode; depth: number; onExpand: (nodeId: string) => void }) {
  const canExpand = depth < MAX_DEPTH

  return (
    <div>
      <div
        className="flex items-center gap-2 py-1.5 rounded-md hover:bg-muted/50"
        style={{ paddingLeft: `${depth * 20 + 8}px`, paddingRight: '8px' }}
      >
        <button
          className="w-4 h-4 flex items-center justify-center text-muted-foreground hover:text-foreground flex-shrink-0"
          onClick={() => canExpand && onExpand(node.sandbox.id)}
          disabled={!canExpand || node.loading}
        >
          {node.loading ? (
            <div className="w-3 h-3 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          ) : canExpand && node.expanded ? (
            <ChevronDown className="w-3.5 h-3.5" />
          ) : canExpand ? (
            <ChevronRight className="w-3.5 h-3.5" />
          ) : null}
        </button>
        <GitFork className="w-3.5 h-3.5 text-muted-foreground flex-shrink-0" />
        <span className="text-sm truncate flex-1">{node.sandbox.name}</span>
        <SandboxState state={node.sandbox.state} />
        <span className="text-xs text-muted-foreground whitespace-nowrap ml-2">
          {getRelativeTimeString(node.sandbox.createdAt).relativeTimeString}
        </span>
      </div>
      {node.expanded && (
        <div>
          {node.children.length > 0 ? (
            node.children.map((child) => (
              <TreeNodeRow key={child.sandbox.id} node={child} depth={depth + 1} onExpand={onExpand} />
            ))
          ) : (
            <div
              className="text-xs text-muted-foreground py-1"
              style={{ paddingLeft: `${(depth + 1) * 20 + 8 + 16 + 8}px` }}
            >
              No forks
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export function ForkTreeDialog({ sandboxId, open, onClose }: ForkTreeDialogProps) {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const [ancestors, setAncestors] = useState<Sandbox[]>([])
  const [currentSandbox, setCurrentSandbox] = useState<Sandbox | null>(null)
  const [childNodes, setChildNodes] = useState<TreeNode[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open || !sandboxId || !selectedOrganization?.id) return

    const fetchData = async () => {
      setLoading(true)
      setAncestors([])
      setCurrentSandbox(null)
      setChildNodes([])
      try {
        const [sandboxRes, ancestorsRes, forksRes] = await Promise.all([
          sandboxApi.getSandbox(sandboxId, selectedOrganization.id),
          sandboxApi.getSandboxAncestors(sandboxId, selectedOrganization.id),
          sandboxApi.getSandboxForks(sandboxId, selectedOrganization.id),
        ])
        setCurrentSandbox(sandboxRes.data)
        setAncestors(ancestorsRes.data)
        setChildNodes(
          forksRes.data.map((s) => ({
            sandbox: s,
            children: [],
            expanded: false,
            loading: false,
          })),
        )
      } catch {
        toast.error('Failed to load fork tree')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [open, sandboxId, selectedOrganization?.id, sandboxApi])

  const handleExpand = useCallback(
    async (nodeId: string) => {
      if (!selectedOrganization?.id) return

      const target = findNode(childNodes, nodeId)
      if (!target) return

      if (target.expanded) {
        setChildNodes((prev) => updateNodes(prev, nodeId, (n) => ({ ...n, expanded: false })))
        return
      }

      if (target.children.length > 0) {
        setChildNodes((prev) => updateNodes(prev, nodeId, (n) => ({ ...n, expanded: true })))
        return
      }

      setChildNodes((prev) => updateNodes(prev, nodeId, (n) => ({ ...n, loading: true })))
      try {
        const forksRes = await sandboxApi.getSandboxForks(nodeId, selectedOrganization.id)
        setChildNodes((prev) =>
          updateNodes(prev, nodeId, (n) => ({
            ...n,
            loading: false,
            expanded: true,
            children: forksRes.data.map((s) => ({
              sandbox: s,
              children: [],
              expanded: false,
              loading: false,
            })),
          })),
        )
      } catch {
        setChildNodes((prev) => updateNodes(prev, nodeId, (n) => ({ ...n, loading: false })))
        toast.error('Failed to load forks')
      }
    },
    [childNodes, sandboxApi, selectedOrganization?.id],
  )

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-lg max-h-[70vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <GitFork className="w-4 h-4" />
            Fork Tree
          </DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-y-auto -mx-6 px-6">
          {loading ? (
            <div className="flex items-center justify-center py-8 text-muted-foreground text-sm">Loading...</div>
          ) : (
            <div className="space-y-0.5 py-1">
              {ancestors.map((ancestor) => (
                <div key={ancestor.id} className="flex items-center gap-2 py-1.5 px-2 rounded-md hover:bg-muted/50">
                  <div className="w-4 h-4 flex-shrink-0" />
                  <GitFork className="w-3.5 h-3.5 text-muted-foreground flex-shrink-0" />
                  <span className="text-sm truncate flex-1">{ancestor.name}</span>
                  <SandboxState state={ancestor.state} />
                  <span className="text-xs text-muted-foreground whitespace-nowrap ml-2">
                    {getRelativeTimeString(ancestor.createdAt).relativeTimeString}
                  </span>
                </div>
              ))}

              {currentSandbox && (
                <div className="flex items-center gap-2 py-1.5 px-2 rounded-md bg-primary/10 border border-primary/20">
                  <div className="w-4 h-4 flex-shrink-0" />
                  <GitFork className="w-3.5 h-3.5 text-primary flex-shrink-0" />
                  <span className="text-sm font-medium text-primary truncate flex-1">
                    {currentSandbox.name}
                    <span className="ml-2 text-xs text-primary/60 font-normal">(current)</span>
                  </span>
                  <SandboxState state={currentSandbox.state} />
                  <span className="text-xs text-muted-foreground whitespace-nowrap ml-2">
                    {getRelativeTimeString(currentSandbox.createdAt).relativeTimeString}
                  </span>
                </div>
              )}

              {childNodes.map((node) => (
                <TreeNodeRow key={node.sandbox.id} node={node} depth={1} onExpand={handleExpand} />
              ))}

              {!loading && childNodes.length === 0 && currentSandbox && ancestors.length === 0 && (
                <div className="text-xs text-muted-foreground py-2 px-2">No forks found</div>
              )}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
