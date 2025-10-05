/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormSelectInput from '../../Inputs/SelectInput'
import FormNumberInput from '../../Inputs/NumberInput'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import { usePlayground } from '@/hooks/usePlayground'
import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import { ScreenshotRegion } from '@daytonaio/sdk'
import { ScreenshotResponse, RegionScreenshotResponse, CompressedScreenshotResponse } from '@daytonaio/api-client'
import {
  CustomizedScreenshotOptions,
  ScreenshotActions,
  ScreenshotActionFormData,
  ParameterFormData,
  VNCInteractionOptionsSectionComponentProps,
  NumberParameterFormItem,
  ParameterFormItem,
  ScreenshotFormatOption,
} from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'
import { useState } from 'react'

const VNCScreenshootOperations: React.FC<VNCInteractionOptionsSectionComponentProps> = ({
  disableActions,
  ComputerUseClient,
  wrapVNCInvokeApi,
}) => {
  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue, runPlaygroundActionWithParams } =
    usePlayground()
  const [screenshotOptions, setScreenshotOptions] = useState<CustomizedScreenshotOptions>(
    VNCInteractionOptionsParamsState['screenshotOptionsConfig'],
  )
  const [screenshotRegion, setScreenshotRegion] = useState<ScreenshotRegion>(
    VNCInteractionOptionsParamsState['screenshotRegionConfig'],
  )

  const screenshotOptionsNumberParametersFormData: (NumberParameterFormItem & { key: 'quality' | 'scale' })[] = [
    { label: 'Scale', key: 'scale', min: 0.1, max: 1, placeholder: '0.5', step: 0.1 },
    { label: 'Quality', key: 'quality', min: 1, max: 100, placeholder: '95' },
  ]

  const screenshotFormatFormData: ParameterFormItem & { key: 'format' } = {
    label: 'Format',
    key: 'format',
    placeholder: 'Select screenshot image format',
  }

  const screenshotFormatOptions = [
    {
      value: ScreenshotFormatOption.PNG,
      label: 'PNG',
    },
    {
      value: ScreenshotFormatOption.JPEG,
      label: 'JPEG',
    },
    {
      value: ScreenshotFormatOption.WEBP,
      label: 'WebP',
    },
  ]

  const screenshotShowCursorFormData: ParameterFormItem & { key: 'showCursor' } = {
    label: 'Show cursor',
    key: 'showCursor',
    placeholder: 'Show cursor in screenshot',
  }

  const screenshotOptionsFormData: ParameterFormData<CustomizedScreenshotOptions> = [
    ...screenshotOptionsNumberParametersFormData,
    screenshotFormatFormData,
    screenshotShowCursorFormData,
  ]

  const screenshotRegionNumberParametersFormData: (NumberParameterFormItem & { key: keyof ScreenshotRegion })[] = [
    { label: 'Top left X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Top left Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Width', key: 'width', min: 0, max: Infinity, placeholder: '300', required: true },
    { label: 'Height', key: 'height', min: 0, max: Infinity, placeholder: '200', required: true },
  ]

  const screenshotActionsFormData: ScreenshotActionFormData<ScreenshotRegion | CustomizedScreenshotOptions>[] = [
    {
      methodName: ScreenshotActions.TAKE_COMPRESSED,
      label: 'takeCompressed()',
      description: 'Takes a compressed screenshot of the entire screen',
      parametersFormItems: screenshotOptionsFormData,
      parametersState: screenshotOptions,
    },
    {
      methodName: ScreenshotActions.TAKE_COMPRESSED_REGION,
      label: 'takeCompressedRegion()',
      description: 'Takes a compressed screenshot of a specific region',
      parametersFormItems: [...screenshotOptionsFormData, ...screenshotRegionNumberParametersFormData],
      parametersState: {
        ...screenshotOptions,
        ...screenshotRegion,
      },
    },
    {
      methodName: ScreenshotActions.TAKE_FULL_SCREEN,
      label: 'takeFullScreen()',
      description: 'Takes a screenshot of the entire screen',
      parametersFormItems: [screenshotShowCursorFormData],
      parametersState: screenshotOptions,
    },
    {
      methodName: ScreenshotActions.TAKE_REGION,
      label: 'takeRegion()',
      description: 'Takes a screenshot of a specific region',
      parametersFormItems: [...screenshotRegionNumberParametersFormData, screenshotShowCursorFormData],
      parametersState: {
        ...screenshotRegion,
        ...screenshotOptions,
      },
    },
  ]

  // Disable logic ensures that this method is called when ComputerUseClient exists
  const screenshotActionAPICall: PlaygroundActionInvokeApi = async (screenshotActionFormData) => {
    const ScreenshotActionsClient = ComputerUseClient.screenshot
    let screenshotActionResponse: ScreenshotResponse | RegionScreenshotResponse | CompressedScreenshotResponse
    switch (screenshotActionFormData.methodName) {
      case ScreenshotActions.TAKE_COMPRESSED: {
        screenshotActionResponse = await ScreenshotActionsClient[ScreenshotActions.TAKE_COMPRESSED](
          screenshotOptions ?? undefined,
        )
        break
      }
      case ScreenshotActions.TAKE_COMPRESSED_REGION: {
        screenshotActionResponse = await ScreenshotActionsClient[ScreenshotActions.TAKE_COMPRESSED_REGION](
          screenshotRegion,
          screenshotOptions ?? undefined,
        )
        break
      }
      case ScreenshotActions.TAKE_FULL_SCREEN: {
        screenshotActionResponse = await ScreenshotActionsClient[ScreenshotActions.TAKE_FULL_SCREEN](
          screenshotOptions.showCursor,
        )
        break
      }
      case ScreenshotActions.TAKE_REGION: {
        screenshotActionResponse = await ScreenshotActionsClient[ScreenshotActions.TAKE_REGION](
          screenshotRegion,
          screenshotOptions.showCursor,
        )
        break
      }
    }
    // All screenshot actions responses have these fields in common
    type CursorPositionType = { x: number; y: number }
    const screenshotActionsResponseText = [
      `Screenshot (base64): ${screenshotActionResponse.screenshot}`,
      `Size: ${screenshotActionResponse.sizeBytes ?? 'unknown'}`,
      `Cursor position: ${screenshotActionResponse.cursorPosition ? `(${(screenshotActionResponse.cursorPosition as CursorPositionType).x}, ${(screenshotActionResponse.cursorPosition as CursorPositionType).y})` : ''}`,
    ].join('\n')
    setVNCInteractionOptionsParamValue('responseText', screenshotActionsResponseText)
  }

  return (
    <div>
      <div className="space-y-2 mt-4">
        <div className="w-full text-center mb-4">
          <Label htmlFor="screenshot-options">Screenshot Options</Label>
        </div>
        <div id="screenshot-options" className="px-4 space-y-2">
          <InlineInputFormControl formItem={screenshotFormatFormData}>
            <FormSelectInput
              selectOptions={screenshotFormatOptions}
              selectValue={screenshotOptions[screenshotFormatFormData.key as 'format']}
              formItem={screenshotFormatFormData}
              onChangeHandler={(value) => {
                const screenshotOptionsNew = {
                  ...screenshotOptions,
                  [screenshotFormatFormData.key]: value as ScreenshotFormatOption,
                }
                setScreenshotOptions(screenshotOptionsNew)
                setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
              }}
            />
          </InlineInputFormControl>
          {screenshotOptionsNumberParametersFormData.map((screenshotOptionParamFormItem) => (
            <InlineInputFormControl key={screenshotOptionParamFormItem.key} formItem={screenshotOptionParamFormItem}>
              <FormNumberInput
                numberValue={screenshotOptions[screenshotOptionParamFormItem.key]}
                numberFormItem={screenshotOptionParamFormItem}
                onChangeHandler={(value) => {
                  const screenshotOptionsNew = { ...screenshotOptions, [screenshotOptionParamFormItem.key]: value }
                  setScreenshotOptions(screenshotOptionsNew)
                  setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
                }}
              />
            </InlineInputFormControl>
          ))}
          <InlineInputFormControl formItem={screenshotShowCursorFormData}>
            <FormCheckboxInput
              checkedValue={screenshotOptions[screenshotShowCursorFormData.key as 'showCursor']}
              formItem={screenshotShowCursorFormData}
              onChangeHandler={(checked) => {
                const screenshotOptionsNew = { ...screenshotOptions, [screenshotShowCursorFormData.key]: checked }
                setScreenshotOptions(screenshotOptionsNew)
                setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
              }}
            />
          </InlineInputFormControl>
        </div>
      </div>
      <div className="space-y-2 mt-4">
        <div className="w-full text-center mb-4">
          <Label htmlFor="screenshot-options">Screenshot Region</Label>
        </div>
        <div id="screenshot-region" className="px-4 space-y-2">
          {screenshotRegionNumberParametersFormData.map((screenshotRegionParamFormItem) => (
            <InlineInputFormControl key={screenshotRegionParamFormItem.key} formItem={screenshotRegionParamFormItem}>
              <FormNumberInput
                numberValue={screenshotRegion[screenshotRegionParamFormItem.key]}
                numberFormItem={screenshotRegionParamFormItem}
                onChangeHandler={(value) => {
                  const screenshotRegionNew = { ...screenshotRegion, [screenshotRegionParamFormItem.key]: value }
                  setScreenshotRegion(screenshotRegionNew)
                  setVNCInteractionOptionsParamValue('screenshotRegionConfig', screenshotRegionNew)
                }}
              />
            </InlineInputFormControl>
          ))}
        </div>
      </div>
      <div className="space-y-6 mt-6">
        {screenshotActionsFormData.map((screenshotAction) => (
          <div key={screenshotAction.methodName}>
            <PlaygroundActionForm<ScreenshotActions>
              actionFormItem={screenshotAction}
              onRunActionClick={() =>
                runPlaygroundActionWithParams(screenshotAction, wrapVNCInvokeApi(screenshotActionAPICall))
              }
              disable={disableActions}
            />
          </div>
        ))}
      </div>
    </div>
  )
}

export default VNCScreenshootOperations
