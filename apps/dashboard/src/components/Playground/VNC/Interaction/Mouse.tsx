/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import {
  MouseActionFormData,
  MouseActions,
  MouseButton,
  MouseClick,
  MouseDrag,
  MouseMove,
  MouseScroll,
  MouseScrollDirection,
  NumberParameterFormItem,
  ParameterFormData,
  ParameterFormItem,
  PlaygroundActionFormDataBasic,
  VNCInteractionOptionsSectionComponentProps,
} from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { ComputerUse } from '@daytonaio/sdk'
import React, { useState } from 'react'
import PlaygroundActionForm from '../../ActionForm'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'

const mouseButtonFormData: ParameterFormItem & { key: 'button' } = {
  label: 'Button',
  key: 'button',
  placeholder: 'Select mouse button',
}

type MouseActionWithParamsFormData = MouseActionFormData<MouseClick | MouseDrag | MouseMove | MouseScroll>

const VNCMouseOperations: React.FC<VNCInteractionOptionsSectionComponentProps> = ({
  disableActions,
  ComputerUseClient,
  wrapVNCInvokeApi,
}) => {
  const {
    VNCInteractionOptionsParamsState,
    setVNCInteractionOptionsParamValue,
    playgroundActionParamValueSetter,
    runPlaygroundActionWithParams,
    runPlaygroundActionWithoutParams,
  } = usePlayground()
  const [mouseClickParams, setMouseClickParams] = useState<MouseClick>(
    VNCInteractionOptionsParamsState['mouseClickParams'],
  )
  const [mouseDragParams, setMouseDragParams] = useState<MouseDrag>(VNCInteractionOptionsParamsState['mouseDragParams'])
  const [mouseMoveParams, setMouseMoveParams] = useState<MouseMove>(VNCInteractionOptionsParamsState['mouseMoveParams'])
  const [mouseScrollParams, setMouseScrollParams] = useState<MouseScroll>(
    VNCInteractionOptionsParamsState['mouseScrollParams'],
  )

  const mouseClickNumberParamsFormData: (NumberParameterFormItem & { key: 'x' | 'y' })[] = [
    { label: 'Coord X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Coord Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
  ]

  const mouseDoubleClickFormData: ParameterFormItem & { key: 'double' } = {
    label: 'Double click',
    key: 'double',
    placeholder: 'Is mouse double click',
  }

  const mouseClickParamsFormData: ParameterFormData<MouseClick> = [
    ...mouseClickNumberParamsFormData,
    mouseButtonFormData,
    mouseDoubleClickFormData,
  ]

  const mouseDragNumberParamsFormData: (NumberParameterFormItem & { key: 'startX' | 'startY' | 'endX' | 'endY' })[] = [
    { label: 'Start X', key: 'startX', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Start Y', key: 'startY', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'End X', key: 'endX', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'End Y', key: 'endY', min: 0, max: Infinity, placeholder: '100', required: true },
  ]
  const mouseDragParamsFormData: ParameterFormData<MouseDrag> = [...mouseDragNumberParamsFormData, mouseButtonFormData]

  const mouseMoveNumberParamsFormData: (NumberParameterFormItem & { key: 'x' | 'y' })[] = [
    { label: 'Coord X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Coord Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
  ]
  const mouseMoveParamsFormData: ParameterFormData<MouseMove> = mouseMoveNumberParamsFormData

  const mouseScrollNumberParamsFormData: (NumberParameterFormItem & { key: 'x' | 'y' })[] = [
    { label: 'Coord X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Coord Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
  ]

  const mouseScrollDirectionFormData: ParameterFormItem & { key: 'direction' } = {
    label: 'Scroll direction',
    key: 'direction',
    placeholder: 'Mouse scroll direction',
  }

  const mouseScrollDirectionOptions = [
    {
      value: MouseScrollDirection.DOWN,
      label: 'Down',
    },
    {
      value: MouseScrollDirection.UP,
      label: 'Up',
    },
  ]

  const mouseScrollAmountFormData: NumberParameterFormItem & { key: 'amount' } = {
    label: 'Scroll amount',
    key: 'amount',
    placeholder: 'Mouse scroll amount',
    min: 1,
    max: Infinity,
  }

  const mouseScrollParamsFormData: ParameterFormData<MouseScroll> = [
    ...mouseScrollNumberParamsFormData,
    mouseScrollDirectionFormData,
    mouseScrollAmountFormData,
  ]

  const mouseActionsWithParamsFormData: MouseActionWithParamsFormData[] = [
    {
      methodName: MouseActions.CLICK,
      label: 'click()',
      description: 'Clicks the mouse at the specified coordinates',
      parametersFormItems: mouseClickParamsFormData,
      parametersState: mouseClickParams,
      onChangeParamsValidationDisabled: true,
    },
    {
      methodName: MouseActions.DRAG,
      label: 'drag()',
      description: 'Drags the mouse from start coordinates to end coordinates',
      parametersFormItems: mouseDragParamsFormData,
      parametersState: mouseDragParams,
      onChangeParamsValidationDisabled: true,
    },
    {
      methodName: MouseActions.MOVE,
      label: 'move()',
      description: 'Moves the mouse cursor to the specified coordinates',
      parametersFormItems: mouseMoveParamsFormData,
      parametersState: mouseMoveParams,
      onChangeParamsValidationDisabled: true,
    },
    {
      methodName: MouseActions.SCROLL,
      label: 'scroll()',
      description: 'Scrolls the mouse wheel at the specified coordinates',
      parametersFormItems: mouseScrollParamsFormData,
      parametersState: mouseScrollParams,
      onChangeParamsValidationDisabled: true,
    },
  ]

  const mouseActionsWithoutParamsFormData: PlaygroundActionFormDataBasic<MouseActions>[] = [
    {
      methodName: MouseActions.GET_POSITION,
      label: 'getPosition()',
      description: 'Gets the current mouse cursor position',
    },
  ]

  // Disable logic ensures that this method is called when ComputerUseClient exists -> we use as ComputerUse to silence TS compiler
  const mouseActionAPICall: PlaygroundActionInvokeApi = async (mouseActionFormData) => {
    const MouseActionsClient = (ComputerUseClient as ComputerUse).mouse
    let mouseActionResponseText = ''
    switch (mouseActionFormData.methodName) {
      case MouseActions.CLICK: {
        const mouseClickResponse = await MouseActionsClient[MouseActions.CLICK](
          mouseClickParams.x,
          mouseClickParams.y,
          mouseClickParams.button ?? undefined,
          mouseClickParams.double,
        )
        mouseActionResponseText = `Mouse clicked at (${mouseClickResponse.x}, ${mouseClickResponse.y})`
        break
      }
      case MouseActions.DRAG: {
        const mouseDragResponse = await MouseActionsClient[MouseActions.DRAG](
          mouseDragParams.startX,
          mouseDragParams.startY,
          mouseDragParams.endX,
          mouseDragParams.endY,
          mouseDragParams.button ?? undefined,
        )
        mouseActionResponseText = `Mouse drag ended at (${mouseDragResponse.x}, ${mouseDragResponse.y})`
        break
      }
      case MouseActions.MOVE: {
        const mouseMoveResponse = await MouseActionsClient[MouseActions.MOVE](mouseMoveParams.x, mouseMoveParams.y)
        mouseActionResponseText = `Mouse moved to (${mouseMoveResponse.x}, ${mouseMoveResponse.y})`
        break
      }
      case MouseActions.SCROLL: {
        const mouseScrollResponse = await MouseActionsClient[MouseActions.SCROLL](
          mouseScrollParams.x,
          mouseScrollParams.y,
          mouseScrollParams.direction,
          mouseScrollParams.amount ?? undefined,
        )
        mouseActionResponseText = mouseScrollResponse
          ? `Mouse scrolled ${mouseScrollParams.direction} at (${mouseScrollParams.x}, ${mouseScrollParams.y}) by ${mouseScrollParams.amount ?? 1}`
          : `Failed to scroll ${mouseScrollParams.direction} at (${mouseScrollParams.x}, ${mouseScrollParams.y})`
        break
      }
      case MouseActions.GET_POSITION: {
        const mousePositionResponse = await MouseActionsClient[MouseActions.GET_POSITION]()
        mouseActionResponseText = `Mouse is at (${mousePositionResponse.x}, ${mousePositionResponse.y})`
        break
      }
    }
    setVNCInteractionOptionsParamValue('responseContent', mouseActionResponseText)
  }

  return (
    <div className="flex flex-col gap-6">
      {mouseActionsWithParamsFormData.map((mouseActionFormData) => (
        <div key={mouseActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<MouseActions>
            actionFormItem={mouseActionFormData}
            onRunActionClick={() =>
              runPlaygroundActionWithParams(mouseActionFormData, wrapVNCInvokeApi(mouseActionAPICall))
            }
            disable={disableActions}
          />
          <div className="space-y-2">
            {mouseActionFormData.methodName === MouseActions.CLICK && (
              <>
                {mouseClickNumberParamsFormData.map((mouseClickNumberParamFormItem) => (
                  <InlineInputFormControl
                    key={mouseClickNumberParamFormItem.key}
                    formItem={mouseClickNumberParamFormItem}
                  >
                    <FormNumberInput
                      numberValue={mouseClickParams[mouseClickNumberParamFormItem.key]}
                      numberFormItem={mouseClickNumberParamFormItem}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          mouseActionFormData,
                          mouseClickNumberParamFormItem,
                          setMouseClickParams,
                          'mouseClickParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
                <MouseButtonSelect<MouseClick>
                  mouseActionFormData={mouseActionFormData}
                  paramsStateObject={mouseClickParams}
                  paramsStateSetter={setMouseClickParams}
                  contextParamsPropertyName="mouseClickParams"
                />
                <InlineInputFormControl formItem={mouseDoubleClickFormData}>
                  <FormCheckboxInput
                    checkedValue={mouseClickParams[mouseDoubleClickFormData.key as 'double']}
                    formItem={mouseDoubleClickFormData}
                    onChangeHandler={(checked) =>
                      playgroundActionParamValueSetter(
                        mouseActionFormData,
                        mouseDoubleClickFormData,
                        setMouseClickParams,
                        'mouseClickParams',
                        checked,
                      )
                    }
                  />
                </InlineInputFormControl>
              </>
            )}
            {mouseActionFormData.methodName === MouseActions.DRAG && (
              <>
                {mouseDragNumberParamsFormData.map((mouseDragNumberParamFormItem) => (
                  <InlineInputFormControl
                    key={mouseDragNumberParamFormItem.key}
                    formItem={mouseDragNumberParamFormItem}
                  >
                    <FormNumberInput
                      numberValue={mouseDragParams[mouseDragNumberParamFormItem.key]}
                      numberFormItem={mouseDragNumberParamFormItem}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          mouseActionFormData,
                          mouseDragNumberParamFormItem,
                          setMouseDragParams,
                          'mouseDragParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
                <MouseButtonSelect<MouseDrag>
                  mouseActionFormData={mouseActionFormData}
                  paramsStateObject={mouseDragParams}
                  paramsStateSetter={setMouseDragParams}
                  contextParamsPropertyName="mouseDragParams"
                />
              </>
            )}
            {mouseActionFormData.methodName === MouseActions.MOVE && (
              <>
                {mouseMoveNumberParamsFormData.map((mouseMoveNumberParamFormItem) => (
                  <InlineInputFormControl
                    key={mouseMoveNumberParamFormItem.key}
                    formItem={mouseMoveNumberParamFormItem}
                  >
                    <FormNumberInput
                      numberValue={mouseMoveParams[mouseMoveNumberParamFormItem.key]}
                      numberFormItem={mouseMoveNumberParamFormItem}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          mouseActionFormData,
                          mouseMoveNumberParamFormItem,
                          setMouseMoveParams,
                          'mouseMoveParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
              </>
            )}
            {mouseActionFormData.methodName === MouseActions.SCROLL && (
              <>
                {mouseScrollNumberParamsFormData.map((mouseScrollNumberParamFormItem) => (
                  <InlineInputFormControl
                    key={mouseScrollNumberParamFormItem.key}
                    formItem={mouseScrollNumberParamFormItem}
                  >
                    <FormNumberInput
                      numberValue={mouseScrollParams[mouseScrollNumberParamFormItem.key]}
                      numberFormItem={mouseScrollNumberParamFormItem}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          mouseActionFormData,
                          mouseScrollNumberParamFormItem,
                          setMouseScrollParams,
                          'mouseScrollParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
                <InlineInputFormControl formItem={mouseScrollDirectionFormData}>
                  <FormSelectInput
                    selectOptions={mouseScrollDirectionOptions}
                    selectValue={mouseScrollParams[mouseScrollDirectionFormData.key]}
                    formItem={mouseScrollDirectionFormData}
                    onChangeHandler={(value) =>
                      playgroundActionParamValueSetter(
                        mouseActionFormData,
                        mouseScrollDirectionFormData,
                        setMouseScrollParams,
                        'mouseScrollParams',
                        value,
                      )
                    }
                  />
                </InlineInputFormControl>
                <InlineInputFormControl formItem={mouseScrollAmountFormData}>
                  <FormNumberInput
                    numberValue={mouseScrollParams[mouseScrollAmountFormData.key]}
                    numberFormItem={mouseScrollAmountFormData}
                    onChangeHandler={(value) =>
                      playgroundActionParamValueSetter(
                        mouseActionFormData,
                        mouseScrollAmountFormData,
                        setMouseScrollParams,
                        'mouseScrollParams',
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
      {mouseActionsWithoutParamsFormData.map((mouseActionFormData) => (
        <div key={mouseActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<MouseActions>
            actionFormItem={mouseActionFormData}
            onRunActionClick={() =>
              runPlaygroundActionWithoutParams(mouseActionFormData, wrapVNCInvokeApi(mouseActionAPICall))
            }
            disable={disableActions}
          />
        </div>
      ))}
    </div>
  )
}

type MouseButtonSelectProps<T> = {
  mouseActionFormData: MouseActionWithParamsFormData
  paramsStateObject: T
  paramsStateSetter: React.Dispatch<React.SetStateAction<T>>
  contextParamsPropertyName: 'mouseClickParams' | 'mouseDragParams'
}

const MouseButtonSelect = <T extends MouseClick | MouseDrag>({
  mouseActionFormData,
  paramsStateObject,
  paramsStateSetter,
  contextParamsPropertyName,
}: MouseButtonSelectProps<T>) => {
  const { playgroundActionParamValueSetter } = usePlayground()

  const mouseButtonOptions = [
    {
      value: MouseButton.LEFT,
      label: 'Left',
    },
    {
      value: MouseButton.MIDDLE,
      label: 'Middle',
    },
    {
      value: MouseButton.RIGHT,
      label: 'Right',
    },
  ]

  return (
    <InlineInputFormControl formItem={mouseButtonFormData}>
      <FormSelectInput
        selectOptions={mouseButtonOptions}
        selectValue={paramsStateObject[mouseButtonFormData.key as 'button']}
        formItem={mouseButtonFormData}
        onChangeHandler={(value) =>
          playgroundActionParamValueSetter(
            mouseActionFormData,
            mouseButtonFormData,
            paramsStateSetter,
            contextParamsPropertyName,
            value,
          )
        }
      />
    </InlineInputFormControl>
  )
}

export default VNCMouseOperations
