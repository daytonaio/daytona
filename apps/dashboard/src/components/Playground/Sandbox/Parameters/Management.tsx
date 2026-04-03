/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Tooltip } from '@/components/Tooltip'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { SANDBOX_SNAPSHOT_DEFAULT_VALUE, SANDBOX_TARGET_DEFAULT_VALUE } from '@/constants/Playground'
import { NumberParameterFormItem, ParameterFormItem } from '@/contexts/PlaygroundContext'
import { usePlayground } from '@/hooks/usePlayground'
import { useRegions } from '@/hooks/useRegions'
import { getLanguageCodeToRun } from '@/lib/playground'
import { cn, getRegionFullDisplayName } from '@/lib/utils'
import { SnapshotDto } from '@daytonaio/api-client'
import { CodeLanguage, Resources } from '@daytonaio/sdk'
import { ChevronDownIcon, HelpCircleIcon } from 'lucide-react'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'
import { useEffect, useState } from 'react'

type SandboxManagementParametersProps = {
  snapshotsData: Array<SnapshotDto>
  snapshotsLoading: boolean
}

const SandboxManagementParameters: React.FC<SandboxManagementParametersProps> = ({
  snapshotsData,
  snapshotsLoading,
}) => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const { availableRegions: regions, loadingAvailableRegions: regionsLoading } = useRegions()
  const [advancedOpen, setAdvancedOpen] = useState(false)

  const sandboxLanguage = sandboxParametersState['language']
  const sandboxSnapshotName = sandboxParametersState['snapshotName']
  const sandboxTarget = sandboxParametersState['sandboxTarget']
  const resources = sandboxParametersState['resources']
  const sandboxFromImageParams = sandboxParametersState['createSandboxBaseParams']

  const languageFormData: ParameterFormItem = {
    label: 'Language',
    key: 'language',
    placeholder: 'Select sandbox language',
  }

  const targetFormData: ParameterFormItem = {
    label: 'Target',
    key: 'sandboxTarget',
    placeholder: 'Select sandbox target',
  }

  const sandboxSnapshotFormData: ParameterFormItem = {
    label: 'Snapshot',
    key: 'snapshotName',
    placeholder: 'Select sandbox snapshot',
  }

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

  useEffect(() => {
    setSandboxParameterValue('codeRunParams', {
      languageCode: getLanguageCodeToRun(sandboxParametersState.language),
    })
  }, [sandboxParametersState.language, setSandboxParameterValue])

  useEffect(() => {
    if (regionsLoading) return
    if (!sandboxTarget || sandboxTarget === SANDBOX_TARGET_DEFAULT_VALUE) return
    const exists = regions.some((r) => r.id === sandboxTarget)
    if (!exists) {
      setSandboxParameterValue('sandboxTarget', SANDBOX_TARGET_DEFAULT_VALUE)
    }
  }, [regions, regionsLoading, sandboxTarget, setSandboxParameterValue])

  useEffect(() => {
    if (snapshotsLoading) return
    if (!sandboxSnapshotName || sandboxSnapshotName === SANDBOX_SNAPSHOT_DEFAULT_VALUE) return
    const exists = snapshotsData.some((s) => s.name === sandboxSnapshotName)
    if (!exists) {
      setSandboxParameterValue('snapshotName', SANDBOX_SNAPSHOT_DEFAULT_VALUE)
    }
  }, [snapshotsData, snapshotsLoading, sandboxSnapshotName, setSandboxParameterValue])

  const nonDefaultSnapshotSelected = sandboxSnapshotName && sandboxSnapshotName !== SANDBOX_SNAPSHOT_DEFAULT_VALUE

  const targetSelectOptions = [
    { value: SANDBOX_TARGET_DEFAULT_VALUE, label: 'Default (organization)' },
    ...regions.map((region) => ({
      value: region.id,
      label: getRegionFullDisplayName(region),
    })),
  ]

  const snapshotSelectOptions = [
    { value: SANDBOX_SNAPSHOT_DEFAULT_VALUE, label: 'Default' },
    ...snapshotsData.map((snapshot) => ({
      value: snapshot.name,
      label: snapshot.name,
    })),
  ]

  return (
    <>
      <StackedInputFormControl formItem={languageFormData}>
        <FormSelectInput
          selectOptions={languageOptions}
          selectValue={sandboxLanguage}
          formItem={languageFormData}
          onChangeHandler={(value) => {
            setSandboxParameterValue(languageFormData.key as 'language', value as CodeLanguage)
          }}
        />
      </StackedInputFormControl>

      <div className="space-y-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-auto px-0 py-1 text-muted-foreground hover:text-foreground"
          aria-expanded={advancedOpen}
          aria-controls="playground-advanced-options"
          onClick={() => setAdvancedOpen((o) => !o)}
        >
          <ChevronDownIcon
            className={cn('mr-1 size-4 shrink-0 transition-transform', advancedOpen && 'rotate-180')}
            aria-hidden
          />
          Advanced options
        </Button>
        {advancedOpen && (
          <div id="playground-advanced-options" className="space-y-4 border-l-2 border-muted pl-3">
            <StackedInputFormControl formItem={targetFormData}>
              <FormSelectInput
                selectOptions={targetSelectOptions}
                loading={regionsLoading}
                selectValue={sandboxTarget ?? SANDBOX_TARGET_DEFAULT_VALUE}
                formItem={targetFormData}
                onChangeHandler={(value) => {
                  setSandboxParameterValue('sandboxTarget', value)
                }}
              />
            </StackedInputFormControl>
            <p className="text-xs text-muted-foreground -mt-2">
              Region for sandbox creation. If not set, your organization default is used.
            </p>
            <StackedInputFormControl formItem={sandboxSnapshotFormData}>
              <FormSelectInput
                selectOptions={snapshotSelectOptions}
                loading={snapshotsLoading}
                selectValue={sandboxSnapshotName ?? SANDBOX_SNAPSHOT_DEFAULT_VALUE}
                formItem={sandboxSnapshotFormData}
                onChangeHandler={(snapshotName) => {
                  setSandboxParameterValue('snapshotName', snapshotName)
                }}
              />
            </StackedInputFormControl>
          </div>
        )}
      </div>

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
                <button type="button" className="rounded-full">
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
                  setSandboxParameterValue('resources', { ...resources, [resourceParamFormItem.key]: value })
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
                  setSandboxParameterValue('createSandboxBaseParams', {
                    ...sandboxFromImageParams,
                    [lifecycleParamFormItem.key]: value,
                  })
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
