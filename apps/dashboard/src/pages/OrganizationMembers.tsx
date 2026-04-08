/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { OrganizationInvitationTable } from '@/components/OrganizationMembers/OrganizationInvitationTable'
import { OrganizationMemberTable } from '@/components/OrganizationMembers/OrganizationMemberTable'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import {
  CreateOrganizationInvitationRoleEnum,
  OrganizationInvitation,
  OrganizationUserRoleEnum,
  UpdateOrganizationInvitationRoleEnum,
} from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

const OrganizationMembers: React.FC = () => {
  const { user } = useAuth()
  const { organizationsApi } = useApi()

  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, organizationMembers, refreshOrganizationMembers, authenticatedUserOrganizationMember } =
    useSelectedOrganization()

  const [invitations, setInvitations] = useState<OrganizationInvitation[]>([])
  const [loadingInvitations, setLoadingInvitations] = useState(true)
  const createInvitationSheetRef = useRef<{ open: () => void }>(null)

  const [loadingMemberAction, setLoadingMemberAction] = useState<Record<string, boolean>>({})
  const [loadingInvitationAction, setLoadingInvitationAction] = useState<Record<string, boolean>>({})

  const fetchInvitations = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingInvitations(true)
      }
      try {
        const response = await organizationsApi.listOrganizationInvitations(selectedOrganization.id)
        setInvitations(response.data)
      } catch (error) {
        handleApiError(error, 'Failed to fetch invitations')
      } finally {
        setLoadingInvitations(false)
      }
    },
    [organizationsApi, selectedOrganization],
  )

  useEffect(() => {
    refreshOrganizationMembers()
    fetchInvitations()
  }, [fetchInvitations, refreshOrganizationMembers])

  const handleUpdateMemberAccess = async (
    userId: string,
    role: OrganizationUserRoleEnum,
    assignedRoleIds: string[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingMemberAction((prev) => ({ ...prev, [userId]: true }))
    try {
      await organizationsApi.updateAccessForOrganizationMember(selectedOrganization.id, userId, {
        role,
        assignedRoleIds,
      })
      toast.success('Access updated successfully')
      await refreshOrganizationMembers()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update access')
      return false
    } finally {
      setLoadingMemberAction((prev) => ({ ...prev, [userId]: false }))
    }
  }

  const handleRemoveMember = async (userId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingMemberAction((prev) => ({ ...prev, [userId]: true }))
    try {
      await organizationsApi.deleteOrganizationMember(selectedOrganization.id, userId)
      toast.success('Member removed successfully')
      if (userId === user?.profile.sub) {
        await refreshOrganizations()
      } else {
        await refreshOrganizationMembers()
      }
      return true
    } catch (error) {
      handleApiError(error, 'Failed to remove member')
      return false
    } finally {
      setLoadingMemberAction((prev) => ({ ...prev, [userId]: false }))
    }
  }

  const handleCreateInvitation = async (
    email: string,
    role: CreateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await organizationsApi.createOrganizationInvitation(selectedOrganization.id, { email, role, assignedRoleIds })
      toast.success('Invitation created successfully')
      await fetchInvitations(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to create invitation')
      return false
    }
  }

  const handleUpdateInvitation = async (
    invitationId: string,
    role: UpdateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingInvitationAction((prev) => ({ ...prev, [invitationId]: true }))
    try {
      await organizationsApi.updateOrganizationInvitation(selectedOrganization.id, invitationId, {
        role,
        assignedRoleIds,
      })
      toast.success('Invitation updated successfully')
      await fetchInvitations(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update invitation')
      return false
    } finally {
      setLoadingInvitationAction((prev) => ({ ...prev, [invitationId]: false }))
    }
  }

  const handleCancelInvitation = async (invitationId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingInvitationAction((prev) => ({ ...prev, [invitationId]: true }))
    try {
      await organizationsApi.cancelOrganizationInvitation(selectedOrganization.id, invitationId)
      toast.success('Invitation cancelled successfully')
      await fetchInvitations(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to cancel invitation')
      return false
    } finally {
      setLoadingInvitationAction((prev) => ({ ...prev, [invitationId]: false }))
    }
  }

  const authenticatedUserIsOwner = useMemo(() => {
    return authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER
  }, [authenticatedUserOrganizationMember])

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!authenticatedUserIsOwner) {
      return []
    }

    return [
      {
        id: 'create-organization-invitation',
        label: 'Invite Member',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => createInvitationSheetRef.current?.open(),
      },
    ]
  }, [authenticatedUserIsOwner])

  useRegisterCommands(rootCommands, { groupId: 'member-actions', groupLabel: 'Member actions', groupOrder: 0 })

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Members</PageTitle>
        {authenticatedUserIsOwner && (
          <UpsertOrganizationAccessSheet
            mode="create"
            className="ml-auto"
            onSubmit={({ email, role, assignedRoleIds }) => handleCreateInvitation(email, role, assignedRoleIds)}
            ref={createInvitationSheetRef}
          />
        )}
      </PageHeader>

      <PageContent>
        <OrganizationMemberTable
          data={organizationMembers}
          loadingData={false}
          onUpdateMemberAccess={handleUpdateMemberAccess}
          onRemoveMember={handleRemoveMember}
          loadingMemberAction={loadingMemberAction}
          ownerMode={authenticatedUserIsOwner}
          currentUserId={user?.profile.sub}
        />

        {authenticatedUserIsOwner && (
          <>
            <div className="mb-2 mt-12 h-12 flex items-center justify-between">
              <h1 className="text-2xl font-medium">Invitations</h1>
            </div>

            <OrganizationInvitationTable
              data={invitations}
              loadingData={loadingInvitations}
              onCancelInvitation={handleCancelInvitation}
              onUpdateInvitation={handleUpdateInvitation}
              loadingInvitationAction={loadingInvitationAction}
            />
          </>
        )}
      </PageContent>
    </PageLayout>
  )
}

export default OrganizationMembers
