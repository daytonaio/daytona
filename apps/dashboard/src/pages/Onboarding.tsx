/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useState } from 'react'
import { Check, ClipboardIcon, Eye, EyeOff, Loader2, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { useNavigate } from 'react-router-dom'
import { CreateApiKeyPermissionsEnum, ApiKeyResponse, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import pythonIcon from '@/assets/python.svg'
import typescriptIcon from '@/assets/typescript.svg'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import CodeBlock from '@/components/CodeBlock'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { RoutePath } from '@/enums/RoutePath'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { getMaskedApiKey } from '@/lib/utils'

const Onboarding: React.FC = () => {
  const { apiKeyApi } = useApi()
  const { organizations } = useOrganizations()
  const { selectedOrganization, onSelectOrganization, authenticatedUserHasPermission } = useSelectedOrganization()
  const navigate = useNavigate()

  const [language, setLanguage] = useState<'typescript' | 'python'>('python')
  const [apiKeyName, setApiKeyName] = useState('')
  const [apiKeyPermissions, setApiKeyPermissions] = useState<CreateApiKeyPermissionsEnum[]>([])
  const [createdApiKey, setCreatedApiKey] = useState<ApiKeyResponse | null>(null)
  const [isApiKeyRevealed, setIsApiKeyRevealed] = useState(false)
  const [isApiKeyCopied, setIsApiKeyCopied] = useState(false)
  const [isLoadingCreateKey, setIsLoadingCreateKey] = useState(false)
  const [hasSufficientPermissions, setHasSufficientPermissions] = useState(false)

  useEffect(() => {
    if (selectedOrganization) {
      setCreatedApiKey(null)
      setHasSufficientPermissions(false)
      setApiKeyPermissions([])
    }
  }, [selectedOrganization])

  useEffect(() => {
    const ensureOnboardingPermissions = async () => {
      if (authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)) {
        setHasSufficientPermissions(true)
        const permissions: CreateApiKeyPermissionsEnum[] = [CreateApiKeyPermissionsEnum.WRITE_SANDBOXES]
        if (authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)) {
          permissions.push(CreateApiKeyPermissionsEnum.DELETE_SANDBOXES)
        }
        setApiKeyPermissions(permissions)
      } else {
        const personalOrg = organizations.find((org) => org.personal)

        if (personalOrg) {
          const success = await onSelectOrganization(personalOrg.id)
          if (success) {
            toast.success('Switched to personal organization', {
              description:
                'You did not have the necessary permissions for creating sandboxes in the previous organization.',
            })
            return
          }
        }

        toast.error('An unexpected issue occurred while preparing your onboarding snippet')
      }
    }

    ensureOnboardingPermissions()
  }, [authenticatedUserHasPermission, onSelectOrganization, organizations])

  const handleCreateApiKey = async () => {
    if (!selectedOrganization) return

    setIsLoadingCreateKey(true)
    try {
      const key = (
        await apiKeyApi.createApiKey(
          {
            name: apiKeyName,
            permissions: apiKeyPermissions,
          },
          selectedOrganization.id,
        )
      ).data
      setCreatedApiKey(key)
      setApiKeyName('')
      toast.success('API key created successfully')
    } catch (error) {
      handleApiError(error, 'Failed to create API key')
    } finally {
      setIsLoadingCreateKey(false)
    }
  }

  const copyToClipboard = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value)
      setIsApiKeyCopied(true)
      setTimeout(() => setIsApiKeyCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy text:', err)
    }
  }

  return (
    <div className="p-6">
      <div className="min-h-screen p-14">
        <div className="max-w-3xl mx-auto">
          <div className="flex justify-between items-center mb-8">
            <div>
              <h1 className="text-2xl font-bold mb-2">Get Started</h1>
              <p className="text-muted-foreground">Install and get your Sandboxes running.</p>
            </div>

            <Tabs value={language} onValueChange={(value) => setLanguage(value as 'typescript' | 'python')}>
              <TabsList className="bg-foreground/10">
                <TabsTrigger value="python">
                  <img src={pythonIcon} alt="Python" className="w-4 h-4" />
                </TabsTrigger>
                <TabsTrigger value="typescript">
                  <img src={typescriptIcon} alt="TypeScript" className="w-4 h-4" />
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <div className="space-y-12">
            {/* Step 2 */}
            <div>
              <h2 className="text-xl font-semibold mb-4">Create an API Key</h2>
              <p className="mb-4">
                This API key will have permissions to only{' '}
                {apiKeyPermissions.includes(CreateApiKeyPermissionsEnum.DELETE_SANDBOXES) ? 'manage' : 'create'}{' '}
                Sandboxes. For full API permissions, head to the{' '}
                <button
                  onClick={() => navigate(RoutePath.KEYS)}
                  className="underline cursor-pointer hover:text-muted-foreground"
                >
                  API Keys
                </button>{' '}
                page.
              </p>

              {!createdApiKey && (
                <form
                  onSubmit={async (e) => {
                    e.preventDefault()
                    await handleCreateApiKey()
                  }}
                >
                  {/* ✅ Added visible label + helper text */}
                  <label htmlFor="key-name" className="block text-sm font-medium mb-1">
                    API Key Name
                  </label>
                  <p className="text-xs text-muted-foreground mb-2">
                    This name is for your reference only and helps you identify the key later.
                  </p>

                  <Input
                    id="key-name"
                    type="text"
                    value={apiKeyName}
                    onChange={(e) => setApiKeyName(e.target.value)}
                    required
                    placeholder="e.g. Onboarding"
                    className="mb-6 md:text-base px-4 h-10.5"
                    disabled={!hasSufficientPermissions}
                  />

                  <Button type="submit" disabled={isLoadingCreateKey || !hasSufficientPermissions}>
                    {isLoadingCreateKey ? <Loader2 className="animate-spin" /> : <Plus />}
                    Create API Key
                  </Button>
                </form>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Onboarding
