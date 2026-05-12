/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FeatureFlags } from '@/enums/FeatureFlags'
import { useRegions } from '@/hooks/useRegions'
import { SandboxState } from '@daytona/api-client'
import { Loader2, MoreHorizontal, Play, Square, Terminal, Wrench } from 'lucide-react'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { useMemo } from 'react'
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
  runnerClass,
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
  const linuxVmEnabled = useFeatureFlagEnabled(FeatureFlags.SANDBOX_LINUX_VM)
  const { getRegionName } = useRegions()
  const isExperimentalRegion = (getRegionName(sandbox.target) ?? '').toLowerCase() === 'experimental'
  const primaryActionTooltip =
    sandbox.state === SandboxState.STARTING
      ? 'Starting sandbox'
      : sandbox.state === SandboxState.STOPPING
        ? 'Stopping sandbox'
        : sandbox.state === SandboxState.STARTED
          ? 'Stop sandbox'
          : sandbox.state === SandboxState.ERROR && sandbox.recoverable
            ? 'Recover sandbox'
            : 'Start sandbox'

  const menuItems = useMemo(() => {
    const items = []

    if (writePermitted) {
      if (sandbox.state === SandboxState.STARTED) {
        items.push({
          key: 'stop',
          label: 'Stop',
          onClick: () => onStop(sandbox.id),
          disabled: isLoading,
        })
      } else if (sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) {
        items.push({
          key: 'start',
          label: 'Start',
          onClick: () => onStart(sandbox.id),
          disabled: isLoading,
        })
      } else if (sandbox.state === SandboxState.ERROR && sandbox.recoverable) {
        items.push({
          key: 'recover',
          label: 'Recover',
          onClick: () => onRecover(sandbox.id),
          disabled: isLoading,
        })
      }

      if (sandbox.state === SandboxState.STOPPED) {
        items.push({
          key: 'archive',
          label: 'Archive',
          onClick: () => onArchive(sandbox.id),
          disabled: isLoading,
        })
      }

      if (items.length > 0) {
        items.push({ key: 'lifecycle-separator', type: 'separator' })
      }

      if (sandbox.state === SandboxState.STARTED) {
        items.push({
          key: 'vnc',
          label: 'VNC',
          onClick: () => onVnc(sandbox.id),
          disabled: isLoading,
        })
        items.push({
          key: 'screen-recordings',
          label: 'Screen Recordings',
          onClick: () => onScreenRecordings(sandbox.id),
          disabled: isLoading,
        })
      }

      if (
        linuxVmEnabled &&
        isExperimentalRegion &&
        (sandbox.state === SandboxState.STARTED || sandbox.state === SandboxState.STOPPED)
      ) {
        items.push({
          key: 'create-snapshot',
          label: 'Create Snapshot',
          onClick: () => onCreateSnapshot?.(),
          disabled: isLoading,
        })

        items.push({
          key: 'fork',
          label: 'Fork',
          onClick: () => onFork?.(),
          disabled: isLoading,
        })
      }

      if (linuxVmEnabled && isExperimentalRegion) {
        items.push({
          key: 'view-forks',
          label: 'View Fork Tree',
          onClick: () => onViewForks?.(),
          disabled: isLoading,
        })
      }

      // Add SSH access options
      items.push({
        key: 'create-ssh',
        label: 'Create SSH Access',
        onClick: () => onCreateSshAccess(sandbox.id),
        disabled: isLoading,
      })
      items.push({
        key: 'revoke-ssh',
        label: 'Revoke SSH Access',
        onClick: () => onRevokeSshAccess(sandbox.id),
        disabled: isLoading,
      })
    }

    if (deletePermitted) {
      if (items.length > 0) {
        items.push({ key: 'delete-separator', type: 'separator' })
      }

      items.push({
        key: 'delete',
        label: 'Delete',
        onClick: () => onDelete(sandbox.id),
        disabled: isLoading,
        className: 'text-red-600 dark:text-red-400',
      })
    }

    return items
  }, [
    writePermitted,
    deletePermitted,
    sandbox.state,
    sandbox.id,
    isLoading,
    runnerClass,
    sandbox.recoverable,
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
    linuxVmEnabled,
    isExperimentalRegion,
    sandbox.target,
  ])

  if (!writePermitted && !deletePermitted) {
    return null
  }

  return (
    <div className="flex items-center justify-end gap-2">
      <TooltipButton
        variant="ghost"
        size="icon-sm"
        tooltipText={primaryActionTooltip}
        aria-label={primaryActionTooltip}
        onClick={(e) => {
          e.stopPropagation()
          if (sandbox.state === SandboxState.STARTED) {
            onStop(sandbox.id)
          } else if (sandbox.state === SandboxState.ERROR && sandbox.recoverable) {
            onRecover(sandbox.id)
          } else {
            onStart(sandbox.id)
          }
        }}
      >
        {sandbox.state === SandboxState.STARTED ? (
          <Square className="w-4 h-4" />
        ) : sandbox.state === SandboxState.STOPPING || sandbox.state === SandboxState.STARTING ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : sandbox.state === SandboxState.ERROR && sandbox.recoverable ? (
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
      >
        <Terminal className="w-4 h-4" />
      </Button>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon-sm" aria-label="Open menu">
            <MoreHorizontal className="w-4 h-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {menuItems.map((item) => {
            if (item.type === 'separator') {
              return <DropdownMenuSeparator key={item.key} />
            }

            return (
              <DropdownMenuItem
                key={item.key}
                onClick={(e) => {
                  e.stopPropagation()
                  item.onClick?.()
                }}
                className={`cursor-pointer ${item.className || ''}`}
                disabled={item.disabled}
              >
                {item.label}
              </DropdownMenuItem>
            )
          })}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
