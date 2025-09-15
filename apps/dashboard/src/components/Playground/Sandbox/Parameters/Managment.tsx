/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CodeLanguage, Resources, CreateSandboxBaseParams } from '@daytonaio/sdk-typescript/src'
import { ApiKeyList } from '@daytonaio/api-client'
import { usePlaygroundSandboxParams } from '../hook'
import { Loader2 } from 'lucide-react'
import { useState, useEffect } from 'react'

type SandboxManagmentParametersProps = {
  apiKeys: ApiKeyList[]
  apiKeysLoading: boolean
}

const SandboxManagmentParameters: React.FC<SandboxManagmentParametersProps> = ({ apiKeys, apiKeysLoading }) => {
  const { playgroundSandboxParametersState, setPlaygroundSandboxParameterValue } = usePlaygroundSandboxParams()
  const [sandboxApiKey, setSandboxApiKey] = useState<string | undefined>(playgroundSandboxParametersState['apiKey']) //*AKO NE POSTOJI NIJEDAN -> VRIJEDNOST NEK BUDE default
  const [sandboxLanguage, setSandboxLanguage] = useState<CodeLanguage | undefined>(
    playgroundSandboxParametersState['language'],
  )
  const [resources, setResources] = useState<Resources>({
    cpu: 2,
    gpu: 0,
    memory: 4,
    disk: 8,
  })
  const [sandboxFromImageParams, setSandboxFromImageParams] = useState<CreateSandboxBaseParams>({
    autoStopInterval: 15,
    autoArchiveInterval: 7,
    autoDeleteInterval: -1,
  })
  // Available languages
  const languageOptions = [
    {
      value: CodeLanguage.PYTHON,
      label: 'Python',
    },
    {
      value: CodeLanguage.TYPESCRIPT,
      label: 'TypeScript',
    },
    {
      value: CodeLanguage.JAVASCRIPT,
      label: 'JavaScript',
    },
  ]
  useEffect(() => {
    if (!apiKeysLoading && !apiKeys.length) setSandboxApiKey('default') // If no available keys set to default value
  }, [apiKeysLoading, apiKeys])
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="api_key">API key</Label>
        <Select
          value={sandboxApiKey}
          onValueChange={(value) => {
            setSandboxApiKey(value)
            setPlaygroundSandboxParameterValue('apiKey', value)
          }}
        >
          <SelectTrigger className="w-full rounded-lg" aria-label="Select API key">
            {apiKeysLoading ? (
              <div className="w-full flex items-center justify-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                <span className="text-muted-foreground">Loading...</span>
              </div>
            ) : (
              <SelectValue id="api_key" placeholder="API key" />
            )}
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            {apiKeys.map((key) => (
              <SelectItem key={key.value} value={key.value}>
                {key.value}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label htmlFor="language">Language</Label>
        <Select
          value={sandboxLanguage}
          onValueChange={(value) => {
            setSandboxLanguage(value as CodeLanguage)
            setPlaygroundSandboxParameterValue('language', value as CodeLanguage)
          }}
        >
          <SelectTrigger className="w-full box-border rounded-lg" aria-label="Select sandbox language">
            <SelectValue id="language" placeholder="Language" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            {languageOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </>
  )
}

export default SandboxManagmentParameters
