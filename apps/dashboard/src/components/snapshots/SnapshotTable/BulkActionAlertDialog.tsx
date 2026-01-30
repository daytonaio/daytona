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
} from '../../ui/alert-dialog'

export enum SnapshotBulkAction {
  Delete = 'delete',
  Deactivate = 'deactivate',
}

interface BulkActionData {
  title: string
  description: string
  buttonLabel: string
  buttonVariant?: 'destructive'
}

function getBulkActionData(action: SnapshotBulkAction, count: number): BulkActionData {
  const countText = count === 1 ? 'this snapshot' : `these ${count} selected snapshots`

  switch (action) {
    case SnapshotBulkAction.Delete:
      return {
        title: 'Delete Snapshots',
        description: `Are you sure you want to delete ${countText}? This action cannot be undone.`,
        buttonLabel: 'Delete',
        buttonVariant: 'destructive',
      }
    case SnapshotBulkAction.Deactivate:
      return {
        title: 'Deactivate Snapshots',
        description: `Are you sure you want to deactivate ${countText}? Deactivated snapshots can be reactivated later.`,
        buttonLabel: 'Deactivate',
      }
  }
}

interface SnapshotBulkActionAlertDialogProps {
  action: SnapshotBulkAction | null
  count: number
  onConfirm: () => void
  onCancel: () => void
}

export function SnapshotBulkActionAlertDialog({
  action,
  count,
  onConfirm,
  onCancel,
}: SnapshotBulkActionAlertDialogProps) {
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
