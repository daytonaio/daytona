/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { useApi } from '@/hooks/useApi'
import { OrganizationInvitation } from '@daytonaio/api-client'
import { UserOrganizationInvitationTable } from '@/components/UserOrganizationInvitations/UserOrganizationInvitationTable'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useUserOrganizationInvitations } from '@/hooks/useUserOrganizationInvitations'
import { OrganizationInvitationActionDialog } from '@/components/UserOrganizationInvitations/OrganizationInvitationActionDialog'
import { handleApiError } from '@/lib/error-handling'

const UserOrganizationInvitations: React.FC = () => {
  const { organizationsApi } = useApi()

  const { refreshOrganizations } = useOrganizations()
  const { setCount } = useUserOrganizationInvitations()

  const [invitations, setInvitations] = useState<OrganizationInvitation[]>([])
  const [loadingInvitations, setLoadingInvitations] = useState(true)

  const [loadingInvitationAction, setLoadingInvitationAction] = useState<Record<string, boolean>>({})

  const fetchInvitations = useCallback(
    async (showTableLoadingState = true) => {
      if (showTableLoadingState) {
        setLoadingInvitations(true)
      }
      try {
        const response = await organizationsApi.listOrganizationInvitationsForAuthenticatedUser()
        setInvitations(response.data)
        setCount(response.data.length)
      } catch (error) {
        handleApiError(error, 'Failed to fetch invitations')
      } finally {
        setLoadingInvitations(false)
      }
    },
    [organizationsApi, setCount],
  )

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

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
      await organizationsApi.acceptOrganizationInvitation(invitation.id)
      toast.success('Invitation accepted successfully')
      await refreshOrganizations(invitation.organizationId)
      await fetchInvitations(false)
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
      await organizationsApi.declineOrganizationInvitation(invitation.id)
      toast.success('Invitation declined successfully')
      await fetchInvitations(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to decline invitation')
      return false
    } finally {
      setLoadingInvitationAction((prev) => ({ ...prev, [invitation.id]: false }))
    }
  }

  return (
    <div className="p-6">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-2xl font-bold">Invitations</h1>
      </div>

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
    </div>
  )
}

export default UserOrganizationInvitations
