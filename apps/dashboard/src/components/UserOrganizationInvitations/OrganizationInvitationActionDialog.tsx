/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { OrganizationInvitation } from '@daytonaio/api-client'

interface OrganizationInvitationActionDialogProps {
  invitation: OrganizationInvitation
  open: boolean
  onOpenChange: (open: boolean) => void
  onAccept: (invitation: OrganizationInvitation) => Promise<boolean>
  onDecline: (invitation: OrganizationInvitation) => Promise<boolean>
}

export const OrganizationInvitationActionDialog: React.FC<OrganizationInvitationActionDialogProps> = ({
  invitation,
  open,
  onOpenChange,
  onAccept,
  onDecline,
}) => {
  const [loadingAccept, setLoadingAccept] = useState(false)
  const [loadingDecline, setLoadingDecline] = useState(false)

  const handleAccept = async () => {
    setLoadingAccept(true)
    const success = await onAccept(invitation)
    if (success) {
      onOpenChange(false)
    }
    setLoadingAccept(false)
  }

  const handleDecline = async () => {
    setLoadingDecline(true)
    const success = await onDecline(invitation)
    if (success) {
      onOpenChange(false)
    }
    setLoadingDecline(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Organization Invitation</DialogTitle>
          <DialogDescription>Would you like to accept or decline this invitation?</DialogDescription>
        </DialogHeader>
        <div>
          <div className="grid grid-cols-[120px_1fr] gap-2">
            <span className="text-muted-foreground">Organization:</span>
            <span className="font-medium">{invitation.organizationName}</span>

            <span className="text-muted-foreground">Invited by:</span>
            <span className="font-medium">{invitation.invitedBy || 'Not specified'}</span>

            <span className="text-muted-foreground">Expires:</span>
            <span className="font-medium">
              {new Date(invitation.expiresAt).toLocaleString('default', {
                year: 'numeric',
                month: 'numeric',
                day: 'numeric',
                hour: 'numeric',
                minute: '2-digit',
              })}
            </span>
          </div>
        </div>
        <DialogFooter>
          {loadingDecline ? (
            <Button type="button" variant="secondary" disabled>
              Declining...
            </Button>
          ) : (
            <Button type="button" variant="secondary" onClick={handleDecline} disabled={loadingAccept}>
              Decline
            </Button>
          )}
          {loadingAccept ? (
            <Button type="button" variant="default" disabled>
              Accepting...
            </Button>
          ) : (
            <Button type="button" variant="default" onClick={handleAccept} disabled={loadingDecline}>
              Accept
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
