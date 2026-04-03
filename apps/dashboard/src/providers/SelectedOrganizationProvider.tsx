/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ISelectedOrganizationContext, SelectedOrganizationContext } from '@/contexts/SelectedOrganizationContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import {
  getOrganizationMembersQueryOptions,
  useOrganizationMembersSuspenseQuery,
} from '@/hooks/queries/useOrganizationMembersQuery'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { Organization, OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { usePostHog } from 'posthog-js/react'
import { ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

type Props = {
  children: ReactNode
}

export function SelectedOrganizationProvider(props: Props) {
  const { user } = useAuth()
  const { organizationsApi } = useApi()
  const queryClient = useQueryClient()
  const posthog = usePostHog()

  const { organizations } = useOrganizations()

  const [selectedOrganizationId, setSelectedOrganizationId] = useState<string | null>(() => {
    const storedId = localStorage.getItem(LocalStorageKey.SelectedOrganizationId)
    if (storedId && organizations.find((org) => org.id === storedId)) {
      return storedId
    } else if (organizations.length > 0) {
      const defaultOrg = organizations.find((org) => org.personal) || organizations[0]
      localStorage.setItem(LocalStorageKey.SelectedOrganizationId, defaultOrg.id)
      return defaultOrg.id
    } else {
      localStorage.removeItem(LocalStorageKey.SelectedOrganizationId)
      return null
    }
  })

  useEffect(() => {
    if (!organizations.length) {
      setSelectedOrganizationId(null)
    }
    if (!selectedOrganizationId || !organizations.some((org) => org.id === selectedOrganizationId)) {
      const defaultOrg = organizations.find((org) => org.personal) || organizations[0]
      localStorage.setItem(LocalStorageKey.SelectedOrganizationId, defaultOrg.id)
      setSelectedOrganizationId(defaultOrg.id)
    }
  }, [organizations, selectedOrganizationId])

  const selectedOrganization = useMemo<Organization | null>(() => {
    if (!selectedOrganizationId) {
      return null
    }
    return organizations.find((org) => org.id === selectedOrganizationId) || null
  }, [organizations, selectedOrganizationId])

  useEffect(() => {
    if (!posthog || !selectedOrganizationId) {
      return
    }

    posthog.group('organization', selectedOrganizationId)
  }, [posthog, selectedOrganizationId])

  const { data: organizationMembers } = useOrganizationMembersSuspenseQuery(selectedOrganizationId)

  const authenticatedUserOrganizationMember = useMemo(() => {
    return organizationMembers.find((member) => member.userId === user?.profile.sub) || null
  }, [organizationMembers, user])

  const authenticatedUserAssignedPermissions = useMemo(() => {
    if (!authenticatedUserOrganizationMember) {
      return null
    }
    return new Set(authenticatedUserOrganizationMember.assignedRoles.flatMap((role) => role.permissions))
  }, [authenticatedUserOrganizationMember])

  const authenticatedUserHasPermission = useCallback(
    (permission: OrganizationRolePermissionsEnum) => {
      if (!authenticatedUserOrganizationMember || !authenticatedUserAssignedPermissions) {
        return false
      }
      if (authenticatedUserOrganizationMember.role === OrganizationUserRoleEnum.OWNER) {
        return true
      }
      return authenticatedUserAssignedPermissions.has(permission)
    },
    [authenticatedUserOrganizationMember, authenticatedUserAssignedPermissions],
  )

  const handleSelectOrganization = useCallback(
    async (organizationId: string): Promise<boolean> => {
      const organizationMembers = await queryClient.fetchQuery(
        getOrganizationMembersQueryOptions(organizationsApi, organizationId),
      )

      // confirm switch if user is a member of the new organization
      if (organizationMembers.some((member) => member.userId === user?.profile.sub)) {
        localStorage.setItem(LocalStorageKey.SelectedOrganizationId, organizationId)
        setSelectedOrganizationId(organizationId)
        return true
      } else {
        toast.error('Failed to switch organization', {
          closeButton: true,
        })
        return false
      }
    },
    [organizationsApi, queryClient, user],
  )

  const contextValue: ISelectedOrganizationContext = useMemo(() => {
    return {
      selectedOrganization,
      organizationMembers,
      authenticatedUserOrganizationMember,
      authenticatedUserHasPermission,
      onSelectOrganization: handleSelectOrganization,
    }
  }, [
    selectedOrganization,
    organizationMembers,
    authenticatedUserOrganizationMember,
    authenticatedUserHasPermission,
    handleSelectOrganization,
  ])

  return (
    <SelectedOrganizationContext.Provider value={contextValue}>{props.children}</SelectedOrganizationContext.Provider>
  )
}
