/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

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

export enum BulkAction {
  Delete = 'delete',
  Start = 'start',
  Stop = 'stop',
  Archive = 'archive',
}

interface BulkActionData {
  title: string
  description: string
  buttonLabel: string
  buttonVariant?: 'destructive'
}

function getBulkActionData(action: BulkAction, count: number): BulkActionData {
  const countText = count === 1 ? 'this sandbox' : `these ${count} selected sandboxes`

  switch (action) {
    case BulkAction.Delete:
      return {
        title: 'Delete Sandboxes',
        description: `Are you sure you want to delete ${countText}? This action cannot be undone.`,
        buttonLabel: 'Delete',
        buttonVariant: 'destructive',
      }
    case BulkAction.Start:
      return {
        title: 'Start Sandboxes',
        description: `Are you sure you want to start ${countText}?`,
        buttonLabel: 'Start',
      }
    case BulkAction.Stop:
      return {
        title: 'Stop Sandboxes',
        description: `Are you sure you want to stop ${countText}?`,
        buttonLabel: 'Stop',
      }
    case BulkAction.Archive:
      return {
        title: 'Archive Sandboxes',
        description: `Are you sure you want to archive ${countText}? Archived sandboxes can be restored later.`,
        buttonLabel: 'Archive',
      }
  }
}

interface BulkActionAlertDialogProps {
  action: BulkAction | null
  count: number
  onConfirm: () => void
  onCancel: () => void
}

export function BulkActionAlertDialog({ action, count, onConfirm, onCancel }: BulkActionAlertDialogProps) {
  const data = action ? getBulkActionData(action, count) : null

  if (!data) return null

  return (
    <AlertDialog open={action !== null} onOpenChange={(open) => !open && onCancel()}>
      <AlertDialogContent>
        <>
          <AlertDialogHeader>
            <AlertDialogTitle>{data.title}</AlertDialogTitle>
            <AlertDialogDescription>{data.description}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={onConfirm} variant={data.buttonVariant}>
              {data.buttonLabel}
            </AlertDialogAction>
          </AlertDialogFooter>
        </>
      </AlertDialogContent>
    </AlertDialog>
  )
}
