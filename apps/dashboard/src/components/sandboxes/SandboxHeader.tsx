/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { SandboxState } from '@/components/SandboxTable/SandboxState'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { isArchivable, isRecoverable, isStartable, isStoppable } from '@/lib/utils/sandbox'
import { Sandbox } from '@daytonaio/api-client'
import { ArrowLeft, MoreHorizontal, Play, RefreshCw, Square, Wrench } from 'lucide-react'

interface SandboxHeaderProps {
  sandbox: Sandbox | undefined
  isLoading: boolean
  isOwner: boolean
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
  mutations: { start: boolean; stop: boolean; archive: boolean; recover: boolean }
}

export function SandboxHeader({
  sandbox,
  isLoading,
  isOwner,
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
  mutations,
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
              <span className="text-[13px] text-muted-foreground font-mono truncate">{sandbox.id}</span>
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
              {isOwner && (
                <ButtonGroup>
                  {isStartable(sandbox) && !sandbox.recoverable && (
                    <Button variant="outline" size="sm" onClick={onStart} disabled={actionsDisabled}>
                      {mutations.start ? <Spinner className="size-4" /> : <Play className="size-4" />}
                      Start
                    </Button>
                  )}
                  {isStoppable(sandbox) && (
                    <Button variant="outline" size="sm" onClick={onStop} disabled={actionsDisabled}>
                      {mutations.stop ? <Spinner className="size-4" /> : <Square className="size-4" />}
                      Stop
                    </Button>
                  )}
                  {isRecoverable(sandbox) && (
                    <Button variant="outline" size="sm" onClick={onRecover} disabled={actionsDisabled}>
                      {mutations.recover ? <Spinner className="size-4" /> : <Wrench className="size-4" />}
                      Recover
                    </Button>
                  )}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" size="icon-sm" aria-label="More actions">
                        <MoreHorizontal className="size-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-48">
                      <DropdownMenuGroup>
                        <DropdownMenuItem onClick={onCreateSshAccess} disabled={actionsDisabled}>
                          Create SSH Access
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={onRevokeSshAccess} disabled={actionsDisabled}>
                          Revoke SSH Access
                        </DropdownMenuItem>
                      </DropdownMenuGroup>
                      {isArchivable(sandbox) && (
                        <>
                          <DropdownMenuSeparator />
                          <DropdownMenuGroup>
                            <DropdownMenuItem onClick={onArchive} disabled={actionsDisabled}>
                              Archive
                            </DropdownMenuItem>
                          </DropdownMenuGroup>
                        </>
                      )}
                      <DropdownMenuSeparator />
                      <DropdownMenuGroup>
                        <DropdownMenuItem variant="destructive" onClick={onDelete} disabled={actionsDisabled}>
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuGroup>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </ButtonGroup>
              )}
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
