/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { OrganizationInvitationActionDialog } from '@/components/UserOrganizationInvitations/OrganizationInvitationActionDialog'
import { UserOrganizationInvitationTable } from '@/components/UserOrganizationInvitations/UserOrganizationInvitationTable'
import { useAcceptUserOrganizationInvitationMutation } from '@/hooks/mutations/useAcceptUserOrganizationInvitationMutation'
import { useDeclineUserOrganizationInvitationMutation } from '@/hooks/mutations/useDeclineUserOrganizationInvitationMutation'
import { useUserOrganizationInvitationsQuery } from '@/hooks/queries/useUserOrganizationInvitationsQuery'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationInvitation } from '@daytona/api-client'
import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

const UserOrganizationInvitations: React.FC = () => {
  const {
    data: invitations = [],
    isLoading: loadingInvitations,
    error: invitationsError,
  } = useUserOrganizationInvitationsQuery()
  const acceptInvitationMutation = useAcceptUserOrganizationInvitationMutation()
  const declineInvitationMutation = useDeclineUserOrganizationInvitationMutation()

  const [loadingInvitationAction, setLoadingInvitationAction] = useState<Record<string, boolean>>({})

  useEffect(() => {
    if (invitationsError) {
      handleApiError(invitationsError, 'Failed to fetch invitations')
    }
  }, [invitationsError])

  const [searchParams, setSearchParams] = useSearchParams()
  const [invitationActionDialogOpen, setInvitationActionDialogOpen] = useState(false)
  const [selectedInvitation, setSelectedInvitation] = useState<OrganizationInvitation | null>(null)

  useEffect(() => {
    const invitationId = searchParams.get('id')
    if (invitationId && invitations.length > 0) {
      const invitation = invitations.find((i) => i.id === invitationId)
      if (invitation) {
        setSelectedInvitation(invitation)
        setInvitationActionDialogOpen(true)
      }
    }
    // clear the query parameter after processing
    if (invitationId && !loadingInvitations) {
      const newSearchParams = new URLSearchParams(searchParams)
      newSearchParams.delete('id')
      setSearchParams(newSearchParams, { replace: true })
    }
  }, [searchParams, invitations, setSearchParams, loadingInvitations])

  const handleAcceptInvitation = async (invitation: OrganizationInvitation): Promise<boolean> => {
    setLoadingInvitationAction((prev) => ({ ...prev, [invitation.id]: true }))
    try {
      await acceptInvitationMutation.mutateAsync({
        invitationId: invitation.id,
        organizationId: invitation.organizationId,
      })
      toast.success('Invitation accepted successfully')
      return true
    } catch (error) {
      handleApiError(error, 'Failed to accept invitation')
      return false
    } finally {
      setLoadingInvitationAction((prev) => ({ ...prev, [invitation.id]: false }))
    }
  }

  const handleDeclineInvitation = async (invitation: OrganizationInvitation): Promise<boolean> => {
    setLoadingInvitationAction((prev) => ({ ...prev, [invitation.id]: true }))
    try {
      await declineInvitationMutation.mutateAsync({ invitationId: invitation.id })
      toast.success('Invitation declined successfully')
      return true
    } catch (error) {
      handleApiError(error, 'Failed to decline invitation')
      return false
    } finally {
      setLoadingInvitationAction((prev) => ({ ...prev, [invitation.id]: false }))
    }
  }

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Invitations</PageTitle>
      </PageHeader>

      <PageContent size="full">
        <UserOrganizationInvitationTable
          data={invitations}
          loadingData={loadingInvitations}
          onAcceptInvitation={handleAcceptInvitation}
          onDeclineInvitation={handleDeclineInvitation}
          loadingInvitationAction={loadingInvitationAction}
        />

        {selectedInvitation && (
          <OrganizationInvitationActionDialog
            invitation={selectedInvitation}
            open={invitationActionDialogOpen}
            onOpenChange={setInvitationActionDialogOpen}
            onAccept={handleAcceptInvitation}
            onDecline={handleDeclineInvitation}
          />
        )}
      </PageContent>
    </PageLayout>
  )
}

export default UserOrganizationInvitations
