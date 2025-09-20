/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { KeyboardActions, KeyboardActionFormData, ParameterFormData, NumberParameterFormItem } from '@/enums/Playground'
import { KeyboardHotKey, KeyboardPress, KeyboardType } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { Loader2, Play } from 'lucide-react'
import { useState } from 'react'

const VNCKeyboardOperations: React.FC = () => {
  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue } = usePlayground()
  const [hotKeyParams, setHotKeyParams] = useState<KeyboardHotKey>(
    VNCInteractionOptionsParamsState['keyboardHotKeyParams'],
  )
  const [pressParams, setPressParams] = useState<KeyboardPress>(VNCInteractionOptionsParamsState['keyboardPressParams'])
  const [typeParams, setTypeParams] = useState<KeyboardType>(VNCInteractionOptionsParamsState['keyboardTypeParams'])
  const [runningKeyboardActionMethod, setRunningKeyboardActionMethod] = useState<KeyboardActions | null>(null)
  const [keyboardActionError, setKeyboardActionError] = useState<Partial<Record<KeyboardActions, string>>>({})

  const hotKeyParamsFormData: ParameterFormData<KeyboardHotKey> = [
    { label: 'Keys', key: 'keys', placeholder: 'ctrl+c, alt+tab', required: true },
  ]

  const pressParamsFormData: ParameterFormData<KeyboardPress> = [
    { label: 'Key', key: 'key', placeholder: 'Enter', required: true },
    { label: 'Modifiers', key: 'modifiers', placeholder: 'ctrl, alt, shift' },
  ]

  const typeParamsFormData: ParameterFormData<KeyboardType> = [
    { label: 'Text', key: 'text', placeholder: 'Daytona', required: true },
    { label: 'Delay(ms)', key: 'delay', placeholder: '50ms', min: 0, max: Infinity, step: 10 },
  ]

  const keyboardActionsFormData: KeyboardActionFormData<KeyboardHotKey | KeyboardPress | KeyboardType>[] = [
    {
      methodName: KeyboardActions.HOTKEY,
      label: 'hotkey()',
      description: 'Presses a hotkey combination',
      parametersFormItems: hotKeyParamsFormData,
      parametersState: hotKeyParams,
    },
    {
      methodName: KeyboardActions.PRESS,
      label: 'press()',
      description: 'Presses a key with optional modifiers',
      parametersFormItems: pressParamsFormData,
      parametersState: pressParams,
    },
    {
      methodName: KeyboardActions.TYPE,
      label: 'type()',
      description: 'Types the specified text',
      parametersFormItems: typeParamsFormData,
      parametersState: typeParams,
    },
  ]

  const onKeyboardActionRunClick = <T extends KeyboardHotKey | KeyboardPress | KeyboardType>(
    keyboardActionFormData: KeyboardActionFormData<T>,
    keyboardActionParamsFormData: ParameterFormData<T>,
    keyboardActionParamsState: T,
  ) => {
    setRunningKeyboardActionMethod(keyboardActionFormData.methodName)
    // Validate if all required params are set if they exist
    if (keyboardActionParamsFormData.some((formItem) => formItem.required)) {
      const keyboardActionEmptyParamFormItem = keyboardActionParamsFormData
        .filter((formItem) => formItem.required)
        .find((formItem) => {
          const value = keyboardActionParamsState[formItem.key]
          return value === '' || value === undefined
        })
      if (keyboardActionEmptyParamFormItem) {
        setKeyboardActionError({
          [keyboardActionFormData.methodName]: `${keyboardActionEmptyParamFormItem?.label} parameter is required for this action`,
        })
        setRunningKeyboardActionMethod(null)
        return
      }
    }
    // KeyboardPress modifiers postprocessing: .split(',').map(item => item.trim()).filter(item => item !== '')
    //TODO -> API CALL
    setKeyboardActionError({}) // Reset error
    setRunningKeyboardActionMethod(null)
  }

  return (
    <div className="space-y-6">
      {keyboardActionsFormData.map((keyboardAction) => (
        <div key={keyboardAction.methodName} className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor={keyboardAction.methodName}>{keyboardAction.label}</Label>
              <p id={keyboardAction.methodName} className="text-sm text-muted-foreground mt-1 pl-1">
                {keyboardAction.description}
              </p>
            </div>
            <div>
              {' '}
              <Button
                disabled={!!runningKeyboardActionMethod}
                variant="outline"
                title="Run"
                onClick={() =>
                  onKeyboardActionRunClick<typeof keyboardAction.parametersState>(
                    keyboardAction,
                    keyboardAction.parametersFormItems,
                    keyboardAction.parametersState,
                  )
                }
              >
                {runningKeyboardActionMethod === keyboardAction.methodName ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>
          <div>
            {keyboardActionError[keyboardAction.methodName] && (
              <p className="text-sm text-red-500 mt-2">{keyboardActionError[keyboardAction.methodName]}</p>
            )}
          </div>
          <div className="px-4 space-y-2">
            {keyboardAction.methodName === KeyboardActions.HOTKEY && (
              <div className="flex items-center gap-4">
                <Label htmlFor={hotKeyParamsFormData[0].key} className="w-32 flex-shrink-0">
                  <span>
                    {hotKeyParamsFormData[0].required ? <span className="text-red-500">* </span> : null}
                    <span>{`${hotKeyParamsFormData[0].label}:`}</span>
                  </span>
                </Label>
                <Input
                  id={hotKeyParamsFormData[0].key}
                  className="w-full"
                  placeholder={hotKeyParamsFormData[0].placeholder}
                  value={hotKeyParams[hotKeyParamsFormData[0].key]}
                  onChange={(e) => {
                    const hotKeyParamsNew = { ...hotKeyParams, [hotKeyParamsFormData[0].key]: e.target.value }
                    setHotKeyParams(hotKeyParamsNew)
                    setVNCInteractionOptionsParamValue('keyboardHotKeyParams', hotKeyParamsNew)
                  }}
                />
              </div>
            )}
            {keyboardAction.methodName === KeyboardActions.PRESS && (
              <>
                <div className="flex items-center gap-4">
                  <Label htmlFor={pressParamsFormData[0].key} className="w-32 flex-shrink-0">
                    <span>
                      {pressParamsFormData[0].required ? <span className="text-red-500">* </span> : null}
                      <span>{`${pressParamsFormData[0].label}:`}</span>
                    </span>
                  </Label>
                  <Input
                    id={pressParamsFormData[0].key}
                    className="w-full"
                    placeholder={pressParamsFormData[0].placeholder}
                    value={pressParams[pressParamsFormData[0].key]}
                    onChange={(e) => {
                      const pressParamsNew = { ...pressParams, [pressParamsFormData[0].key]: e.target.value }
                      setPressParams(pressParamsNew)
                      setVNCInteractionOptionsParamValue('keyboardPressParams', pressParamsNew)
                    }}
                  />
                </div>
                <div className="flex items-center gap-4">
                  <Label htmlFor={pressParamsFormData[1].key} className="w-32 flex-shrink-0">
                    <span>
                      {pressParamsFormData[1].required ? <span className="text-red-500">* </span> : null}
                      <span>{`${pressParamsFormData[1].label}:`}</span>
                    </span>
                  </Label>
                  <Input
                    id={pressParamsFormData[1].key}
                    className="w-full"
                    placeholder={pressParamsFormData[1].placeholder}
                    value={pressParams[pressParamsFormData[1].key]}
                    onChange={(e) => {
                      const pressParamsNew = { ...pressParams, [pressParamsFormData[1].key]: e.target.value }
                      setPressParams(pressParamsNew)
                      setVNCInteractionOptionsParamValue('keyboardPressParams', pressParamsNew)
                    }}
                  />
                </div>
              </>
            )}
            {keyboardAction.methodName === KeyboardActions.TYPE && (
              <>
                <div className="flex items-center gap-4">
                  <Label htmlFor={typeParamsFormData[0].key} className="w-32 flex-shrink-0">
                    <span>
                      {typeParamsFormData[0].required ? <span className="text-red-500">* </span> : null}
                      <span>{`${typeParamsFormData[0].label}:`}</span>
                    </span>
                  </Label>
                  <Input
                    id={typeParamsFormData[0].key}
                    className="w-full"
                    placeholder={typeParamsFormData[0].placeholder}
                    value={typeParams[typeParamsFormData[0].key]}
                    onChange={(e) => {
                      const typeParamsNew = { ...typeParams, [typeParamsFormData[0].key]: e.target.value }
                      setTypeParams(typeParamsNew)
                      setVNCInteractionOptionsParamValue('keyboardTypeParams', typeParamsNew)
                    }}
                  />
                </div>
                <div className="flex items-center gap-4">
                  <Label htmlFor={typeParamsFormData[1].key} className="w-32 flex-shrink-0">
                    <span>
                      {typeParamsFormData[1].required ? <span className="text-red-500">* </span> : null}
                      <span>{`${typeParamsFormData[1].label}:`}</span>
                    </span>
                  </Label>
                  <Input
                    id={typeParamsFormData[1].key}
                    type="number"
                    className="w-full"
                    min={(typeParamsFormData[1] as NumberParameterFormItem).min}
                    max={(typeParamsFormData[1] as NumberParameterFormItem).max}
                    placeholder={typeParamsFormData[1].placeholder}
                    step={(typeParamsFormData[1] as NumberParameterFormItem).step}
                    value={typeParams[typeParamsFormData[1].key]}
                    onChange={(e) => {
                      const newValue = e.target.value ? Number(e.target.value) : undefined
                      const typeParamsNew = { ...typeParams, [typeParamsFormData[1].key]: newValue }
                      setTypeParams(typeParamsNew)
                      setVNCInteractionOptionsParamValue('keyboardTypeParams', typeParamsNew)
                    }}
                  />
                </div>
              </>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default VNCKeyboardOperations
