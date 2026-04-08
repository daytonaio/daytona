/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateApiKeySheet } from '@/components/CreateApiKeySheet'
import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { useRevokeApiKeyMutation } from '@/hooks/mutations/useRevokeApiKeyMutation'
import { useApiKeysQuery } from '@/hooks/queries/useApiKeysQuery'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { ApiKeyList, CreateApiKeyPermissionsEnum, OrganizationUserRoleEnum } from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import { useCallback, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { ApiKeyTable } from '../components/ApiKeyTable'

const Keys: React.FC = () => {
  const { apiUrl } = useConfig()
  const [loadingKeys, setLoadingKeys] = useState<Record<string, boolean>>({})
  const createApiKeySheetRef = useRef<{ open: () => void }>(null)

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const revokeApiKeyMutation = useRevokeApiKeyMutation()
  const apiKeysQuery = useApiKeysQuery(selectedOrganization?.id)

  const availablePermissions = useMemo<CreateApiKeyPermissionsEnum[]>(() => {
    if (!authenticatedUserOrganizationMember) {
      return []
    }
    if (authenticatedUserOrganizationMember.role === OrganizationUserRoleEnum.OWNER) {
      return Object.values(CreateApiKeyPermissionsEnum)
    }
    return Array.from(new Set(authenticatedUserOrganizationMember.assignedRoles.flatMap((role) => role.permissions)))
  }, [authenticatedUserOrganizationMember])

  const handleRevoke = async (key: ApiKeyList) => {
    if (!selectedOrganization) {
      return
    }
    const loadingId = getLoadingKeyId(key)
    setLoadingKeys((prev) => ({ ...prev, [loadingId]: true }))
    try {
      await revokeApiKeyMutation.mutateAsync({
        userId: key.userId,
        name: key.name,
        organizationId: selectedOrganization.id,
      })
      toast.success('API key revoked successfully')
    } catch (error) {
      handleApiError(error, 'Failed to revoke API key')
    } finally {
      setLoadingKeys((prev) => ({ ...prev, [loadingId]: false }))
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

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!selectedOrganization?.id) {
      return []
    }

    return [
      {
        id: 'create-key',
        label: 'Create Key',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => createApiKeySheetRef.current?.open(),
      },
    ]
  }, [selectedOrganization?.id])

  useRegisterCommands(rootCommands, { groupId: 'api-key-actions', groupLabel: 'API key actions', groupOrder: 0 })

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>API Keys</PageTitle>
        <CreateApiKeySheet
          className="ml-auto"
          availablePermissions={availablePermissions}
          apiUrl={apiUrl}
          organizationId={selectedOrganization?.id}
          ref={createApiKeySheetRef}
        />
      </PageHeader>

      <PageContent>
        <ApiKeyTable
          data={apiKeysQuery.data ?? []}
          loading={apiKeysQuery.isLoading || apiKeysQuery.isRefetching}
          isLoadingKey={isLoadingKey}
          onRevoke={handleRevoke}
        />
      </PageContent>
    </PageLayout>
  )
}

export default Keys
