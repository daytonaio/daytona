/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useAuth } from 'react-oidc-context'
import { ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { suspend } from 'suspend-react'
import { Organization, OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { useApi } from '@/hooks/useApi'
import { ISelectedOrganizationContext, SelectedOrganizationContext } from '@/contexts/SelectedOrganizationContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { useOrganizations } from '@/hooks/useOrganizations'
import { usePostHog } from 'posthog-js/react'
import { handleApiError } from '@/lib/error-handling'

type Props = {
  children: ReactNode
}

export function SelectedOrganizationProvider(props: Props) {
  const { user } = useAuth()
  const { organizationsApi } = useApi()
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

  const getOrganizationMembers = useCallback(
    async (selectedOrganizationId: string | null) => {
      if (!selectedOrganizationId) {
        return []
      }
      try {
        return (await organizationsApi.listOrganizationMembers(selectedOrganizationId)).data
      } catch (error) {
        handleApiError(error, 'Failed to fetch organization members')
        throw error
      }
    },
    [organizationsApi],
  )

  const [organizationMembers, setOrganizationMembers] = useState(
    suspend(() => getOrganizationMembers(selectedOrganizationId), [organizationsApi, 'organizationMembers']),
  )

  const refreshOrganizationMembers = useCallback(
    async (organizationId?: string) => {
      const organizationMembers = await getOrganizationMembers(organizationId || selectedOrganizationId)
      setOrganizationMembers(organizationMembers)
      return organizationMembers
    },
    [getOrganizationMembers, selectedOrganizationId],
  )

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
      const organizationMembers = await refreshOrganizationMembers(organizationId)

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
    [refreshOrganizationMembers, user],
  )

  const contextValue: ISelectedOrganizationContext = useMemo(() => {
    return {
      selectedOrganization,
      organizationMembers,
      refreshOrganizationMembers,
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
    refreshOrganizationMembers,
  ])

  return (
    <SelectedOrganizationContext.Provider value={contextValue}>{props.children}</SelectedOrganizationContext.Provider>
  )
}
