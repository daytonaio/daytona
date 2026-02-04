/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Input } from '@/components/ui/input'
import { ParameterFormItem } from '@/enums/Playground'

type FormTextInputProps = {
  textValue: string | undefined
  formItem: ParameterFormItem
  onChangeHandler: (value: string) => void
}

const FormTextInput: React.FC<FormTextInputProps> = ({ textValue, formItem, onChangeHandler }) => {
  return (
    <Input
      id={formItem.key}
      className="w-full"
      placeholder={formItem.placeholder}
      value={textValue}
      onChange={(e) => onChangeHandler(e.target.value)}
    />
  )
}

export default FormTextInput
