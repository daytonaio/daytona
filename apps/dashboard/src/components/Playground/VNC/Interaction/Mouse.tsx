/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { usePlayground } from '@/hooks/usePlayground'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import {
  MouseClick,
  MouseButton,
  MouseDrag,
  MouseMove,
  MouseScroll,
  ParameterFormData,
  MouseActions,
  NumberParameterFormItem,
  MouseScrollDirection,
  ParameterFormItem,
  PlaygroundActionFormDataBasic,
  MouseActionFormData,
} from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'
import React, { useState } from 'react'

const mouseButtonFormData: ParameterFormItem & { key: 'button' } = {
  label: 'Button',
  key: 'button',
  placeholder: 'Select mouse button',
}

const VNCMouseOperations: React.FC = () => {
  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue } = usePlayground()
  const [mouseClickParams, setMouseClickParams] = useState<MouseClick>(
    VNCInteractionOptionsParamsState['mouseClickParams'],
  )
  const [mouseDragParams, setMouseDragParams] = useState<MouseDrag>(VNCInteractionOptionsParamsState['mouseDragParams'])
  const [mouseMoveParams, setMouseMoveParams] = useState<MouseMove>(VNCInteractionOptionsParamsState['mouseMoveParams'])
  const [mouseScrollParams, setMouseScrollParams] = useState<MouseScroll>(
    VNCInteractionOptionsParamsState['mouseScrollParams'],
  )
  const [runningMouseActionMethod, setRunningMouseActionMethod] = useState<MouseActions | null>(null)
  const [mouseActionError, setMouseActionError] = useState<Partial<Record<MouseActions, string>>>({})

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

  const mouseActionsWithParamsFormData: MouseActionFormData<MouseClick | MouseDrag | MouseMove | MouseScroll>[] = [
    {
      methodName: MouseActions.CLICK,
      label: 'click()',
      description: 'Clicks the mouse at the specified coordinates',
      parametersFormItems: mouseClickParamsFormData,
      parametersState: mouseClickParams,
    },
    {
      methodName: MouseActions.DRAG,
      label: 'drag()',
      description: 'Drags the mouse from start coordinates to end coordinates',
      parametersFormItems: mouseDragParamsFormData,
      parametersState: mouseDragParams,
    },
    {
      methodName: MouseActions.MOVE,
      label: 'move()',
      description: 'Moves the mouse cursor to the specified coordinates',
      parametersFormItems: mouseMoveParamsFormData,
      parametersState: mouseMoveParams,
    },
    {
      methodName: MouseActions.SCROLL,
      label: 'scroll()',
      description: 'Scrolls the mouse wheel at the specified coordinates',
      parametersFormItems: mouseScrollParamsFormData,
      parametersState: mouseScrollParams,
    },
  ]

  const mouseActionsWithoutParamsFormData: PlaygroundActionFormDataBasic<MouseActions>[] = [
    {
      methodName: MouseActions.GET_POSITION,
      label: 'getPosition()',
      description: 'Gets the current mouse cursor position',
    },
  ]

  const onMouseActionRunClick = async <T extends MouseClick | MouseDrag | MouseMove | MouseScroll>(
    mouseActionFormData: MouseActionFormData<T>,
    mouseActionParamsFormData: ParameterFormData<T>,
    mouseActionParamsState: T,
  ) => {
    setRunningMouseActionMethod(mouseActionFormData.methodName)
    // Validate if all required params are set if they exist
    if (mouseActionParamsFormData.some((formItem) => formItem.required)) {
      const mouseActionEmptyParamFormItem = mouseActionParamsFormData
        .filter((formItem) => formItem.required)
        .find((formItem) => {
          const value = mouseActionParamsState[formItem.key]
          return value === '' || value === undefined
        })
      if (mouseActionEmptyParamFormItem) {
        setMouseActionError({
          [mouseActionFormData.methodName]: `${mouseActionEmptyParamFormItem?.label} parameter is required for this action`,
        })
        setRunningMouseActionMethod(null)
        return
      }
    }
    //TODO -> API call + set API response as responseText if present
    setMouseActionError({}) // Reset error
    setRunningMouseActionMethod(null)
  }

  return (
    <div className="space-y-6">
      {mouseActionsWithParamsFormData.map((mouseActionFormData) => (
        <div key={mouseActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<MouseActions>
            actionFormItem={mouseActionFormData}
            onRunActionClick={() =>
              onMouseActionRunClick<typeof mouseActionFormData.parametersState>(
                mouseActionFormData,
                mouseActionFormData.parametersFormItems,
                mouseActionFormData.parametersState,
              )
            }
            runningActionMethodName={runningMouseActionMethod}
            actionError={mouseActionError[mouseActionFormData.methodName]}
          />
          <div className="px-4 space-y-2">
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
                      onChangeHandler={(value) => {
                        const mouseClickParamsNew = { ...mouseClickParams, [mouseClickNumberParamFormItem.key]: value }
                        setMouseClickParams(mouseClickParamsNew)
                        setVNCInteractionOptionsParamValue('mouseClickParams', mouseClickParamsNew)
                      }}
                    />
                  </InlineInputFormControl>
                ))}
                <MouseButtonSelect<MouseClick>
                  paramsStateObject={mouseClickParams}
                  paramsStateSetter={setMouseClickParams}
                  contextParamsPropertyName="mouseClickParams"
                />
                <InlineInputFormControl formItem={mouseDoubleClickFormData}>
                  <FormCheckboxInput
                    checkedValue={mouseClickParams[mouseDoubleClickFormData.key as 'double']}
                    formItem={mouseDoubleClickFormData}
                    onChangeHandler={(checked) => {
                      const mouseClickParamsNew = { ...mouseClickParams, [mouseDoubleClickFormData.key]: checked }
                      setMouseClickParams(mouseClickParamsNew)
                      setVNCInteractionOptionsParamValue('mouseClickParams', mouseClickParamsNew)
                    }}
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
                      onChangeHandler={(value) => {
                        const mouseDragParamsNew = { ...mouseDragParams, [mouseDragNumberParamFormItem.key]: value }
                        setMouseDragParams(mouseDragParamsNew)
                        setVNCInteractionOptionsParamValue('mouseDragParams', mouseDragParamsNew)
                      }}
                    />
                  </InlineInputFormControl>
                ))}
                <MouseButtonSelect<MouseDrag>
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
                      onChangeHandler={(value) => {
                        const mouseMoveParamsNew = { ...mouseMoveParams, [mouseMoveNumberParamFormItem.key]: value }
                        setMouseMoveParams(mouseMoveParamsNew)
                        setVNCInteractionOptionsParamValue('mouseMoveParams', mouseMoveParamsNew)
                      }}
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
                      onChangeHandler={(value) => {
                        const mouseScrollParamsNew = {
                          ...mouseScrollParams,
                          [mouseScrollNumberParamFormItem.key]: value,
                        }
                        setMouseScrollParams(mouseScrollParamsNew)
                        setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
                      }}
                    />
                  </InlineInputFormControl>
                ))}
                <InlineInputFormControl formItem={mouseScrollDirectionFormData}>
                  <FormSelectInput
                    selectOptions={mouseScrollDirectionOptions}
                    selectValue={mouseScrollParams[mouseScrollDirectionFormData.key]}
                    formItem={mouseScrollDirectionFormData}
                    onChangeHandler={(value) => {
                      const mouseScrollParamsNew = {
                        ...mouseScrollParams,
                        [mouseScrollDirectionFormData.key]: value as MouseScrollDirection,
                      }
                      setMouseScrollParams(mouseScrollParamsNew)
                      setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
                    }}
                  />
                </InlineInputFormControl>
                <InlineInputFormControl formItem={mouseScrollAmountFormData}>
                  <FormNumberInput
                    numberValue={mouseScrollParams[mouseScrollAmountFormData.key]}
                    numberFormItem={mouseScrollAmountFormData}
                    onChangeHandler={(value) => {
                      const mouseScrollParamsNew = { ...mouseScrollParams, [mouseScrollAmountFormData.key]: value }
                      setMouseScrollParams(mouseScrollParamsNew)
                      setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
                    }}
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
            onRunActionClick={async () => {
              return
            }}
            runningActionMethodName={runningMouseActionMethod}
            actionError={mouseActionError[mouseActionFormData.methodName]}
          />
        </div>
      ))}
    </div>
  )
}

type MouseButtonSelectProps<T> = {
  paramsStateObject: T
  paramsStateSetter: React.Dispatch<React.SetStateAction<T>>
  contextParamsPropertyName: 'mouseClickParams' | 'mouseDragParams'
}

const MouseButtonSelect = <T extends MouseClick | MouseDrag>({
  paramsStateObject,
  paramsStateSetter,
  contextParamsPropertyName,
}: MouseButtonSelectProps<T>) => {
  const { setVNCInteractionOptionsParamValue } = usePlayground()

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
        onChangeHandler={(value) => {
          const paramsStateObjectNew = { ...paramsStateObject, [mouseButtonFormData.key]: value as MouseButton }
          paramsStateSetter(paramsStateObjectNew)
          setVNCInteractionOptionsParamValue(contextParamsPropertyName, paramsStateObjectNew)
        }}
      />
    </InlineInputFormControl>
  )
}

export default VNCMouseOperations
