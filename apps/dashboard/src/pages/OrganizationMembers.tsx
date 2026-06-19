/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { OrganizationInvitationTable } from '@/components/OrganizationMembers/OrganizationInvitationTable'
import { OrganizationMemberTable } from '@/components/OrganizationMembers/OrganizationMemberTable'
import { UpsertOrganizationAccessSheet } from '@/components/OrganizationMembers/UpsertOrganizationAccessSheet'
import { CreateOrganizationSheet } from '@/components/Organizations/CreateOrganizationSheet'
import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { mutationKeys } from '@/hooks/mutations/mutationKeys'
import { useCancelOrganizationInvitationMutation } from '@/hooks/mutations/useCancelOrganizationInvitationMutation'
import { useCreateOrganizationInvitationMutation } from '@/hooks/mutations/useCreateOrganizationInvitationMutation'
import { useDeleteOrganizationMemberMutation } from '@/hooks/mutations/useDeleteOrganizationMemberMutation'
import { useUpdateOrganizationInvitationMutation } from '@/hooks/mutations/useUpdateOrganizationInvitationMutation'
import { useUpdateOrganizationMemberAccessMutation } from '@/hooks/mutations/useUpdateOrganizationMemberAccessMutation'
import { useOrganizationInvitationsQuery } from '@/hooks/queries/useOrganizationInvitationsQuery'
import { useOrganizationMembersQuery } from '@/hooks/queries/useOrganizationMembersQuery'
import { useSharedRegionsQuery } from '@/hooks/queries/useRegionsQuery'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { usePendingMutationKeys } from '@/hooks/usePendingMutationKeys'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { EMPTY_REGIONS } from '@/lib/regions'
import {
  CreateOrganizationInvitationRoleEnum,
  type Organization,
  OrganizationUserRoleEnum,
  UpdateOrganizationInvitationRoleEnum,
} from '@daytona/api-client'
import { AlertCircle, Building2, PlusIcon, RefreshCw } from 'lucide-react'
import React, { useMemo, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

function usePendingMemberIds() {
  return usePendingMutationKeys<string, { userId?: string }>([
    {
      mutationKey: mutationKeys.organization.members.all,
      getKey: (variables) => variables?.userId,
    },
  ])
}

function usePendingInvitationIds() {
  return usePendingMutationKeys<string, { invitationId?: string }>([
    {
      mutationKey: mutationKeys.organization.invitations.all,
      getKey: (variables) => variables?.invitationId,
    },
  ])
}

const OrganizationMembers: React.FC = () => {
  const { user } = useAuth()

  const { organizationsApi } = useApi()
  const { refreshOrganizations } = useOrganizations()
  const { data: regions = EMPTY_REGIONS, isLoading: loadingRegions } = useSharedRegionsQuery()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const isPersonalOrganization = !!selectedOrganization?.personal
  const { data: organizationMembers = [], isLoading: loadingMembers } = useOrganizationMembersQuery(
    isPersonalOrganization ? null : selectedOrganization?.id,
  )

  const {
    data: invitations = [],
    isLoading: loadingInvitations,
    isError: invitationsError,
    refetch: refetchInvitations,
  } = useOrganizationInvitationsQuery({ enabled: !isPersonalOrganization })
  const updateMemberAccessMutation = useUpdateOrganizationMemberAccessMutation()
  const removeMemberMutation = useDeleteOrganizationMemberMutation()
  const createInvitationMutation = useCreateOrganizationInvitationMutation()
  const updateInvitationMutation = useUpdateOrganizationInvitationMutation()
  const cancelInvitationMutation = useCancelOrganizationInvitationMutation()
  const createInvitationSheetRef = useRef<{ open: () => void }>(null)
  const createOrganizationSheetRef = useRef<{ open: () => void }>(null)

  const pendingMemberIds = usePendingMemberIds()
  const pendingInvitationIds = usePendingInvitationIds()

  const handleUpdateMemberAccess = async (
    userId: string,
    role: OrganizationUserRoleEnum,
    assignedRoleIds: string[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await updateMemberAccessMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        userId,
        access: { role, assignedRoleIds },
      })
      toast.success('Access updated successfully')
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update access')
      return false
    }
  }

  const handleRemoveMember = async (userId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await removeMemberMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        userId,
      })
      toast.success('Member removed successfully')
      if (userId === user?.profile.sub) {
        await refreshOrganizations()
      }
      return true
    } catch (error) {
      handleApiError(error, 'Failed to remove member')
      return false
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
      await createInvitationMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        invitation: { email, role, assignedRoleIds },
      })
      toast.success('Invitation created successfully')
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
    try {
      await updateInvitationMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        invitationId,
        invitation: { role, assignedRoleIds },
      })
      toast.success('Invitation updated successfully')
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update invitation')
      return false
    }
  }

  const handleCancelInvitation = async (invitationId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await cancelInvitationMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        invitationId,
      })
      toast.success('Invitation cancelled successfully')
      return true
    } catch (error) {
      handleApiError(error, 'Failed to cancel invitation')
      return false
    }
  }

  const authenticatedUserIsOwner = authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER
  const canInviteMembers = authenticatedUserIsOwner && !isPersonalOrganization

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!canInviteMembers) {
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
  }, [canInviteMembers])

  useRegisterCommands(rootCommands, { groupId: 'member-actions', groupLabel: 'Member actions', groupOrder: 0 })

  const handleCreateOrganization = async (name: string, defaultRegionId: string): Promise<Organization | null> => {
    try {
      const organization = (
        await organizationsApi.createOrganization({
          name: name.trim(),
          defaultRegionId,
        })
      ).data
      toast.success('Organization created successfully')
      await refreshOrganizations(organization.id)
      return organization
    } catch (error) {
      handleApiError(error, 'Failed to create organization')
      return null
    }
  }

  return (
    <PageLayout>
      <PageHeader />

      <PageContent>
        <PageIntro
          title="Members"
          actions={
            canInviteMembers ? (
              <UpsertOrganizationAccessSheet
                mode="create"
                onSubmit={({ email, role, assignedRoleIds }) => handleCreateInvitation(email, role, assignedRoleIds)}
                ref={createInvitationSheetRef}
              />
            ) : undefined
          }
        />
        {isPersonalOrganization ? (
          <>
            <Empty className="flex-none py-12 rounded-md border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Building2 />
                </EmptyMedia>
                <EmptyTitle>Organizations support member invitations</EmptyTitle>
                <EmptyDescription>
                  Personal accounts cannot invite members. Create an organization to collaborate with other users.
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button variant="secondary" size="sm" onClick={() => createOrganizationSheetRef.current?.open()}>
                  <PlusIcon />
                  Create Organization
                </Button>
              </EmptyContent>
            </Empty>
            <CreateOrganizationSheet
              ref={createOrganizationSheetRef}
              regions={regions}
              loadingRegions={loadingRegions}
              onCreateOrganization={handleCreateOrganization}
            />
          </>
        ) : (
          <div className="flex flex-col gap-14">
            <OrganizationMemberTable
              data={organizationMembers}
              loadingData={loadingMembers}
              onUpdateMemberAccess={handleUpdateMemberAccess}
              onRemoveMember={handleRemoveMember}
              pendingMemberIds={pendingMemberIds}
              ownerMode={authenticatedUserIsOwner}
              currentUserId={user?.profile.sub}
            />

            {authenticatedUserIsOwner && (
              <div>
                <h1 className="text-2xl font-medium mb-3">Invitations</h1>

                {invitationsError ? (
                  <Empty className="py-12 rounded-md border">
                    <EmptyHeader>
                      <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                        <AlertCircle />
                      </EmptyMedia>
                      <EmptyTitle className="text-destructive">Failed to load invitations</EmptyTitle>
                      <EmptyDescription>
                        Something went wrong while fetching organization invitations. Please try again.
                      </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                      <Button variant="secondary" size="sm" onClick={() => refetchInvitations()}>
                        <RefreshCw />
                        Retry
                      </Button>
                    </EmptyContent>
                  </Empty>
                ) : (
                  <OrganizationInvitationTable
                    data={invitations}
                    loadingData={loadingInvitations}
                    onCancelInvitation={handleCancelInvitation}
                    onUpdateInvitation={handleUpdateInvitation}
                    pendingInvitationIds={pendingInvitationIds}
                  />
                )}
              </div>
            )}
          </div>
        )}
      </PageContent>
    </PageLayout>
  )
}

export default OrganizationMembers
