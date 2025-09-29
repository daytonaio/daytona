/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { usePlayground } from '@/hooks/usePlayground'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormNumberInput from '../../Inputs/NumberInput'
import FormSelectInput from '../../Inputs/SelectInput'
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
} from '@/enums/Playground'
import { Loader2, Play } from 'lucide-react'
import React, { useState, MouseEvent } from 'react'
import FormCheckboxInput from '../../Inputs/CheckboxInput'

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

  const onMouseActionRunClick = <T extends MouseClick | MouseDrag | MouseMove | MouseScroll | object>(
    mouseActionMethodName: MouseActions,
    mouseActionParamsFormData: ParameterFormData<T>,
    mouseActionParamsState: T,
  ) => {
    setRunningMouseActionMethod(mouseActionMethodName)
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
          [mouseActionMethodName]: `${mouseActionEmptyParamFormItem?.label} parameter is required for this action`,
        })
        setRunningMouseActionMethod(null)
        return
      }
    }
    //TODO -> API call + set API response as responseText if present
    setMouseActionError({}) // Reset error
    setRunningMouseActionMethod(null)
  }

  type MouseActionRunButtonProps = {
    mouseActionMethodName: MouseActions
    onClickHandler: (event: MouseEvent<HTMLButtonElement>) => void
  }

  const MouseActionRunButton = ({ onClickHandler, mouseActionMethodName }: MouseActionRunButtonProps) => {
    return (
      <div>
        <Button disabled={!!runningMouseActionMethod} variant="outline" title="Run" onClick={onClickHandler}>
          {runningMouseActionMethod === mouseActionMethodName ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Play className="w-4 h-4" />
          )}
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={MouseActions.CLICK}>click()</Label>
            <p id={MouseActions.CLICK} className="text-sm text-muted-foreground mt-1 pl-1">
              Clicks the mouse at the specified coordinates
            </p>
          </div>
          <MouseActionRunButton
            mouseActionMethodName={MouseActions.CLICK}
            onClickHandler={() => onMouseActionRunClick(MouseActions.CLICK, mouseClickParamsFormData, mouseClickParams)}
          />
        </div>
        <div>
          {mouseActionError[MouseActions.CLICK] && (
            <p className="text-sm text-red-500 mt-2">{mouseActionError[MouseActions.CLICK]}</p>
          )}
        </div>
        <div className="px-4 space-y-2">
          {mouseClickNumberParamsFormData.map((mouseClickNumberParamFormItem) => (
            <InlineInputFormControl key={mouseClickNumberParamFormItem.key} formItem={mouseClickNumberParamFormItem}>
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
        </div>
      </div>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={MouseActions.DRAG}>drag()</Label>
            <p id={MouseActions.DRAG} className="text-sm text-muted-foreground mt-1 pl-1">
              Drags the mouse from start coordinates to end coordinates
            </p>
          </div>
          <MouseActionRunButton
            mouseActionMethodName={MouseActions.DRAG}
            onClickHandler={() => onMouseActionRunClick(MouseActions.DRAG, mouseDragParamsFormData, mouseDragParams)}
          />
        </div>
        <div>
          {mouseActionError[MouseActions.DRAG] && (
            <p className="text-sm text-red-500 mt-2">{mouseActionError[MouseActions.DRAG]}</p>
          )}
        </div>
        <div className="px-4 space-y-2">
          {mouseDragNumberParamsFormData.map((mouseDragNumberParamFormItem) => (
            <InlineInputFormControl key={mouseDragNumberParamFormItem.key} formItem={mouseDragNumberParamFormItem}>
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
        </div>
      </div>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={MouseActions.GET_POSITION}>getPosition()</Label>
            <p id={MouseActions.GET_POSITION} className="text-sm text-muted-foreground mt-1 pl-1">
              Gets the current mouse cursor position
            </p>
          </div>
          <MouseActionRunButton
            mouseActionMethodName={MouseActions.GET_POSITION}
            onClickHandler={() => onMouseActionRunClick(MouseActions.GET_POSITION, [], {})} // No parameters required for this action
          />
        </div>
      </div>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={MouseActions.MOVE}>move()</Label>
            <p id={MouseActions.MOVE} className="text-sm text-muted-foreground mt-1 pl-1">
              Moves the mouse cursor to the specified coordinates
            </p>
          </div>
          <MouseActionRunButton
            mouseActionMethodName={MouseActions.MOVE}
            onClickHandler={() => onMouseActionRunClick(MouseActions.MOVE, mouseMoveParamsFormData, mouseMoveParams)}
          />
        </div>
        <div>
          {mouseActionError[MouseActions.MOVE] && (
            <p className="text-sm text-red-500 mt-2">{mouseActionError[MouseActions.MOVE]}</p>
          )}
        </div>
        <div className="px-4 space-y-2">
          {mouseMoveNumberParamsFormData.map((mouseMoveNumberParamFormItem) => (
            <InlineInputFormControl key={mouseMoveNumberParamFormItem.key} formItem={mouseMoveNumberParamFormItem}>
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
        </div>
      </div>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={MouseActions.SCROLL}>scroll()</Label>
            <p id={MouseActions.SCROLL} className="text-sm text-muted-foreground mt-1 pl-1">
              Scrolls the mouse wheel at the specified coordinates
            </p>
          </div>
          <MouseActionRunButton
            mouseActionMethodName={MouseActions.SCROLL}
            onClickHandler={() =>
              onMouseActionRunClick(MouseActions.SCROLL, mouseScrollParamsFormData, mouseScrollParams)
            }
          />
        </div>
        <div>
          {mouseActionError[MouseActions.SCROLL] && (
            <p className="text-sm text-red-500 mt-2">{mouseActionError[MouseActions.SCROLL]}</p>
          )}
        </div>
        <div className="px-4 space-y-2">
          {mouseScrollNumberParamsFormData.map((mouseScrollNumberParamFormItem) => (
            <InlineInputFormControl key={mouseScrollNumberParamFormItem.key} formItem={mouseScrollNumberParamFormItem}>
              <FormNumberInput
                numberValue={mouseScrollParams[mouseScrollNumberParamFormItem.key]}
                numberFormItem={mouseScrollNumberParamFormItem}
                onChangeHandler={(value) => {
                  const mouseScrollParamsNew = { ...mouseScrollParams, [mouseScrollNumberParamFormItem.key]: value }
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
        </div>
      </div>
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
