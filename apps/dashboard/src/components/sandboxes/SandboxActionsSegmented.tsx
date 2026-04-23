/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { isArchivable, isPausable, isRecoverable, isStartable, isStoppable, isTransitioning } from '@/lib/utils/sandbox'
import { SandboxClass, SandboxListItem, SandboxState } from '@daytona/api-client'
import { Archive, Play, Square, Trash2, Wrench, Pause } from 'lucide-react'

interface SandboxActionsSegmentedProps {
  sandbox: SandboxListItem
  writePermitted: boolean
  deletePermitted: boolean
  actionsDisabled: boolean
  onStart: () => void
  onStop: () => void
  onPause: () => void
  onArchive: () => void
  onRecover: () => void
  onDelete: () => void
  onCreateSshAccess: () => void
  onRevokeSshAccess: () => void
  onScreenRecordings: () => void
}

type PrimaryAction = 'start' | 'stop' | 'archive' | 'recover'
type ActionVisibility = Record<PrimaryAction, boolean>

const emptyActions: ActionVisibility = {
  start: false,
  stop: false,
  archive: false,
  recover: false,
}

function getVisibleActions(sandbox: SandboxListItem, writePermitted: boolean): ActionVisibility {
  if (!writePermitted) {
    return emptyActions
  }

  switch (sandbox.state) {
    case SandboxState.CREATING:
    case SandboxState.PULLING_SNAPSHOT:
    case SandboxState.BUILDING_SNAPSHOT:
    case SandboxState.STARTING:
    case SandboxState.RESTORING:
      return { ...emptyActions, stop: true }
    case SandboxState.STOPPING:
      return { ...emptyActions, start: true, archive: true }
    case SandboxState.ARCHIVING:
      return { ...emptyActions, start: true }
  }

  return {
    start: isStartable(sandbox) && !sandbox.recoverable,
    stop: isStoppable(sandbox),
    archive: isArchivable(sandbox),
    recover: isRecoverable(sandbox),
  }
}

export function SandboxActionsSegmented({
  sandbox,
  writePermitted,
  deletePermitted,
  actionsDisabled,
  onStart,
  onStop,
  onPause,
  onArchive,
  onRecover,
  onDelete,
}: SandboxActionsSegmentedProps) {
  const isVmSandbox = sandbox.sandboxClass === SandboxClass.LINUX_VM || sandbox.sandboxClass === SandboxClass.WINDOWS
  const pauseResumeAvailable = isVmSandbox && writePermitted

  const actionsLocked = actionsDisabled || isTransitioning(sandbox)
  const visibleActions = getVisibleActions(sandbox, writePermitted)
  const showStart = visibleActions.start
  const showStop = visibleActions.stop
  const showPause = pauseResumeAvailable && isPausable(sandbox)
  const showArchive = visibleActions.archive
  const showRecover = visibleActions.recover
  const showDelete = deletePermitted

  return (
    <ButtonGroup className="empty:hidden">
      {showStart && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="outline" size="sm" onClick={onStart} disabled={actionsLocked}>
              <Play className="size-4" />
              Start
            </Button>
          </TooltipTrigger>
          <TooltipContent>{sandbox.state === SandboxState.PAUSED ? 'Resume sandbox' : 'Start sandbox'}</TooltipContent>
        </Tooltip>
      )}
      {showStop && (
        <Button variant="outline" size="sm" onClick={onStop} disabled={actionsLocked}>
          <Square className="size-4" />
          Stop
        </Button>
      )}
      {showPause && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon-sm"
              onClick={onPause}
              disabled={actionsLocked}
              aria-label="Pause sandbox"
            >
              <Pause className="size-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Pause</TooltipContent>
        </Tooltip>
      )}
      {showRecover && (
        <Button variant="outline" size="sm" onClick={onRecover} disabled={actionsLocked}>
          <Wrench className="size-4" />
          Recover
        </Button>
      )}
      {showArchive && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon-sm"
              onClick={onArchive}
              disabled={actionsLocked}
              aria-label="Archive sandbox"
            >
              <Archive className="size-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Archive</TooltipContent>
        </Tooltip>
      )}
      {showDelete && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon-sm"
              onClick={onDelete}
              disabled={actionsLocked}
              aria-label="Delete sandbox"
              className="text-destructive-foreground hover:bg-destructive/10 hover:text-destructive-foreground"
            >
              <Trash2 className="size-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Delete</TooltipContent>
        </Tooltip>
      )}
    </ButtonGroup>
  )
}
