/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { useAuth } from 'react-oidc-context'
import { useApi } from '@/hooks/useApi'
import {
  CreateOrganizationInvitationRoleEnum,
  OrganizationInvitation,
  OrganizationUserRoleEnum,
  UpdateOrganizationInvitationRoleEnum,
} from '@daytonaio/api-client'
import { OrganizationMemberTable } from '@/components/OrganizationMembers/OrganizationMemberTable'
import { CreateOrganizationInvitationDialog } from '@/components/OrganizationMembers/CreateOrganizationInvitationDialog'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { OrganizationInvitationTable } from '@/components/OrganizationMembers/OrganizationInvitationTable'
import { useOrganizationRoles } from '@/hooks/useOrganizationRoles'
import { useOrganizations } from '@/hooks/useOrganizations'
import { handleApiError } from '@/lib/error-handling'

const OrganizationMembers: React.FC = () => {
  const { user } = useAuth()
  const { organizationsApi } = useApi()

  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, organizationMembers, refreshOrganizationMembers, authenticatedUserOrganizationMember } =
    useSelectedOrganization()
  const { roles, loadingRoles } = useOrganizationRoles()

  const [invitations, setInvitations] = useState<OrganizationInvitation[]>([])
  const [loadingInvitations, setLoadingInvitations] = useState(true)

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

  const handleUpdateMemberRole = async (userId: string, role: OrganizationUserRoleEnum): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingMemberAction((prev) => ({ ...prev, [userId]: true }))
    try {
      await organizationsApi.updateRoleForOrganizationMember(selectedOrganization.id, userId, { role })
      toast.success('Role updated successfully')
      await refreshOrganizationMembers()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update member role')
      return false
    } finally {
      setLoadingMemberAction((prev) => ({ ...prev, [userId]: false }))
    }
  }

  const handleUpdateAssignedOrganizationRoles = async (userId: string, roleIds: string[]): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingMemberAction((prev) => ({ ...prev, [userId]: true }))
    try {
      await organizationsApi.updateAssignedOrganizationRoles(selectedOrganization.id, userId, { roleIds })
      toast.success('Assignments updated successfully')
      await refreshOrganizationMembers()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update assignments')
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

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Members</h1>
        {authenticatedUserIsOwner && (
          <CreateOrganizationInvitationDialog
            availableRoles={roles}
            loadingAvailableRoles={loadingRoles}
            onCreateInvitation={handleCreateInvitation}
          />
        )}
      </div>

      <OrganizationMemberTable
        data={organizationMembers}
        loadingData={loadingRoles}
        availableOrganizationRoles={roles}
        loadingAvailableOrganizationRoles={loadingRoles}
        onUpdateMemberRole={handleUpdateMemberRole}
        onUpdateAssignedOrganizationRoles={handleUpdateAssignedOrganizationRoles}
        onRemoveMember={handleRemoveMember}
        loadingMemberAction={loadingMemberAction}
        ownerMode={authenticatedUserIsOwner}
      />

      {authenticatedUserIsOwner && (
        <>
          <div className="mb-2 mt-12 h-12 flex items-center justify-between">
            <h1 className="text-2xl font-medium">Invitations</h1>
          </div>

          <OrganizationInvitationTable
            data={invitations}
            loadingData={loadingInvitations}
            availableRoles={roles}
            loadingAvailableRoles={loadingRoles}
            onCancelInvitation={handleCancelInvitation}
            onUpdateInvitation={handleUpdateInvitation}
            loadingInvitationAction={loadingInvitationAction}
          />
        </>
      )}
    </div>
  )
}

export default OrganizationMembers
