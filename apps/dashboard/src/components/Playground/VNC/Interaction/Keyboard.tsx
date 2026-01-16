/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import {
  KeyboardActionFormData,
  KeyboardActions,
  KeyboardHotKey,
  KeyboardPress,
  KeyboardType,
  NumberParameterFormItem,
  ParameterFormData,
  VNCInteractionOptionsSectionComponentProps,
} from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { ComputerUse } from '@daytonaio/sdk'
import { useState } from 'react'
import PlaygroundActionForm from '../../ActionForm'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormTextInput from '../../Inputs/TextInput'

const VNCKeyboardOperations: React.FC<VNCInteractionOptionsSectionComponentProps> = ({
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
  const [hotKeyParams, setHotKeyParams] = useState<KeyboardHotKey>(
    VNCInteractionOptionsParamsState['keyboardHotKeyParams'],
  )
  const [pressParams, setPressParams] = useState<KeyboardPress>(VNCInteractionOptionsParamsState['keyboardPressParams'])
  const [typeParams, setTypeParams] = useState<KeyboardType>(VNCInteractionOptionsParamsState['keyboardTypeParams'])

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
      onChangeParamsValidationDisabled: true,
    },
    {
      methodName: KeyboardActions.PRESS,
      label: 'press()',
      description: 'Presses a key with optional modifiers',
      parametersFormItems: pressParamsFormData,
      parametersState: pressParams,
      onChangeParamsValidationDisabled: true,
    },
    {
      methodName: KeyboardActions.TYPE,
      label: 'type()',
      description: 'Types the specified text',
      parametersFormItems: typeParamsFormData,
      parametersState: typeParams,
      onChangeParamsValidationDisabled: true,
    },
  ]

  // Disable logic ensures that this method is called when ComputerUseClient exists -> we use as ComputerUse to silence TS compiler
  const keyboardActionAPICall: PlaygroundActionInvokeApi = async (keyboardActionFormData) => {
    const KeyboardActionsClient = (ComputerUseClient as ComputerUse).keyboard
    // All keyboard actions have Promise<void> return type -> we don't need the reponse
    switch (keyboardActionFormData.methodName) {
      case KeyboardActions.HOTKEY:
        await KeyboardActionsClient[KeyboardActions.HOTKEY](hotKeyParams.keys)
        break
      case KeyboardActions.PRESS:
        await KeyboardActionsClient[KeyboardActions.PRESS](
          pressParams.key,
          pressParams.modifiers
            ? pressParams.modifiers
                .split(',')
                .map((item) => item.trim())
                .filter((item) => item !== '')
            : undefined,
        )
        break
      case KeyboardActions.TYPE:
        await KeyboardActionsClient[KeyboardActions.TYPE](typeParams.text, typeParams.delay ?? undefined)
        break
    }
    setVNCInteractionOptionsParamValue('responseContent', '')
  }

  return (
    <div className="space-y-6">
      {keyboardActionsFormData.map((keyboardAction) => (
        <div key={keyboardAction.methodName} className="space-y-4">
          <PlaygroundActionForm<KeyboardActions>
            actionFormItem={keyboardAction}
            onRunActionClick={() =>
              runPlaygroundActionWithParams(keyboardAction, wrapVNCInvokeApi(keyboardActionAPICall))
            }
            disable={disableActions}
          />
          <div className="space-y-2">
            {keyboardAction.methodName === KeyboardActions.HOTKEY && (
              <InlineInputFormControl formItem={hotKeyParamsFormData[0]}>
                <FormTextInput
                  formItem={hotKeyParamsFormData[0]}
                  textValue={hotKeyParams[hotKeyParamsFormData[0].key]}
                  onChangeHandler={(value) =>
                    playgroundActionParamValueSetter(
                      keyboardAction,
                      hotKeyParamsFormData[0],
                      setHotKeyParams,
                      'keyboardHotKeyParams',
                      value,
                    )
                  }
                />
              </InlineInputFormControl>
            )}
            {keyboardAction.methodName === KeyboardActions.PRESS && (
              <>
                {pressParamsFormData.map((pressParamFormItem) => (
                  <InlineInputFormControl key={pressParamFormItem.key} formItem={pressParamFormItem}>
                    <FormTextInput
                      formItem={pressParamFormItem}
                      textValue={pressParams[pressParamFormItem.key]}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          keyboardAction,
                          pressParamFormItem,
                          setPressParams,
                          'keyboardPressParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
              </>
            )}
            {keyboardAction.methodName === KeyboardActions.TYPE && (
              <>
                <InlineInputFormControl formItem={typeParamsFormData[0]}>
                  <FormTextInput
                    formItem={typeParamsFormData[0]}
                    textValue={typeParams[typeParamsFormData[0].key as 'text']}
                    onChangeHandler={(value) =>
                      playgroundActionParamValueSetter(
                        keyboardAction,
                        typeParamsFormData[0],
                        setTypeParams,
                        'keyboardTypeParams',
                        value,
                      )
                    }
                  />
                </InlineInputFormControl>
                <InlineInputFormControl formItem={typeParamsFormData[1]}>
                  <FormNumberInput
                    numberFormItem={typeParamsFormData[1] as NumberParameterFormItem}
                    numberValue={typeParams[typeParamsFormData[1].key as 'delay']}
                    onChangeHandler={(value) =>
                      playgroundActionParamValueSetter(
                        keyboardAction,
                        typeParamsFormData[1],
                        setTypeParams,
                        'keyboardTypeParams',
                        value,
                      )
                    }
                  />
                </InlineInputFormControl>
              </>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default VNCKeyboardOperations
