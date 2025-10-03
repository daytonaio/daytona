/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormSelectInput from '../../Inputs/SelectInput'
import FormNumberInput from '../../Inputs/NumberInput'
import { CodeLanguage, Resources, CreateSandboxBaseParams } from '@daytonaio/sdk-typescript/src'
import { usePlayground } from '@/hooks/usePlayground'
import { NumberParameterFormItem, ParameterFormItem } from '@/enums/Playground'
import { useState } from 'react'

const SandboxManagmentParameters: React.FC = () => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const [sandboxLanguage, setSandboxLanguage] = useState<CodeLanguage | undefined>(sandboxParametersState['language'])
  const [resources, setResources] = useState<Resources>(sandboxParametersState['resources'])
  const [sandboxFromImageParams, setSandboxFromImageParams] = useState<CreateSandboxBaseParams>(
    sandboxParametersState['createSandboxBaseParams'],
  )

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
    { label: 'Compute (vCPU)', key: 'cpu', min: 1, max: Infinity, placeholder: '1' },
    { label: 'Memory (GiB)', key: 'memory', min: 1, max: Infinity, placeholder: '1' },
    { label: 'Storage (GiB)', key: 'disk', min: 1, max: Infinity, placeholder: '3' },
  ]

  const lifecycleParamsFormData: (NumberParameterFormItem & {
    key: 'autoStopInterval' | 'autoArchiveInterval' | 'autoDeleteInterval'
  })[] = [
    { label: 'Stop (min)', key: 'autoStopInterval', min: 0, max: Infinity, placeholder: '15' },
    { label: 'Archive (min)', key: 'autoArchiveInterval', min: 0, max: Infinity, placeholder: '7' },
    { label: 'Delete (min)', key: 'autoDeleteInterval', min: -1, max: Infinity, placeholder: '' },
  ]

  return (
    <>
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
            <InlineInputFormControl key={resourceParamFormItem.key} formItem={resourceParamFormItem}>
              <FormNumberInput
                numberValue={resources[resourceParamFormItem.key]}
                numberFormItem={resourceParamFormItem}
                onChangeHandler={(value) => {
                  const resourcesNew = { ...resources, [resourceParamFormItem.key]: value }
                  setResources(resourcesNew)
                  setSandboxParameterValue('resources', resourcesNew)
                }}
              />
            </InlineInputFormControl>
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="lifecycle">Lifecycle</Label>
        <div id="lifecycle" className="px-4 space-y-2">
          {lifecycleParamsFormData.map((lifecycleParamFormItem) => (
            <InlineInputFormControl key={lifecycleParamFormItem.key} formItem={lifecycleParamFormItem}>
              <FormNumberInput
                numberValue={sandboxFromImageParams[lifecycleParamFormItem.key]}
                numberFormItem={lifecycleParamFormItem}
                onChangeHandler={(value) => {
                  const sandboxFromImageParamsNew = { ...sandboxFromImageParams, [lifecycleParamFormItem.key]: value }
                  setSandboxFromImageParams(sandboxFromImageParamsNew)
                  setSandboxParameterValue('createSandboxBaseParams', sandboxFromImageParamsNew)
                }}
              />
            </InlineInputFormControl>
          ))}
        </div>
      </div>
    </>
  )
}

export default SandboxManagmentParameters
