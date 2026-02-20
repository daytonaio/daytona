/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { cn, pluralize } from '@/lib/utils'
import { DeletePrerequisite, DeletePrerequisiteItem, parseDeleteErrors } from '@/components/DeletePrerequisiteItem'
import { Spinner } from '@/components/ui/spinner'
import React, { useState } from 'react'

interface DeleteOrganizationDialogProps {
  organizationName: string
  onDeleteOrganization: () => Promise<{ success: boolean; reasons: string[] }>
  loading: boolean
}

export const DeleteOrganizationDialog: React.FC<DeleteOrganizationDialogProps> = ({
  organizationName,
  onDeleteOrganization,
  loading,
}) => {
  const [open, setOpen] = useState(false)
  const [confirmName, setConfirmName] = useState('')
  const [prerequisites, setPrerequisites] = useState<DeletePrerequisite[]>([])

  const handleDeleteOrganization = async () => {
    setPrerequisites([])
    const result = await onDeleteOrganization()
    if (result.success) {
      setOpen(false)
      setConfirmName('')
      setPrerequisites([])
    } else {
      setPrerequisites(parseDeleteErrors(result.reasons))
    }
  }

  const hasBlockers = prerequisites.length > 0

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setConfirmName('')
          setPrerequisites([])
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="destructive" className="w-auto px-4">
          Delete Organization
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className={cn('flex items-center gap-2', hasBlockers && 'text-destructive')}>
            {hasBlockers ? 'Cannot Delete Organization' : 'Delete Organization'}
          </DialogTitle>
          <DialogDescription>
            {hasBlockers
              ? 'We found active resources or requirements that must be resolved before deletion can proceed.'
              : 'This will permanently delete all associated data. This action cannot be undone.'}
          </DialogDescription>
        </DialogHeader>

        {hasBlockers ? (
          <div className="-mx-6 border-y">
            {prerequisites.map((prereq, i) => (
              <React.Fragment key={prereq.id}>
                {i > 0 && <Separator />}
                <DeletePrerequisiteItem prereq={prereq} />
              </React.Fragment>
            ))}
          </div>
        ) : (
          <form
            id="delete-organization-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleDeleteOrganization()
            }}
          >
            <div className="space-y-6">
              <div className="space-y-3">
                <Label htmlFor="confirm-action">
                  Please type <span className="font-bold cursor-text select-all">{organizationName}</span> to confirm
                </Label>
                <Input
                  id="confirm-action"
                  value={confirmName}
                  onChange={(e) => setConfirmName(e.target.value)}
                  placeholder={organizationName}
                />
              </div>
            </div>
          </form>
        )}

        <DialogFooter className="sm:justify-between">
          {hasBlockers && (
            <p className="text-sm text-muted-foreground">
              {pluralize(prerequisites.length, 'issue', 'issues')} preventing deletion.
            </p>
          )}
          <div className="flex gap-2 ml-auto">
            <DialogClose asChild>
              <Button type="button" variant="secondary" disabled={loading}>
                Cancel
              </Button>
            </DialogClose>
            <Button
              type={hasBlockers ? 'button' : 'submit'}
              form={hasBlockers ? undefined : 'delete-organization-form'}
              variant="destructive"
              disabled={hasBlockers || confirmName !== organizationName || loading}
            >
              {loading && <Spinner />}
              Delete Organization
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
