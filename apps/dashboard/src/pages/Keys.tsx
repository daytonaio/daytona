/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateApiKeyDialog } from '@/components/CreateApiKeyDialog'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import {
  ApiKeyList,
  ApiKeyResponse,
  CreateApiKeyPermissionsEnum,
  OrganizationUserRoleEnum,
} from '@daytonaio/api-client'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { ApiKeyTable } from '../components/ApiKeyTable'

const Keys: React.FC = () => {
  const { apiKeyApi } = useApi()
  const { apiUrl } = useConfig()
  const [keys, setKeys] = useState<ApiKeyList[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingKeys, setLoadingKeys] = useState<Record<string, boolean>>({})

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const availablePermissions = useMemo<CreateApiKeyPermissionsEnum[]>(() => {
    if (!authenticatedUserOrganizationMember) {
      return []
    }
    if (authenticatedUserOrganizationMember.role === OrganizationUserRoleEnum.OWNER) {
      return Object.values(CreateApiKeyPermissionsEnum)
    }
    return Array.from(new Set(authenticatedUserOrganizationMember.assignedRoles.flatMap((role) => role.permissions)))
  }, [authenticatedUserOrganizationMember])

  const fetchKeys = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoading(true)
      }
      try {
        const response = await apiKeyApi.listApiKeys(selectedOrganization.id)
        setKeys(response.data)
      } catch (error) {
        handleApiError(error, 'Failed to fetch API keys')
      } finally {
        setLoading(false)
      }
    },
    [apiKeyApi, selectedOrganization],
  )

  useEffect(() => {
    fetchKeys()
  }, [fetchKeys])

  const handleRevoke = async (key: ApiKeyList) => {
    const loadingId = getLoadingKeyId(key)
    setLoadingKeys((prev) => ({ ...prev, [loadingId]: true }))
    try {
      await apiKeyApi.deleteApiKeyForUser(key.userId, key.name, selectedOrganization?.id)
      toast.success('API key revoked successfully')
      await fetchKeys(false)
    } catch (error) {
      handleApiError(error, 'Failed to revoke API key')
    } finally {
      setLoadingKeys((prev) => ({ ...prev, [loadingId]: false }))
    }
  }

  const handleCreateKey = async (
    name: string,
    permissions: CreateApiKeyPermissionsEnum[],
    expiresAt: Date | null,
  ): Promise<ApiKeyResponse | null> => {
    try {
      const key = (await apiKeyApi.createApiKey({ name, permissions, expiresAt }, selectedOrganization?.id)).data
      toast.success('API key created successfully')
      await fetchKeys(false)
      return key
    } catch (error) {
      handleApiError(error, 'Failed to create API key')
      return null
    }
  }

  const getLoadingKeyId = useCallback((key: ApiKeyList) => {
    return `${key.userId}-${key.name}`
  }, [])

  const isLoadingKey = useCallback(
    (key: ApiKeyList) => {
      const loadingId = getLoadingKeyId(key)
      return loadingKeys[loadingId]
    },
    [getLoadingKeyId, loadingKeys],
  )

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>API Keys</PageTitle>
        <CreateApiKeyDialog
          className="ml-auto"
          availablePermissions={availablePermissions}
          onCreateApiKey={handleCreateKey}
          apiUrl={apiUrl}
        />
      </PageHeader>

      <PageContent size="full">
        <ApiKeyTable data={keys} loading={loading} isLoadingKey={isLoadingKey} onRevoke={handleRevoke} />
      </PageContent>
    </PageLayout>
  )
}

export default Keys
