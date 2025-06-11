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

  // Reset onboarding when switching organizations
  useEffect(() => {
    if (selectedOrganization) {
      setCreatedApiKey(null)
      setHasSufficientPermissions(false)
      setApiKeyPermissions([])
    }
  }, [selectedOrganization])

  // User must have permission to create sandboxes to use the onboarding snippet
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
    if (!selectedOrganization) {
      return
    }

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
            <div className="flex items-center space-x-2">
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
          </div>

          <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-[15px] top-[40px] bottom-0 w-[2px] bg-muted-foreground/50" />

            {/* Steps */}
            <div className="space-y-12">
              {/* Step 1 */}
              <div className="relative pl-12">
                <div className="absolute left-0 w-8 h-8 text-background rounded-full bg-muted-foreground flex items-center justify-center text-sm">
                  1
                </div>
                <div>
                  <h2 className="text-xl font-semibold mb-4">Install the SDK</h2>
                  <p className="mb-4">Run the following command in your terminal to install the Daytona SDK:</p>
                  <div className="transition-all duration-500">
                    <CodeBlock code={codeExamples[language].install} language="bash" showCopy />
                  </div>
                </div>
              </div>

              {/* Step 2 */}
              <div className="relative pl-12">
                <div className="absolute left-0 w-8 h-8 text-background rounded-full bg-muted-foreground flex items-center justify-center text-sm">
                  2
                </div>
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
                      Keys
                    </button>{' '}
                    page.
                  </p>
                  {createdApiKey ? (
                    <div className="p-4 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span className="overflow-x-auto pr-2 cursor-text select-all">
                        {isApiKeyRevealed ? createdApiKey.value : getMaskedApiKey(createdApiKey.value)}
                      </span>
                      <div className="flex items-center space-x-3 pl-3">
                        {isApiKeyRevealed ? (
                          <EyeOff
                            className="w-4 h-4 cursor-pointer hover:text-green-400 dark:hover:text-green-200 transition-colors"
                            onClick={() => setIsApiKeyRevealed(false)}
                          />
                        ) : (
                          <Eye
                            className="w-4 h-4 cursor-pointer hover:text-green-400 dark:hover:text-green-200 transition-colors"
                            onClick={() => setIsApiKeyRevealed(true)}
                          />
                        )}
                        {isApiKeyCopied ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <ClipboardIcon
                            className="w-4 h-4 cursor-pointer hover:text-green-400 dark:hover:text-green-200 transition-colors"
                            onClick={() => copyToClipboard(createdApiKey.value)}
                          />
                        )}
                      </div>
                    </div>
                  ) : (
                    <form
                      onSubmit={async (e) => {
                        e.preventDefault()
                        await handleCreateApiKey()
                      }}
                    >
                      <Input
                        id="key-name"
                        type="text"
                        value={apiKeyName}
                        onChange={(e) => setApiKeyName(e.target.value)}
                        required
                        placeholder="e.g. 'Onboarding'"
                        className="mb-6 md:text-base px-4 h-10.5"
                        disabled={!hasSufficientPermissions}
                      />
                      <Button
                        type="submit"
                        disabled={isLoadingCreateKey || !hasSufficientPermissions}
                        className="text-base"
                      >
                        {isLoadingCreateKey ? (
                          <Loader2 className="h-6 w-6 animate-spin" />
                        ) : (
                          <Plus className="w-6 h-6" />
                        )}
                        Create API Key
                      </Button>
                    </form>
                  )}
                </div>
              </div>

              {/* Step 3 */}
              <div className="relative pl-12">
                <div
                  className={`absolute left-0 w-8 h-8 text-background rounded-full flex items-center justify-center text-sm ${
                    !createdApiKey ? 'bg-secondary' : 'bg-muted-foreground'
                  }`}
                >
                  3
                </div>
                <div className={!createdApiKey ? 'opacity-40 pointer-events-none' : ''}>
                  <h2 className="text-xl font-semibold mb-4">Create a Sandbox</h2>
                  <p className="mb-4">The example below will create a Sandbox and run a simple code snippet:</p>
                  <div className="transition-all duration-500">
                    <CodeBlock
                      code={
                        createdApiKey && isApiKeyRevealed
                          ? codeExamples[language].example.replace('your-api-key', createdApiKey.value)
                          : codeExamples[language].example
                      }
                      language={language}
                      showCopy
                    />
                  </div>
                </div>
              </div>

              {/* Step 4 */}
              <div className="relative pl-12">
                <div
                  className={`absolute left-0 w-8 h-8 text-background rounded-full flex items-center justify-center text-sm ${
                    !createdApiKey ? 'bg-secondary' : 'bg-muted-foreground'
                  }`}
                >
                  4
                </div>
                <div className={!createdApiKey ? 'opacity-40 pointer-events-none' : ''}>
                  <h2 className="text-xl font-semibold mb-4">Run the Example</h2>
                  <p className="mb-4">Run the following command in your terminal to run the example:</p>
                  <div className="transition-all duration-500">
                    <CodeBlock code={codeExamples[language].run} language="bash" showCopy />
                  </div>
                </div>
              </div>

              {/* Step 5 */}
              <div className="relative pl-12">
                <div
                  className={`absolute left-0 w-8 h-8 text-background rounded-full flex items-center justify-center text-sm ${
                    !createdApiKey ? 'bg-secondary' : 'bg-muted-foreground'
                  }`}
                >
                  5
                </div>
                <div className={!createdApiKey ? 'opacity-40 pointer-events-none' : ''}>
                  <h2 className="text-xl font-semibold mb-4">That's It</h2>
                  <p className="text-muted-foreground">
                    It's as easy as that. For more examples check out the{' '}
                    <a href={DAYTONA_DOCS_URL} target="_blank" rel="noopener noreferrer" className="text-primary">
                      Docs
                    </a>
                    .
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

const codeExamples = {
  typescript: {
    install: `npm install @daytonaio/sdk`,
    run: `npx tsx index.mts`,
    example: `import { Daytona } from '@daytonaio/sdk'
  
// Initialize the Daytona client
const daytona = new Daytona({ apiKey: 'your-api-key' });

// Create the Sandbox instance
const sandbox = await daytona.create({
  language: 'typescript',
});

// Run the code securely inside the Sandbox
const response = await sandbox.process.codeRun('console.log("Hello World from code!")')
console.log(response.result);
  `,
  },
  python: {
    install: `pip install daytona`,
    run: `python main.py`,
    example: `from daytona import Daytona, DaytonaConfig
  
# Define the configuration
config = DaytonaConfig(api_key="your-api-key")

# Initialize the Daytona client
daytona = Daytona(config)

# Create the Sandbox instance
sandbox = daytona.create()

# Run the code securely inside the Sandbox
response = sandbox.process.code_run('print("Hello World from code!")')
if response.exit_code != 0:
  print(f"Error: {response.exit_code} {response.result}")
else:
    print(response.result)
  `,
  },
}

export default Onboarding
