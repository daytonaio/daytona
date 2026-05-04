/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { SandboxState } from './SandboxState'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Sandbox } from '@daytona/api-client'
import { ArrowLeft, RefreshCw } from 'lucide-react'
import { SandboxActionsSegmented } from './SandboxActionsSegmented'

interface SandboxHeaderProps {
  sandbox: Sandbox | undefined
  isLoading: boolean
  writePermitted: boolean
  deletePermitted: boolean
  actionsDisabled: boolean
  isFetching: boolean
  onStart: () => void
  onStop: () => void
  onArchive: () => void
  onRecover: () => void
  onDelete: () => void
  onRefresh: () => void
  onBack: () => void
  onCreateSshAccess: () => void
  onRevokeSshAccess: () => void
  onScreenRecordings: () => void
}

export function SandboxHeader({
  sandbox,
  isLoading,
  writePermitted,
  deletePermitted,
  actionsDisabled,
  isFetching,
  onStart,
  onStop,
  onArchive,
  onRecover,
  onDelete,
  onRefresh,
  onBack,
  onCreateSshAccess,
  onRevokeSshAccess,
  onScreenRecordings,
}: SandboxHeaderProps) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 min-w-0 px-4 sm:px-5 py-2 border-b border-border shrink-0">
      <div className="flex items-center gap-2 min-w-0">
        <Button variant="ghost" size="icon-sm" className="shrink-0" onClick={onBack}>
          <ArrowLeft className="size-4" />
        </Button>
        {isLoading ? (
          <SandboxHeaderSkeleton />
        ) : sandbox ? (
          <div className="min-w-0">
            <div className="flex items-center gap-1 min-w-0">
              <h2 className="text-base font-medium truncate">{sandbox.name || sandbox.id}</h2>
              <CopyButton value={sandbox.name || sandbox.id} tooltipText="Copy name" size="icon-xs" />
            </div>
            <div className="flex items-center gap-1 min-w-0">
              <span className="text-xs text-muted-foreground shrink-0">UUID</span>
              <span className="text-sm text-muted-foreground font-mono truncate">{sandbox.id}</span>
              <CopyButton value={sandbox.id} tooltipText="Copy ID" size="icon-xs" />
            </div>
          </div>
        ) : null}
      </div>

      <div className="flex items-center gap-3 shrink-0 ml-8 sm:ml-0">
        {isLoading ? (
          <div className="flex items-center gap-2">
            <Skeleton className="h-6 w-16" />
            <Skeleton className="h-8 w-20" />
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : sandbox ? (
          <>
            <SandboxState state={sandbox.state} errorReason={sandbox.errorReason} recoverable={sandbox.recoverable} />
            <div className="flex items-center gap-2">
              <SandboxActionsSegmented
                sandbox={sandbox}
                writePermitted={writePermitted}
                deletePermitted={deletePermitted}
                actionsDisabled={actionsDisabled}
                onStart={onStart}
                onStop={onStop}
                onArchive={onArchive}
                onRecover={onRecover}
                onDelete={onDelete}
                onCreateSshAccess={onCreateSshAccess}
                onRevokeSshAccess={onRevokeSshAccess}
                onScreenRecordings={onScreenRecordings}
              />
              <Button variant="ghost" size="icon-sm" onClick={onRefresh} disabled={isFetching} title="Refresh">
                {isFetching ? <Spinner className="size-4" /> : <RefreshCw className="size-4" />}
              </Button>
            </div>
          </>
        ) : null}
      </div>
    </div>
  )
}

function SandboxHeaderSkeleton() {
  return (
    <div className="flex flex-col gap-1">
      <div className="h-6 flex items-center">
        <Skeleton className="h-4 w-40" />
      </div>
      <div className="h-6 flex items-center">
        <Skeleton className="h-3.5 w-52" />
      </div>
    </div>
  )
}
