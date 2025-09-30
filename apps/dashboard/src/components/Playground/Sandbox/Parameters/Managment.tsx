/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'
import FormSelectInput from '../../Inputs/SelectInput'
import FormNumberInput from '../../Inputs/NumberInput'
import { CodeLanguage, Resources, CreateSandboxBaseParams } from '@daytonaio/sdk-typescript/src'
import { ApiKeyList } from '@daytonaio/api-client'
import { usePlayground } from '@/hooks/usePlayground'
import { NumberParameterFormItem, ParameterFormItem } from '@/enums/Playground'
import { useState, useEffect } from 'react'

type SandboxManagmentParametersProps = {
  apiKeys: (ApiKeyList & { label: string })[] // For FormSelectInput selectOptions prop compatibility
  apiKeysLoading: boolean
}

const SandboxManagmentParameters: React.FC<SandboxManagmentParametersProps> = ({ apiKeys, apiKeysLoading }) => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const [sandboxApiKey, setSandboxApiKey] = useState<string | undefined>(sandboxParametersState['apiKey'])
  const [sandboxLanguage, setSandboxLanguage] = useState<CodeLanguage | undefined>(sandboxParametersState['language'])
  const [resources, setResources] = useState<Resources>(sandboxParametersState['resources'])
  const [sandboxFromImageParams, setSandboxFromImageParams] = useState<CreateSandboxBaseParams>(
    sandboxParametersState['createSandboxBaseParams'],
  )
  const apiKeyFormData: ParameterFormItem = {
    label: 'API key',
    key: 'apiKey',
    placeholder: 'API key',
  }

  const languageFormData: ParameterFormItem = {
    label: 'Language',
    key: 'language',
    placeholder: 'Select sandbox language',
  }

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
      <StackedInputFormControl formItem={apiKeyFormData}>
        <FormSelectInput
          selectOptions={apiKeys}
          selectValue={sandboxApiKey}
          formItem={apiKeyFormData}
          onChangeHandler={(value) => {
            setSandboxApiKey(value)
            setSandboxParameterValue(apiKeyFormData.key as 'apiKey', value)
          }}
          loading={apiKeysLoading}
        />
      </StackedInputFormControl>
      <StackedInputFormControl formItem={languageFormData}>
        <FormSelectInput
          selectOptions={languageOptions}
          selectValue={sandboxLanguage}
          formItem={languageFormData}
          onChangeHandler={(value) => {
            setSandboxLanguage(value as CodeLanguage)
            setSandboxParameterValue(languageFormData.key as 'language', value as CodeLanguage)
          }}
        />
      </StackedInputFormControl>
      <div className="space-y-2">
        <Label htmlFor="resources">Resources</Label>
        <div id="resources" className="px-4 space-y-2">
          {resourcesFormData.map((resourceParamFormItem) => (
            <FormNumberInput
              key={resourceParamFormItem.key}
              numberValue={resources[resourceParamFormItem.key]}
              numberFormItem={resourceParamFormItem}
              onChangeHandler={(value) => {
                const resourcesNew = { ...resources, [resourceParamFormItem.key]: value }
                setResources(resourcesNew)
                setSandboxParameterValue('resources', resourcesNew)
              }}
            />
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="lifecycle">Lifecycle</Label>
        <div id="lifecycle" className="px-4 space-y-2">
          {lifecycleParamsFormData.map((lifecycleParamFormItem) => (
            <FormNumberInput
              key={lifecycleParamFormItem.key}
              numberValue={sandboxFromImageParams[lifecycleParamFormItem.key]}
              numberFormItem={lifecycleParamFormItem}
              onChangeHandler={(value) => {
                const sandboxFromImageParamsNew = { ...sandboxFromImageParams, [lifecycleParamFormItem.key]: value }
                setSandboxFromImageParams(sandboxFromImageParamsNew)
                setSandboxParameterValue('createSandboxBaseParams', sandboxFromImageParamsNew)
              }}
            />
          ))}
        </div>
      </div>
    </>
  )
}

export default SandboxManagmentParameters
