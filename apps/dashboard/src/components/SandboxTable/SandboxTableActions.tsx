/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '@daytonaio/api-client'
import { Terminal, MoreVertical, Play, Square, Loader2 } from 'lucide-react'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { SandboxTableActionsProps } from './types'
import { useMemo } from 'react'

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
  onOpenWebTerminal,
}: SandboxTableActionsProps) {
  const menuItems = useMemo(() => {
    const items = []

    if (writePermitted) {
      if (sandbox.state === SandboxState.STARTED) {
        items.push({
          key: 'vnc',
          label: 'VNC',
          onClick: () => onVnc(sandbox.id),
          disabled: isLoading,
        })
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
      }

      if (sandbox.state === SandboxState.STOPPED) {
        items.push({
          key: 'archive',
          label: 'Archive',
          onClick: () => onArchive(sandbox.id),
          disabled: isLoading,
        })
      }
    }

    if (deletePermitted) {
      if (items.length > 0 && (sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.STARTED)) {
        items.push({ key: 'separator', type: 'separator' })
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
    onStart,
    onStop,
    onDelete,
    onArchive,
    onVnc,
  ])

  if (!writePermitted && !deletePermitted) {
    return null
  }

  return (
    <div className="flex items-center justify-end gap-2">
      <Button
        variant="outline"
        className="h-7 w-7 p-0 text-muted-foreground"
        onClick={(e) => {
          e.stopPropagation()
          if (sandbox.state === SandboxState.STARTED) {
            onStop(sandbox.id)
          } else {
            onStart(sandbox.id)
          }
        }}
      >
        {sandbox.state === SandboxState.STARTED ? (
          <Square className="w-4 h-4" />
        ) : sandbox.state === SandboxState.STOPPING || sandbox.state === SandboxState.STARTING ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <Play className="w-4 h-4" />
        )}
      </Button>

      {sandbox.state === SandboxState.STARTED ? (
        <Button
          variant="outline"
          className="h-7 w-7 p-0 text-muted-foreground"
          onClick={() => onOpenWebTerminal(sandbox.id)}
        >
          <Terminal className="w-4 h-4" />
        </Button>
      ) : (
        <Button variant="outline" className="h-7 w-7 p-0 text-muted-foreground" disabled>
          <Terminal className="w-4 h-4" />
        </Button>
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="h-7 w-7 p-0 text-muted-foreground">
            <span className="sr-only">Open menu</span>
            <MoreVertical />
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
