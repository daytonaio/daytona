/*
 * Copyright Daytona Platforms Inc.
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

export enum VolumeBulkAction {
  Delete = 'delete',
}

interface VolumeBulkActionAlertDialogProps {
  action: VolumeBulkAction | null
  count: number
  onConfirm: () => void
  onCancel: () => void
}

export function VolumeBulkActionAlertDialog({ action, count, onConfirm, onCancel }: VolumeBulkActionAlertDialogProps) {
  if (!action) return null

  const countText = count === 1 ? 'this volume' : `these ${count} selected volumes`

  return (
    <AlertDialog open={action !== null} onOpenChange={(open) => !open && onCancel()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Volumes</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete {countText}? This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} variant="destructive">
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
