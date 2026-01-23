/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { pluralize } from '@/lib/utils'
import { BulkActionCounts } from '@/lib/utils/sandbox'
import { ArchiveIcon, CheckIcon, PlayIcon, SquareIcon, StopCircleIcon, TrashIcon } from 'lucide-react'
import { useMemo } from 'react'
import { CommandConfig, useRegisterCommands } from '../CommandPalette'

interface UseSandboxCommandsProps {
  writePermitted: boolean
  deletePermitted: boolean
  selectedCount: number
  totalCount: number
  toggleAllRowsSelected: (selected: boolean) => void
  bulkActionCounts: BulkActionCounts
  onDelete: () => void
  onStart: () => void
  onStop: () => void
  onArchive: () => void
}

export function useSandboxCommands({
  writePermitted,
  deletePermitted,
  selectedCount,
  totalCount,
  toggleAllRowsSelected,
  bulkActionCounts,
  onDelete,
  onStart,
  onStop,
  onArchive,
}: UseSandboxCommandsProps) {
  const rootCommands: CommandConfig[] = useMemo(() => {
    const commands: CommandConfig[] = []

    if (totalCount !== selectedCount) {
      commands.push({
        id: 'select-all-sandboxes',
        label: 'Select All Sandboxes',
        icon: <CheckIcon className="w-4 h-4" />,
        onSelect: () => toggleAllRowsSelected(true),
        chainable: true,
      })
    }

    if (selectedCount > 0) {
      commands.push({
        id: 'deselect-all-sandboxes',
        label: 'Deselect All Sandboxes',
        icon: <SquareIcon className="w-4 h-4" />,
        onSelect: () => toggleAllRowsSelected(false),
        chainable: true,
      })
    }

    if (writePermitted && bulkActionCounts.startable > 0) {
      commands.push({
        id: 'start-sandboxes',
        label: `Start ${pluralize(bulkActionCounts.startable, 'Sandbox', 'Sandboxes')}`,
        icon: <PlayIcon className="w-4 h-4" />,
        onSelect: onStart,
      })
    }

    if (writePermitted && bulkActionCounts.stoppable > 0) {
      commands.push({
        id: 'stop-sandboxes',
        label: `Stop ${pluralize(bulkActionCounts.stoppable, 'Sandbox', 'Sandboxes')}`,
        icon: <StopCircleIcon className="w-4 h-4" />,
        onSelect: onStop,
      })
    }

    if (writePermitted && bulkActionCounts.archivable > 0) {
      commands.push({
        id: 'archive-sandboxes',
        label: `Archive ${pluralize(bulkActionCounts.archivable, 'Sandbox', 'Sandboxes')}`,
        icon: <ArchiveIcon className="w-4 h-4" />,
        onSelect: onArchive,
      })
    }

    if (deletePermitted && bulkActionCounts.deletable > 0) {
      commands.push({
        id: 'delete-sandboxes',
        label: `Delete ${pluralize(bulkActionCounts.deletable, 'Sandbox', 'Sandboxes')}`,
        icon: <TrashIcon className="w-4 h-4" />,
        onSelect: onDelete,
      })
    }

    return commands
  }, [
    selectedCount,
    totalCount,
    toggleAllRowsSelected,
    writePermitted,
    deletePermitted,
    bulkActionCounts,
    onDelete,
    onStart,
    onStop,
    onArchive,
  ])

  useRegisterCommands(rootCommands, { groupId: 'sandbox-actions', groupLabel: 'Sandbox actions', groupOrder: 0 })
}
