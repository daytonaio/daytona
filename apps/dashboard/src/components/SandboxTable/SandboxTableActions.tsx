/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass, SandboxState } from '@daytona/api-client'
import { Loader2, MoreHorizontal, Play, Square, Terminal, Wrench } from 'lucide-react'
import TooltipButton from '../TooltipButton'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { SandboxTableActionsProps } from './types'

export function SandboxTableActions({
  sandbox,
  writePermitted,
  deletePermitted,
  isLoading,
  onStart,
  onStop,
  onDelete,
  onArchive,
  onVnc,
  onCreateSshAccess,
  onRevokeSshAccess,
  onCreateSnapshot,
  onRecover,
  onScreenRecordings,
  onFork,
  onViewForks,
  onOpenTerminal,
}: SandboxTableActionsProps) {
  const isVmSandbox = sandbox.sandboxClass === SandboxClass.LINUX_VM || sandbox.sandboxClass === SandboxClass.WINDOWS
  const isStarted = sandbox.state === SandboxState.STARTED
  const isStarting = sandbox.state === SandboxState.STARTING
  const isStopping = sandbox.state === SandboxState.STOPPING
  const isStopped = sandbox.state === SandboxState.STOPPED
  const isArchived = sandbox.state === SandboxState.ARCHIVED
  const isRecoverableError = sandbox.state === SandboxState.ERROR && sandbox.recoverable

  const canStopSandbox = writePermitted && isStarted
  const canStartSandbox = writePermitted && (isStopped || isArchived)
  const canRecoverSandbox = writePermitted && isRecoverableError
  const canArchiveSandbox = writePermitted && isStopped
  const canOpenVnc = writePermitted && isStarted
  const canViewScreenRecordings = writePermitted && isStarted
  const canCreateSnapshot = writePermitted && sandbox.gpu === 0 && (isStarted || isStopped)
  const canForkSandbox = writePermitted && isVmSandbox && isStarted
  const canCreateSshAccess = writePermitted
  const canRevokeSshAccess = writePermitted
  const canViewForkTree = isVmSandbox

  const hasLifecycleActions = canStopSandbox || canStartSandbox || canRecoverSandbox || canArchiveSandbox
  const hasMenuItemsBeforeDelete = writePermitted || canViewForkTree
  const hasMenuItems = hasMenuItemsBeforeDelete || deletePermitted
  const showLifecycleSeparator = hasLifecycleActions
  const showDeleteSeparator = deletePermitted && hasMenuItemsBeforeDelete

  const primaryActionTooltip = isStarting
    ? 'Starting sandbox'
    : isStopping
      ? 'Stopping sandbox'
      : isStarted
        ? 'Stop sandbox'
        : isRecoverableError
          ? 'Recover sandbox'
          : 'Start sandbox'

  if (!hasMenuItems) {
    return null
  }

  // The primary start/stop/recover and terminal controls are write actions; hide them from
  // users who can only view (e.g. read-only users reaching the read-only "View Fork Tree").
  const showWriteControls = writePermitted || deletePermitted

  return (
    <div className="flex items-center justify-end gap-2">
      {showWriteControls && (
        <>
          <TooltipButton
            variant="ghost"
            size="icon-sm"
            tooltipText={primaryActionTooltip}
            aria-label={primaryActionTooltip}
            onClick={(e) => {
              e.stopPropagation()
              if (isStarted) {
                onStop(sandbox.id)
              } else if (isRecoverableError) {
                onRecover(sandbox.id)
              } else {
                onStart(sandbox.id)
              }
            }}
            disabled={isLoading}
          >
            {isStarted ? (
              <Square className="w-4 h-4" />
            ) : isStopping || isStarting ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : isRecoverableError ? (
              <Wrench className="w-4 h-4" />
            ) : (
              <Play className="w-4 h-4" />
            )}
          </TooltipButton>

          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Open terminal"
            onClick={(e) => {
              e.stopPropagation()
              onOpenTerminal?.()
            }}
            disabled={isLoading}
          >
            <Terminal className="w-4 h-4" />
          </Button>
        </>
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon-sm" aria-label="Open menu">
            <MoreHorizontal className="w-4 h-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {canStopSandbox && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onStop(sandbox.id)
              }}
              disabled={isLoading}
            >
              Stop
            </DropdownMenuItem>
          )}
          {canStartSandbox && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onStart(sandbox.id)
              }}
              disabled={isLoading}
            >
              Start
            </DropdownMenuItem>
          )}
          {canRecoverSandbox && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onRecover(sandbox.id)
              }}
              disabled={isLoading}
            >
              Recover
            </DropdownMenuItem>
          )}
          {canArchiveSandbox && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onArchive(sandbox.id)
              }}
              disabled={isLoading}
            >
              Archive
            </DropdownMenuItem>
          )}
          {showLifecycleSeparator && <DropdownMenuSeparator />}
          {canOpenVnc && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onVnc(sandbox.id)
              }}
              disabled={isLoading}
            >
              VNC
            </DropdownMenuItem>
          )}
          {canViewScreenRecordings && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onScreenRecordings(sandbox.id)
              }}
              disabled={isLoading}
            >
              Screen Recordings
            </DropdownMenuItem>
          )}
          {canCreateSnapshot && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onCreateSnapshot?.()
              }}
              disabled={isLoading}
            >
              Create Snapshot
            </DropdownMenuItem>
          )}
          {canForkSandbox && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onFork?.()
              }}
              disabled={isLoading}
            >
              Fork
            </DropdownMenuItem>
          )}
          {canCreateSshAccess && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onCreateSshAccess(sandbox.id)
              }}
              disabled={isLoading}
            >
              Create SSH Access
            </DropdownMenuItem>
          )}
          {canRevokeSshAccess && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onRevokeSshAccess(sandbox.id)
              }}
              disabled={isLoading}
            >
              Revoke SSH Access
            </DropdownMenuItem>
          )}
          {canViewForkTree && (
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                onViewForks?.()
              }}
              disabled={isLoading}
            >
              View Fork Tree
            </DropdownMenuItem>
          )}
          {showDeleteSeparator && <DropdownMenuSeparator />}
          {deletePermitted && (
            <DropdownMenuItem
              variant="destructive"
              onClick={(e) => {
                e.stopPropagation()
                onDelete(sandbox.id)
              }}
              disabled={isLoading}
            >
              Delete
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
