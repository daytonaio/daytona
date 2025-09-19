/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CodeLanguage, Resources, CreateSandboxBaseParams } from '@daytonaio/sdk-typescript/src'
import { ApiKeyList } from '@daytonaio/api-client'
import { usePlayground } from '@/hooks/usePlayground'
import { NumberParameterFormItem } from '@/enums/Playground'
import { Loader2 } from 'lucide-react'
import { useState, useEffect } from 'react'

type SandboxManagmentParametersProps = {
  apiKeys: ApiKeyList[]
  apiKeysLoading: boolean
}

const SandboxManagmentParameters: React.FC<SandboxManagmentParametersProps> = ({ apiKeys, apiKeysLoading }) => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const [sandboxApiKey, setSandboxApiKey] = useState<string | undefined>(sandboxParametersState['apiKey']) //*AKO NE POSTOJI NIJEDAN -> VRIJEDNOST NEK BUDE default
  const [sandboxLanguage, setSandboxLanguage] = useState<CodeLanguage | undefined>(sandboxParametersState['language'])
  const [resources, setResources] = useState<Resources>(sandboxParametersState['resources'])
  const [sandboxFromImageParams, setSandboxFromImageParams] = useState<CreateSandboxBaseParams>(
    sandboxParametersState['createSandboxBaseParams'],
  )
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
  const resourcesFormData: (NumberParameterFormItem & { key: keyof Resources })[] = [
    { label: 'Compute (vCPU):', key: 'cpu', min: 1, max: Infinity, placeholder: '1' },
    { label: 'Memory (GiB):', key: 'memory', min: 1, max: Infinity, placeholder: '1' },
    { label: 'Storage (GiB):', key: 'disk', min: 1, max: Infinity, placeholder: '3' },
  ]
  const lifecycleParamsFormData: (NumberParameterFormItem & {
    key: 'autoStopInterval' | 'autoArchiveInterval' | 'autoDeleteInterval'
  })[] = [
    { label: 'Stop (min):', key: 'autoStopInterval', min: 0, max: Infinity, placeholder: '15' },
    { label: 'Archive (min)', key: 'autoArchiveInterval', min: 0, max: Infinity, placeholder: '7' },
    { label: 'Delete (min):', key: 'autoDeleteInterval', min: -1, max: Infinity, placeholder: '' },
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
            setSandboxParameterValue('apiKey', value)
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
            setSandboxParameterValue('language', value as CodeLanguage)
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
      <div className="space-y-2">
        <Label htmlFor="resources">Resources</Label>
        <div id="resources" className="px-4 space-y-2">
          {resourcesFormData.map((resource) => (
            <div key={resource.key} className="flex items-center gap-4">
              <Label htmlFor={resource.key} className="w-32 flex-shrink-0">
                {resource.label}
              </Label>
              <Input
                id={resource.key}
                type="number"
                className="w-full"
                min={resource.min}
                max={resource.max}
                placeholder={resource.placeholder}
                value={resources[resource.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const resourcesNew = { ...resources, [resource.key]: newValue }
                  setResources(resourcesNew)
                  setSandboxParameterValue('resources', resourcesNew)
                }}
              />
            </div>
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="lifecycle">Lifecycle</Label>
        <div id="lifecycle" className="px-4 space-y-2">
          {lifecycleParamsFormData.map((lifecycleParam) => (
            <div key={lifecycleParam.key} className="flex items-center gap-4">
              <Label htmlFor={lifecycleParam.key} className="w-32 flex-shrink-0">
                {lifecycleParam.label}
              </Label>
              <Input
                id={lifecycleParam.key}
                type="number"
                className="w-full"
                min={lifecycleParam.min}
                max={lifecycleParam.max}
                placeholder={lifecycleParam.placeholder}
                value={sandboxFromImageParams[lifecycleParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const sandboxFromImageParamsNew = { ...sandboxFromImageParams, [lifecycleParam.key]: newValue }
                  setSandboxFromImageParams(sandboxFromImageParamsNew)
                  setSandboxParameterValue('createSandboxBaseParams', sandboxFromImageParamsNew)
                }}
              />
            </div>
          ))}
        </div>
      </div>
    </>
  )
}

export default SandboxManagmentParameters
