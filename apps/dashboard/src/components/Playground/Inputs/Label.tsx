/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import { ParameterFormItem } from '@/enums/Playground'

type InputLabelProps = {
  formItem: ParameterFormItem
  isInline?: boolean
}

const InputLabel: React.FC<InputLabelProps> = ({ formItem, isInline }) => {
  return (
    <Label htmlFor={formItem.key} className="w-32 flex-shrink-0">
      <span className="relative">
        {`${formItem.label}${isInline ? '' : ''}`}
        {formItem.required ? <span className="text-muted-foreground">* </span> : null}
      </span>
    </Label>
  )
}

export default InputLabel
