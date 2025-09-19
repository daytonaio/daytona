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
import {
  CustomizedScreenshotOptions,
  ScreenshotActions,
  ScreenshotActionFormData,
  ParameterFormData,
} from '@/enums/Playground'
import { NumberParameterFormItem, ParameterFormItem, ScreenshotFormatOption } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'
import { useState } from 'react'

const VNCScreenshootOperations: React.FC = () => {
  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue } = usePlayground()
  const [screenshotOptions, setScreenshotOptions] = useState<CustomizedScreenshotOptions>(
    VNCInteractionOptionsParamsState['screenshotOptionsConfig'],
  )
  const [screenshotRegion, setScreenshotRegion] = useState<ScreenshotRegion>(
    VNCInteractionOptionsParamsState['screenshotRegionConfig'],
  )
  const [runningScreenshotActionMethod, setRunningScreenshotActionMethod] = useState<ScreenshotActions | null>(null)
  const [screenshotActionError, setScreenshotActionError] = useState<Partial<Record<ScreenshotActions, string>>>({})

  const screenshotOptionsNumberParametersFormData: (NumberParameterFormItem & { key: 'quality' | 'scale' })[] = [
    { label: 'Scale', key: 'scale', min: 0.1, max: 1, placeholder: '0.5', step: 0.1 },
    { label: 'Quality', key: 'quality', min: 1, max: 100, placeholder: '95' },
  ]

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

  const screenshotRegionNumberParametersFormData: (NumberParameterFormItem & { key: keyof ScreenshotRegion })[] = [
    { label: 'Top left X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Top left Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Width', key: 'width', min: 0, max: Infinity, placeholder: '300', required: true },
    { label: 'Height', key: 'height', min: 0, max: Infinity, placeholder: '200', required: true },
  ]

  const screenshotActionsFromData: ScreenshotActionFormData[] = [
    {
      methodName: ScreenshotActions.TAKE_COMPRESSED,
      label: 'takeCompressed()',
      description: 'Takes a compressed screenshot of the entire screen',
    },
    {
      methodName: ScreenshotActions.TAKE_COMPRESSED_REGION,
      label: 'takeCompressedRegion()',
      description: 'Takes a compressed screenshot of a specific region',
      usesScreenshotRegion: true,
    },
    {
      methodName: ScreenshotActions.TAKE_FULL_SCREEN,
      label: 'takeFullScreen()',
      description: 'Takes a screenshot of the entire screen',
    },
    {
      methodName: ScreenshotActions.TAKE_REGION,
      label: 'takeRegion()',
      description: 'Takes a screenshot of a specific region',
      usesScreenshotRegion: true,
    },
  ]

  const onScreenshotActionRunClick = (screenshotActionFormData: ScreenshotActionFormData) => {
    setRunningScreenshotActionMethod(screenshotActionFormData.methodName)
    // Validate if all ScreenshotRegion parameters are set
    if (screenshotActionFormData.usesScreenshotRegion) {
      const screenshotRegionEmptyParamKey = Object.keys(screenshotRegion).find((key) => {
        const value = screenshotRegion[key as keyof ScreenshotRegion]
        return value === undefined
      })
      if (screenshotRegionEmptyParamKey) {
        setScreenshotActionError({
          [screenshotActionFormData.methodName]: `${screenshotRegionNumberParametersFormData.find((screenshotRegionParam) => screenshotRegionParam.key === screenshotRegionEmptyParamKey)?.label} parameter is required for this action`,
        })
        setRunningScreenshotActionMethod(null)
        return
      }
    }
    //TODO -> API CALL
    setScreenshotActionError({}) // Reset error
    setRunningScreenshotActionMethod(null)
  }

  return (
    <div>
      <div className="space-y-2 mt-4">
        <div className="w-full text-center mb-4">
          {' '}
          <Label htmlFor="screenshot-options">Screenshot Options</Label>
        </div>
        <div id="screenshot-options" className="px-4 space-y-2">
          <div className="flex items-center gap-4">
            <Label htmlFor="format" className="w-32 flex-shrink-0">
              Format:
            </Label>
            <Select
              value={screenshotOptions['format']}
              onValueChange={(format) => {
                const screenshotOptionsNew = { ...screenshotOptions, format: format as ScreenshotFormatOption }
                setScreenshotOptions(screenshotOptionsNew)
                setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
              }}
            >
              <SelectTrigger className="w-full box-border rounded-lg" aria-label="Select screenshot format">
                <SelectValue id="format" placeholder="Format" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {screenshotFormatOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {screenshotOptionsNumberParametersFormData.map((screenshotOptionParam) => (
            <div key={screenshotOptionParam.key} className="flex items-center gap-4">
              <Label htmlFor={screenshotOptionParam.key} className="w-32 flex-shrink-0">
                {`${screenshotOptionParam.label}:`}
              </Label>
              <Input
                id={screenshotOptionParam.key}
                type="number"
                className="w-full"
                min={screenshotOptionParam.min}
                max={screenshotOptionParam.max}
                placeholder={screenshotOptionParam.placeholder}
                step={screenshotOptionParam.step}
                value={screenshotOptions[screenshotOptionParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const screenshotOptionsNew = { ...screenshotOptions, [screenshotOptionParam.key]: newValue }
                  setScreenshotOptions(screenshotOptionsNew)
                  setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
                }}
              />
            </div>
          ))}
          <div className="flex items-center gap-4">
            <Label htmlFor="show_cursor" className="w-32 flex-shrink-0">
              Show cursor:
            </Label>
            <div className="flex-1 text-center">
              <Checkbox
                id="show_cursor"
                checked={screenshotOptions['showCursor']}
                onCheckedChange={(value) => {
                  const screenshotOptionsNew = { ...screenshotOptions, showCursor: !!value }
                  setScreenshotOptions(screenshotOptionsNew)
                  setVNCInteractionOptionsParamValue('screenshotOptionsConfig', screenshotOptionsNew)
                }}
              />
            </div>
          </div>
        </div>
      </div>
      <div className="space-y-2 mt-4">
        <div className="w-full text-center mb-4">
          {' '}
          <Label htmlFor="screenshot-options">Screenshot Region</Label>
        </div>
        <div id="screenshot-region" className="px-4 space-y-2">
          {screenshotRegionNumberParametersFormData.map((screenshotRegionParam) => (
            <div key={screenshotRegionParam.key} className="flex items-center gap-4">
              <Label htmlFor={screenshotRegionParam.key} className="w-32 flex-shrink-0">
                <span>
                  {screenshotRegionParam.required ? <span className="text-red-500">* </span> : null}
                  <span>{`${screenshotRegionParam.label}:`}</span>
                </span>
              </Label>
              <Input
                id={screenshotRegionParam.key}
                type="number"
                className="w-full"
                min={screenshotRegionParam.min}
                max={screenshotRegionParam.max}
                placeholder={screenshotRegionParam.placeholder}
                step={screenshotRegionParam.step}
                value={screenshotRegion[screenshotRegionParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const screenshotRegionNew = { ...screenshotRegion, [screenshotRegionParam.key]: newValue }
                  setScreenshotRegion(screenshotRegionNew)
                  setVNCInteractionOptionsParamValue('screenshotRegionConfig', screenshotRegionNew)
                }}
              />
            </div>
          ))}
        </div>
      </div>
      <div className="space-y-6 mt-6">
        {screenshotActionsFromData.map((screenshotAction) => (
          <div key={screenshotAction.methodName}>
            <div className="flex items-center justify-between">
              <div>
                <Label>{screenshotAction.label}</Label>
                <p className="text-sm text-muted-foreground mt-1 pl-1">{screenshotAction.description}</p>
              </div>
              <Button
                disabled={!!runningScreenshotActionMethod}
                variant="outline"
                title="Run"
                onClick={() => onScreenshotActionRunClick(screenshotAction)}
              >
                {runningScreenshotActionMethod === screenshotAction.methodName ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
              </Button>
            </div>
            {screenshotActionError[screenshotAction.methodName] && (
              <p className="text-sm text-red-500 mt-2">{screenshotActionError[screenshotAction.methodName]}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export default VNCScreenshootOperations
