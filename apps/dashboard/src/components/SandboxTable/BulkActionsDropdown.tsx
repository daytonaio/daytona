/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '../ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog'
import { ChevronDown, Play, Square, Archive } from 'lucide-react'
import { Sandbox, SandboxState } from '@daytonaio/api-client'

interface BulkActionsDropdownProps {
  selectedSandboxes: Sandbox[]
  selectedCount: number
  onBulkStart: (ids: string[]) => void
  onBulkStop: (ids: string[]) => void
  onBulkArchive: (ids: string[]) => void
  onBulkDelete: (ids: string[]) => void
  onClearSelection: () => void
}

interface BulkAction {
  label: string
  icon: React.ReactNode
  action: 'start' | 'stop' | 'archive' | 'delete'
  color?: string
  destructive?: boolean
}

export function BulkActionsDropdown({
  selectedSandboxes,
  selectedCount,
  onBulkStart,
  onBulkStop,
  onBulkArchive,
  onBulkDelete,
  onClearSelection,
}: BulkActionsDropdownProps) {
  const [confirmAction, setConfirmAction] = useState<BulkAction | null>(null)
  const [isOpen, setIsOpen] = useState(false)

  // Check which actions are available based on selected sandboxes states
  const getAvailableActions = (): BulkAction[] => {
    const actions: BulkAction[] = []

    // Count sandboxes by state
    const stoppedCount = selectedSandboxes.filter((s) => s.state === SandboxState.STOPPED).length
    const startedCount = selectedSandboxes.filter((s) => s.state === SandboxState.STARTED).length
    const archivedCount = selectedSandboxes.filter((s) => s.state === SandboxState.ARCHIVED).length

    // Start action - available if any sandbox is stopped or archived
    if (stoppedCount > 0 || archivedCount > 0) {
      actions.push({
        label: `Start ${stoppedCount + archivedCount > 1 ? 'All' : ''}`,
        icon: <Play className="w-4 h-4" />,
        action: 'start',
      })
    }

    // Stop action - available if any sandbox is started
    if (startedCount > 0) {
      actions.push({
        label: `Stop ${startedCount > 1 ? 'All' : ''}`,
        icon: <Square className="w-4 h-4" />,
        action: 'stop',
      })
    }

    // Archive action - available if any sandbox is stopped
    if (stoppedCount > 0) {
      actions.push({
        label: `Archive ${stoppedCount > 1 ? 'All' : ''}`,
        icon: <Archive className="w-4 h-4" />,
        action: 'archive',
      })
    }

    return actions
  }

  const handleActionClick = (action: BulkAction) => {
    setConfirmAction(action)
    setIsOpen(false)
  }

  const handleConfirmAction = () => {
    if (!confirmAction) return

    const selectedIds = selectedSandboxes.map((s) => s.id)

    switch (confirmAction.action) {
      case 'start':
        onBulkStart(
          selectedIds.filter((id) => {
            const sandbox = selectedSandboxes.find((s) => s.id === id)
            return sandbox?.state === SandboxState.STOPPED || sandbox?.state === SandboxState.ARCHIVED
          }),
        )
        break
      case 'stop':
        onBulkStop(
          selectedIds.filter((id) => {
            const sandbox = selectedSandboxes.find((s) => s.id === id)
            return sandbox?.state === SandboxState.STARTED
          }),
        )
        break
      case 'archive':
        onBulkArchive(
          selectedIds.filter((id) => {
            const sandbox = selectedSandboxes.find((s) => s.id === id)
            return sandbox?.state === SandboxState.STOPPED
          }),
        )
        break
      case 'delete':
        onBulkDelete(selectedIds)
        break
    }

    setConfirmAction(null)
    onClearSelection()
  }

  const getActionDescription = (action: BulkAction): string => {
    const count = selectedCount
    const actionWord = action.action

    switch (action.action) {
      case 'start':
        return `This will start ${count === 1 ? 'this sandbox' : `these ${count} sandboxes`}. Only stopped and archived sandboxes will be affected.`
      case 'stop':
        return `This will stop ${count === 1 ? 'this sandbox' : `these ${count} sandboxes`}. Only started sandboxes will be affected.`
      case 'archive':
        return `This will archive ${count === 1 ? 'this sandbox' : `these ${count} sandboxes`}. Only stopped sandboxes will be affected. Archived sandboxes are moved to cost-effective storage.`
      case 'delete':
        return `This will permanently delete ${count === 1 ? 'this sandbox' : `these ${count} sandboxes`}. This action cannot be undone.`
      default:
        return `This will ${actionWord} ${count === 1 ? 'this sandbox' : `these ${count} sandboxes`}.`
    }
  }

  const availableActions = getAvailableActions()

  return (
    <>
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="h-8">
            Actions
            <ChevronDown className="w-4 h-4 ml-1" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48">
          {availableActions.map((action, index) => (
            <div key={action.action}>
              <DropdownMenuItem onClick={() => handleActionClick(action)} className={action.color}>
                {action.icon}
                {action.label}
              </DropdownMenuItem>
              {index === availableActions.length - 2 && availableActions[availableActions.length - 1].destructive && (
                <DropdownMenuSeparator />
              )}
            </div>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      {confirmAction && (
        <AlertDialog open={true} onOpenChange={() => setConfirmAction(null)}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{confirmAction.label} Sandboxes</AlertDialogTitle>
              <AlertDialogDescription>{getActionDescription(confirmAction)}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirmAction}
                className={
                  confirmAction.destructive ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90' : ''
                }
              >
                {confirmAction.label}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </>
  )
}
