/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { usePlayground } from '@/hooks/usePlayground'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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
} from '@/enums/Playground'
import { Loader2, Play } from 'lucide-react'
import React, { useState, MouseEvent } from 'react'

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
  const mouseClickParamsFormData: ParameterFormData<MouseClick> = [
    ...mouseClickNumberParamsFormData,
    { label: 'Button', key: 'button', placeholder: 'Mouse button' },
    { label: 'Double click', key: 'double', placeholder: '' },
  ]

  const mouseDragNumberParamsFormData: (NumberParameterFormItem & { key: 'startX' | 'startY' | 'endX' | 'endY' })[] = [
    { label: 'Start X', key: 'startX', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Start Y', key: 'startY', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'End X', key: 'endX', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'End Y', key: 'endY', min: 0, max: Infinity, placeholder: '100', required: true },
  ]
  const mouseDragParamsFormData: ParameterFormData<MouseDrag> = [
    ...mouseDragNumberParamsFormData,
    { label: 'Button', key: 'button', placeholder: 'Mouse button' },
  ]

  const mouseMoveNumberParamsFormData: (NumberParameterFormItem & { key: 'x' | 'y' })[] = [
    { label: 'Coord X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Coord Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
  ]
  const mouseMoveParamsFormData: ParameterFormData<MouseMove> = mouseMoveNumberParamsFormData

  const mouseScrollNumberParamsFormData: (NumberParameterFormItem & { key: 'x' | 'y' })[] = [
    { label: 'Coord X', key: 'x', min: 0, max: Infinity, placeholder: '100', required: true },
    { label: 'Coord Y', key: 'y', min: 0, max: Infinity, placeholder: '100', required: true },
  ]
  const mouseScrollParamsFormData: ParameterFormData<MouseScroll> = [
    ...mouseScrollNumberParamsFormData,
    { label: 'Scroll direction', key: 'direction', placeholder: 'Mouse scroll direction' },
    { label: 'Scroll amount', key: 'amount', placeholder: 'Mouse scroll amount' },
  ]

  const scrollDirectionOptions = [
    {
      value: MouseScrollDirection.DOWN,
      label: 'Down',
    },
    {
      value: MouseScrollDirection.UP,
      label: 'Up',
    },
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
          {mouseClickNumberParamsFormData.map((mouseClickNumberParam) => (
            <div key={mouseClickNumberParam.key} className="flex items-center gap-4">
              <Label htmlFor={mouseClickNumberParam.key} className="w-32 flex-shrink-0">
                <span>
                  {mouseClickNumberParam.required ? <span className="text-red-500">* </span> : null}
                  <span>{`${mouseClickNumberParam.label}:`}</span>
                </span>
              </Label>
              <Input
                id={mouseClickNumberParam.key}
                type="number"
                className="w-full"
                min={mouseClickNumberParam.min}
                max={mouseClickNumberParam.max}
                placeholder={mouseClickNumberParam.placeholder}
                step={mouseClickNumberParam.step}
                value={mouseClickParams[mouseClickNumberParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const mouseClickParamsNew = { ...mouseClickParams, [mouseClickNumberParam.key]: newValue }
                  setMouseClickParams(mouseClickParamsNew)
                  setVNCInteractionOptionsParamValue('mouseClickParams', mouseClickParamsNew)
                }}
              />
            </div>
          ))}
          <MouseButtonSelect<MouseClick>
            paramsStateObject={mouseClickParams}
            paramsStateSetter={setMouseClickParams}
            contextParamsPropertyName="mouseClickParams"
          />
          <div className="flex items-center gap-4">
            <Label htmlFor="double_click" className="w-32 flex-shrink-0">
              Double:
            </Label>
            <div className="flex-1 text-center">
              <Checkbox
                id="double_click"
                checked={mouseClickParams['double']}
                onCheckedChange={(value) => {
                  const mouseClickParamsNew = { ...mouseClickParams, double: !!value }
                  setMouseClickParams(mouseClickParamsNew)
                  setVNCInteractionOptionsParamValue('mouseClickParams', mouseClickParamsNew)
                }}
              />
            </div>
          </div>
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
          {mouseDragNumberParamsFormData.map((mouseDragNumberParam) => (
            <div key={mouseDragNumberParam.key} className="flex items-center gap-4">
              <Label htmlFor={mouseDragNumberParam.key} className="w-32 flex-shrink-0">
                <span>
                  {mouseDragNumberParam.required ? <span className="text-red-500">* </span> : null}
                  <span>{`${mouseDragNumberParam.label}:`}</span>
                </span>
              </Label>
              <Input
                id={mouseDragNumberParam.key}
                type="number"
                className="w-full"
                min={mouseDragNumberParam.min}
                max={mouseDragNumberParam.max}
                placeholder={mouseDragNumberParam.placeholder}
                step={mouseDragNumberParam.step}
                value={mouseDragParams[mouseDragNumberParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const mouseDragParamsNew = { ...mouseDragParams, [mouseDragNumberParam.key]: newValue }
                  setMouseDragParams(mouseDragParamsNew)
                  setVNCInteractionOptionsParamValue('mouseDragParams', mouseDragParamsNew)
                }}
              />
            </div>
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
          {mouseMoveNumberParamsFormData.map((mouseMoveNumberParam) => (
            <div key={mouseMoveNumberParam.key} className="flex items-center gap-4">
              <Label htmlFor={mouseMoveNumberParam.key} className="w-32 flex-shrink-0">
                <span>
                  {mouseMoveNumberParam.required ? <span className="text-red-500">* </span> : null}
                  <span>{`${mouseMoveNumberParam.label}:`}</span>
                </span>
              </Label>
              <Input
                id={mouseMoveNumberParam.key}
                type="number"
                className="w-full"
                min={mouseMoveNumberParam.min}
                max={mouseMoveNumberParam.max}
                placeholder={mouseMoveNumberParam.placeholder}
                step={mouseMoveNumberParam.step}
                value={mouseMoveParams[mouseMoveNumberParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const mouseMoveParamsNew = { ...mouseMoveParams, [mouseMoveNumberParam.key]: newValue }
                  setMouseMoveParams(mouseMoveParamsNew)
                  setVNCInteractionOptionsParamValue('mouseMoveParams', mouseMoveParamsNew)
                }}
              />
            </div>
          ))}
        </div>
        <div></div>
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
          {mouseScrollNumberParamsFormData.map((mouseScrollNumberParam) => (
            <div key={mouseScrollNumberParam.key} className="flex items-center gap-4">
              <Label htmlFor={mouseScrollNumberParam.key} className="w-32 flex-shrink-0">
                <span>
                  {mouseScrollNumberParam.required ? <span className="text-red-500">* </span> : null}
                  <span>{`${mouseScrollNumberParam.label}:`}</span>
                </span>
              </Label>
              <Input
                id={mouseScrollNumberParam.key}
                type="number"
                className="w-full"
                min={mouseScrollNumberParam.min}
                max={mouseScrollNumberParam.max}
                placeholder={mouseScrollNumberParam.placeholder}
                step={mouseScrollNumberParam.step}
                value={mouseScrollParams[mouseScrollNumberParam.key]}
                onChange={(e) => {
                  const newValue = e.target.value ? Number(e.target.value) : undefined
                  const mouseScrollParamsNew = { ...mouseScrollParams, [mouseScrollNumberParam.key]: newValue }
                  setMouseScrollParams(mouseScrollParamsNew)
                  setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
                }}
              />
            </div>
          ))}
          <div className="flex items-center gap-4">
            <Label htmlFor="scroll_direction" className="w-32 flex-shrink-0">
              <span>
                <span className="text-red-500">* </span>
                <span>Direction:</span>
              </span>
            </Label>
            <Select
              value={mouseScrollParams['direction']}
              onValueChange={(direction) => {
                const mouseScrollParamsNew = { ...mouseScrollParams, direction: direction as MouseScrollDirection }
                setMouseScrollParams(mouseScrollParamsNew)
                setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
              }}
            >
              <SelectTrigger className="w-full box-border rounded-lg" aria-label="Select scroll direction">
                <SelectValue id="scroll_direction" placeholder="Scroll direction" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {scrollDirectionOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center gap-4">
            <Label htmlFor="scroll_amount" className="w-32 flex-shrink-0">
              <span>Amount:</span>
            </Label>
            <Input
              id="scroll_amount"
              type="number"
              className="w-full"
              min={1}
              max={Infinity}
              placeholder="The amount to scroll"
              value={mouseScrollParams['amount']}
              onChange={(e) => {
                const newValue = e.target.value ? Number(e.target.value) : undefined
                const mouseScrollParamsNew = { ...mouseScrollParams, amount: newValue }
                setMouseScrollParams(mouseScrollParamsNew)
                setVNCInteractionOptionsParamValue('mouseScrollParams', mouseScrollParamsNew)
              }}
            />
          </div>
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
    <div className="flex items-center gap-4">
      <Label htmlFor="mouse_button" className="w-32 flex-shrink-0">
        Button:
      </Label>
      <Select
        value={paramsStateObject['button']}
        onValueChange={(button) => {
          const paramsStateObjectNew = { ...paramsStateObject, button: button as MouseButton }
          paramsStateSetter(paramsStateObjectNew)
          setVNCInteractionOptionsParamValue(contextParamsPropertyName, paramsStateObjectNew)
        }}
      >
        <SelectTrigger className="w-full box-border rounded-lg" aria-label="Select mouse button">
          <SelectValue id="mouse_button" placeholder="Mouse button" />
        </SelectTrigger>
        <SelectContent className="rounded-xl">
          {mouseButtonOptions.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}

export default VNCMouseOperations
