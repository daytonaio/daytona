/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui/accordion'
import { SandboxParametersSections, sandboxParametersSectionsData } from '@/enums/Playground'
import { ApiKeyList } from '@daytonaio/api-client'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import SandboxManagmentParameters from './Managment'
import { Plus, Minus } from 'lucide-react'
import { useState, useEffect, useCallback } from 'react'

const SandboxParameters: React.FC = () => {
  const [openedParametersSections, setOpenedParametersSections] = useState<SandboxParametersSections[]>([
    SandboxParametersSections.SANDBOX_MANAGMENT,
  ])

  // Available API keys -> fetch here instead of SandboxManagmentParameters to prevent fetch on every accordion open/close
  const [apiKeys, setApiKeys] = useState<ApiKeyList[]>([])
  const [apiKeysLoading, setApiKeysLoading] = useState(true)
  const { apiKeyApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const fetchKeys = useCallback(async () => {
    if (!selectedOrganization) return
    setApiKeysLoading(true)
    try {
      const response = await apiKeyApi.listApiKeys(selectedOrganization.id)
      setApiKeys(response.data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch API keys')
    } finally {
      setApiKeysLoading(false)
    }
  }, [apiKeyApi, selectedOrganization])

  useEffect(() => {
    fetchKeys()
  }, [fetchKeys])

  return (
    <div className="flex flex-col space-y-2">
      <Accordion
        type="multiple"
        value={openedParametersSections}
        onValueChange={(parametersSections) =>
          setOpenedParametersSections(parametersSections as SandboxParametersSections[])
        }
      >
        {sandboxParametersSectionsData.map((section) => {
          const isCollapsed = !openedParametersSections.includes(section.value as SandboxParametersSections)
          return (
            <AccordionItem key={section.value} value={section.value}>
              <AccordionTrigger className="text-lg" icon={isCollapsed ? <Plus /> : <Minus />}>
                {section.label}
              </AccordionTrigger>
              <AccordionContent>
                {!isCollapsed && (
                  <div className="px-2 space-y-4">
                    {section.value === SandboxParametersSections.SANDBOX_MANAGMENT && (
                      <SandboxManagmentParameters
                        apiKeys={apiKeys.map((apiKey) => ({ ...apiKey, label: apiKey.name }))}
                        apiKeysLoading={apiKeysLoading}
                      />
                    )}
                  </div>
                )}
              </AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
    </div>
  )
}

export default SandboxParameters
