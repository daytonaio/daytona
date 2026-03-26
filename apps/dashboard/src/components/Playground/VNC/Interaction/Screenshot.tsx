/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import {
  CustomizedScreenshotOptions,
  NumberParameterFormItem,
  ParameterFormData,
  ParameterFormItem,
  PlaygroundActionInvokeApi,
  ScreenshotActionFormData,
  VNCInteractionOptionsSectionComponentProps,
} from '@/contexts/PlaygroundContext'
import { ScreenshotActions, ScreenshotFormatOption } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { CompressedScreenshotResponse, RegionScreenshotResponse } from '@daytonaio/api-client'
import { ComputerUse, ScreenshotRegion } from '@daytonaio/sdk'
import { ScreenshotResponse } from '@daytonaio/toolbox-api-client'
import PlaygroundActionForm from '../../ActionForm'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'

const VNCScreenshotOperations: React.FC<VNCInteractionOptionsSectionComponentProps> = ({
  disableActions,
  ComputerUseClient,
  wrapVNCInvokeApi,
}) => {
  const {
    VNCInteractionOptionsParamsState,
    setVNCInteractionOptionsParamValue,
    playgroundActionParamValueSetter,
    runPlaygroundActionWithParams,
  } = usePlayground()
  const screenshotOptions = VNCInteractionOptionsParamsState['screenshotOptionsConfig']
  const screenshotRegion = VNCInteractionOptionsParamsState['screenshotRegionConfig']

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
      onChangeParamsValidationDisabled: true,
    },
    // {
    //   methodName: ScreenshotActions.TAKE_COMPRESSED_REGION,
    //   label: 'takeCompressedRegion()',
    //   description: 'Takes a compressed screenshot of a specific region',
    //   parametersFormItems: [...screenshotOptionsFormData, ...screenshotRegionNumberParametersFormData],
    //   parametersState: {
    //     ...screenshotOptions,
    //     ...screenshotRegion,
    //   },
    //   onChangeParamsValidationDisabled: true,
    // },
    {
      methodName: ScreenshotActions.TAKE_FULL_SCREEN,
      label: 'takeFullScreen()',
      description: 'Takes a screenshot of the entire screen',
      parametersFormItems: [screenshotShowCursorFormData],
      parametersState: screenshotOptions,
      onChangeParamsValidationDisabled: true,
    },
    // {
    //   methodName: ScreenshotActions.TAKE_REGION,
    //   label: 'takeRegion()',
    //   description: 'Takes a screenshot of a specific region',
    //   parametersFormItems: [...screenshotRegionNumberParametersFormData, screenshotShowCursorFormData],
    //   parametersState: {
    //     ...screenshotRegion,
    //     ...screenshotOptions,
    //   },
    //   onChangeParamsValidationDisabled: true,
    // },
  ]

  // Disable logic ensures that this method is called when ComputerUseClient exists -> we use as ComputerUse to silence TS compiler
  const screenshotActionAPICall: PlaygroundActionInvokeApi = async (screenshotActionFormData) => {
    const ScreenshotActionsClient = (ComputerUseClient as ComputerUse).screenshot
    let screenshotActionResponse: ScreenshotResponse | RegionScreenshotResponse | CompressedScreenshotResponse = {
      screenshot: '',
    }
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
      screenshotActionResponse.cursorPosition
        ? `Cursor position: (${(screenshotActionResponse.cursorPosition as CursorPositionType).x}, ${(screenshotActionResponse.cursorPosition as CursorPositionType).y})`
        : '',
    ].join('\n')
    setVNCInteractionOptionsParamValue('responseContent', screenshotActionsResponseText)

    // Auto-download the screenshot image
    if (screenshotActionResponse.screenshot) {
      const format = screenshotOptions.format ?? ScreenshotFormatOption.PNG
      const mimeType =
        format === ScreenshotFormatOption.JPEG
          ? 'image/jpeg'
          : format === ScreenshotFormatOption.WEBP
            ? 'image/webp'
            : 'image/png'
      const link = document.createElement('a')
      link.href = `data:${mimeType};base64,${screenshotActionResponse.screenshot}`
      link.download = `screenshot-${Date.now()}.${format}`
      link.click()
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <div className="w-full">
          <Label htmlFor="screenshot-options" className="text-sm text-muted-foreground">
            Screenshot Options
          </Label>
        </div>
        <div id="screenshot-options" className="space-y-2">
          <InlineInputFormControl formItem={screenshotFormatFormData}>
            <FormSelectInput
              selectOptions={screenshotFormatOptions}
              selectValue={screenshotOptions[screenshotFormatFormData.key as 'format']}
              formItem={screenshotFormatFormData}
              onChangeHandler={(value) => {
                // Since all screenshot actions have onChangeParamsValidationDisabled set, the actionFormData parameter is irrelevant. We pass the first action simply to satisfy the method's parameter requirements.
                playgroundActionParamValueSetter(
                  screenshotActionsFormData[0],
                  screenshotFormatFormData,
                  'screenshotOptionsConfig',
                  value,
                )
              }}
            />
          </InlineInputFormControl>
          {screenshotOptionsNumberParametersFormData.map((screenshotOptionParamFormItem) => (
            <InlineInputFormControl key={screenshotOptionParamFormItem.key} formItem={screenshotOptionParamFormItem}>
              <FormNumberInput
                numberValue={screenshotOptions[screenshotOptionParamFormItem.key]}
                numberFormItem={screenshotOptionParamFormItem}
                onChangeHandler={(value) => {
                  // Since all screenshot actions have onChangeParamsValidationDisabled set, the actionFormData parameter is irrelevant. We pass the first action simply to satisfy the method's parameter requirements.
                  playgroundActionParamValueSetter(
                    screenshotActionsFormData[0],
                    screenshotOptionParamFormItem,
                    'screenshotOptionsConfig',
                    value,
                  )
                }}
              />
            </InlineInputFormControl>
          ))}
          <InlineInputFormControl formItem={screenshotShowCursorFormData}>
            <FormCheckboxInput
              checkedValue={screenshotOptions[screenshotShowCursorFormData.key as 'showCursor']}
              formItem={screenshotShowCursorFormData}
              onChangeHandler={(checked) => {
                // Since all screenshot actions have onChangeParamsValidationDisabled set, the actionFormData parameter is irrelevant. We pass the first action simply to satisfy the method's parameter requirements.
                playgroundActionParamValueSetter(
                  screenshotActionsFormData[0],
                  screenshotShowCursorFormData,
                  'screenshotOptionsConfig',
                  checked,
                )
              }}
            />
          </InlineInputFormControl>
        </div>
      </div>
      <div className="space-y-2">
        <div className="w-full">
          <Label htmlFor="screenshot-options" className="text-sm text-muted-foreground">
            Screenshot Region
          </Label>
        </div>
        <div id="screenshot-region" className="space-y-2">
          {screenshotRegionNumberParametersFormData.map((screenshotRegionParamFormItem) => (
            <InlineInputFormControl key={screenshotRegionParamFormItem.key} formItem={screenshotRegionParamFormItem}>
              <FormNumberInput
                numberValue={screenshotRegion[screenshotRegionParamFormItem.key]}
                numberFormItem={screenshotRegionParamFormItem}
                onChangeHandler={(value) => {
                  // Since all screenshot actions have onChangeParamsValidationDisabled set, the actionFormData parameter is irrelevant. We pass the first action simply to satisfy the method's parameter requirements.
                  playgroundActionParamValueSetter(
                    screenshotActionsFormData[0],
                    screenshotRegionParamFormItem,
                    'screenshotRegionConfig',
                    value,
                  )
                }}
              />
            </InlineInputFormControl>
          ))}
        </div>
      </div>
      <div className="flex flex-col gap-4">
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

export default VNCScreenshotOperations
