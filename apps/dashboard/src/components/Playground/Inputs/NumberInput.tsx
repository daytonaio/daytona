/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Input } from '@/components/ui/input'
import { NumberParameterFormItem } from '@/contexts/PlaygroundContext'
import React from 'react'

type FormNumberInputProps = {
  numberValue: number | undefined
  numberFormItem: NumberParameterFormItem
  onChangeHandler: (value: number | undefined) => void
  disabled?: boolean
}

const FormNumberInput: React.FC<FormNumberInputProps> = ({
  numberValue,
  numberFormItem,
  onChangeHandler,
  disabled,
}) => {
  return (
    <Input
      id={numberFormItem.key}
      type="number"
      className="w-full"
      min={numberFormItem.min}
      max={numberFormItem.max}
      placeholder={numberFormItem.placeholder}
      step={numberFormItem.step}
      value={numberValue ?? ''}
      onChange={(e) => {
        const newValue = e.target.value ? Number(e.target.value) : undefined
        onChangeHandler(newValue)
      }}
      disabled={disabled}
    />
  )
}

export default FormNumberInput
