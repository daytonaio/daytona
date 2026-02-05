/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Tooltip } from '@/components/Tooltip'
import { Label } from '@/components/ui/label'
import { SANDBOX_SNAPSHOT_DEFAULT_VALUE } from '@/constants/Playground'
import { NumberParameterFormItem, ParameterFormItem } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { SnapshotDto } from '@daytonaio/api-client'
import { CodeLanguage, CreateSandboxBaseParams, Resources } from '@daytonaio/sdk'
import { HelpCircleIcon } from 'lucide-react'
import { useState } from 'react'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'

type SandboxManagementParametersProps = {
  snapshotsData: Array<SnapshotDto>
  snapshotsLoading: boolean
}

const SandboxManagementParameters: React.FC<SandboxManagementParametersProps> = ({
  snapshotsData,
  snapshotsLoading,
}) => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const [sandboxLanguage, setSandboxLanguage] = useState<CodeLanguage | undefined>(sandboxParametersState['language'])
  const [sandboxSnapshotName, setSandboxSnapshotName] = useState<string | undefined>(
    sandboxParametersState['snapshotName'],
  )
  const [resources, setResources] = useState<Resources>(sandboxParametersState['resources'])
  const [sandboxFromImageParams, setSandboxFromImageParams] = useState<CreateSandboxBaseParams>(
    sandboxParametersState['createSandboxBaseParams'],
  )

  const languageFormData: ParameterFormItem = {
    label: 'Language',
    key: 'language',
    placeholder: 'Select sandbox language',
  }

  const sandboxSnapshotFormData: ParameterFormItem = {
    label: 'Snapshot',
    key: 'snapshotName',
    placeholder: 'Select sandbox snapshot',
  }

  // Available languages
  const languageOptions = [
    {
      value: CodeLanguage.PYTHON,
      label: 'Python (default)',
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

  const nonDefaultSnapshotSelected = sandboxSnapshotName && sandboxSnapshotName !== SANDBOX_SNAPSHOT_DEFAULT_VALUE

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
      <StackedInputFormControl formItem={sandboxSnapshotFormData}>
        <FormSelectInput
          selectOptions={[
            { value: SANDBOX_SNAPSHOT_DEFAULT_VALUE, label: 'Default' },
            ...snapshotsData.map((snapshot) => ({
              value: snapshot.name,
              label: snapshot.name,
            })),
          ]}
          loading={snapshotsLoading}
          selectValue={sandboxSnapshotName}
          formItem={sandboxSnapshotFormData}
          onChangeHandler={(snapshotName) => {
            setSandboxSnapshotName(snapshotName)
            setSandboxParameterValue(sandboxSnapshotFormData.key as 'snapshotName', snapshotName)
          }}
        />
      </StackedInputFormControl>
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Label htmlFor="resources" className="text-sm text-muted-foreground">
            Resources
          </Label>
          {nonDefaultSnapshotSelected && (
            <Tooltip
              content={
                <div className="text-balance text-center max-w-[300px]">
                  Resources cannot be modified when a non-default snapshot is selected.
                </div>
              }
              label={
                <button className="rounded-full">
                  <HelpCircleIcon className="h-4 w-4 text-muted-foreground" />
                </button>
              }
            />
          )}
        </div>
        <div id="resources" className="space-y-2">
          {resourcesFormData.map((resourceParamFormItem) => (
            <InlineInputFormControl key={resourceParamFormItem.key} formItem={resourceParamFormItem}>
              <FormNumberInput
                disabled={Boolean(nonDefaultSnapshotSelected)}
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
        <Label htmlFor="lifecycle" className="text-sm text-muted-foreground">
          Lifecycle
        </Label>
        <div id="lifecycle" className="space-y-2">
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

export default SandboxManagementParameters
